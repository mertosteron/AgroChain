#!/usr/bin/env python3
"""Rehearse the Stage 8 guide on fresh identities/disks, then restore the pilot.

Reuses the already safety-tested Stage 7 isolated-volume and source-copy mechanism.
Build caches are reused; credentials, ledgers and backend journals are never copied.
"""
import fcntl
import hashlib
import json
import os
from pathlib import Path
import secrets
import shutil
import signal
import subprocess
import sys
import tempfile
import time
import urllib.request

ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(ROOT / "backend/test"))
import stage7_run as runner


def main():
    os.umask(0o077)
    if os.environ.get("AGROCHAIN_LEDGER_VOLUME_PREFIX", "agrochain") != "agrochain":
        raise RuntimeError("Run rehearsal from the default pilot workspace")
    runner.OUT = ROOT / "network/runtime/stage8-rehearsal"
    runner.OUT.mkdir(parents=True, exist_ok=True)
    with (ROOT / "network/runtime/stage7/run.lock").open("w") as lock:
        fcntl.flock(lock, fcntl.LOCK_EX | fcntl.LOCK_NB)
        report = {"schemaVersion": "agrochain.stage8-rehearsal.v1", "result": "FAIL", "checks": [], "scope": "same Linux host, fresh identities/disks, cached tools/dependencies"}
        manifest = ROOT / "network/channel-artifacts/generated-manifest.json"
        before = manifest.read_bytes()
        volumes = json.loads(runner.capture(["docker", "volume", "inspect", *[f"agrochain-{s}-data" for s in runner.SERVICES]]))
        original = {v["Name"]: v["CreatedAt"] for v in volumes}
        workspace = Path(tempfile.mkdtemp(prefix="agrochain-stage8-"))
        prefix = "agrochain-stage7-" + secrets.token_hex(6)
        env = dict(os.environ, AGROCHAIN_LEDGER_VOLUME_PREFIX=prefix, CHAINCODE_SEQUENCE="1")
        process = None
        def interrupted(signum, frame):
            raise RuntimeError("Rehearsal interrupted; restoring original network")
        handlers = {s: signal.signal(s, interrupted) for s in (signal.SIGINT, signal.SIGTERM)}
        try:
            runner.clean_workspace(workspace)
            cache = ROOT / "network/runtime/maven"
            if cache.exists():
                shutil.copytree(cache, workspace / "network/runtime/maven")
            for service in runner.SERVICES:
                if subprocess.run(["docker", "volume", "inspect", f"{prefix}-{service}-data"], capture_output=True).returncode == 0:
                    raise RuntimeError("Refusing existing test volume")
            runner.execute("original-stop-preserve", ["make", "network-down"], report)
            runner.execute("documented-fresh-setup", ["make", "demo-setup"], report, workspace, env)
            backend_env = dict(env, AGROCHAIN_ROOT=str(workspace), AGROCHAIN_API_PORT="18082")
            with (runner.OUT / "backend.log").open("w") as output:
                process = subprocess.Popen(["make", "backend-run"], cwd=workspace, env=backend_env, stdout=output, stderr=subprocess.STDOUT, start_new_session=True)
                for _ in range(100):
                    if process.poll() is not None:
                        raise RuntimeError("Fresh backend exited")
                    try:
                        with urllib.request.urlopen("http://127.0.0.1:18082/api/v1/health", timeout=2) as response:
                            if response.status == 200:
                                break
                    except OSError:
                        pass
                    time.sleep(.3)
                else:
                    raise RuntimeError("Fresh backend did not start")
                cmd = [sys.executable, "scripts/demo.py", "seed", "--url", "http://127.0.0.1:18082"]
                runner.execute("first-seed", cmd, report, workspace, env)
                first = json.loads((workspace / "network/runtime/demo/evidence.json").read_text())
                runner.execute("idempotent-second-seed", cmd, report, workspace, env)
                second = json.loads((workspace / "network/runtime/demo/evidence.json").read_text())
                if first != second:
                    raise AssertionError("Second seed changed transaction identities or results")
                runner.execute("documented-demo-check", [sys.executable, "scripts/demo.py", "check", "--url", "http://127.0.0.1:18082"], report, workspace, env)
                report["fixedFixtures"] = first
                report["sameTransactionsOnReplay"] = True
                report["freshGenesisSha256"] = hashlib.sha256((workspace / "network/channel-artifacts/agrochannel.block").read_bytes()).hexdigest()
            report["result"] = "PASS"
        finally:
            try:
                if process and process.poll() is None:
                    os.killpg(process.pid, signal.SIGTERM)
                    try:
                        process.wait(timeout=20)
                    except subprocess.TimeoutExpired:
                        os.killpg(process.pid, signal.SIGKILL)
                        process.wait(timeout=5)
                if (workspace / "network/channel-artifacts/generated-manifest.json").exists():
                    runner.execute("temporary-stop", ["make", "network-down"], report, workspace, env)
                    runner.execute("temporary-clean-only", ["make", "clean-generated"], report, workspace, env)
                shutil.rmtree(workspace)
            except Exception:
                report["result"] = "FAIL"
                raise
            finally:
                try:
                    runner.execute("original-restore", ["make", "network-up", "chaincode-check"], report)
                    restored = json.loads(runner.capture(["docker", "volume", "inspect", *original]))
                    if {v["Name"]: v["CreatedAt"] for v in restored} != original or manifest.read_bytes() != before:
                        report["result"] = "FAIL"
                        raise AssertionError("Original data or identity manifest changed")
                    report["originalRestored"] = True
                except Exception:
                    report["result"] = "FAIL"
                    raise
                finally:
                    for signum, handler in handlers.items():
                        signal.signal(signum, handler)
                    (runner.OUT / "report.json").write_text(json.dumps(report, indent=2) + "\n")
        print("PASS documented Stage 8 setup, stable replay and original data restoration")


if __name__ == "__main__":
    main()
