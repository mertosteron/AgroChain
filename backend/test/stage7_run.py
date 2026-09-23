#!/usr/bin/env python3
"""One sequential gate with persisted evidence and disposable clean-network runs."""
import argparse
import datetime
import fcntl
import hashlib
import json
import os
from pathlib import Path
import platform
import secrets
import shutil
import subprocess
import sys
import tempfile
import time
from stage7_measure import summarize

ROOT = Path(__file__).resolve().parents[2]
OUT = ROOT / "network/runtime/stage7"
SERVICES = ("orderer1", "orderer2", "orderer3", "producer", "logistics", "retailer", "regulator")


def capture(args, cwd=ROOT, env=None):
    return subprocess.check_output(args, cwd=cwd, env=env, text=True, stderr=subprocess.STDOUT).strip()


def execute(label, args, report, cwd=ROOT, env=None, timeout=900):
    log = OUT / (label+".log")
    started = time.perf_counter()
    entry = {"name": label, "command": args, "result": "FAIL", "log": log.name}
    report["checks"].append(entry)
    print("RUN", label, flush=True)
    try:
        with log.open("w") as output:
            process = subprocess.run(args, cwd=cwd, env=env, stdout=output, stderr=subprocess.STDOUT, timeout=timeout)
        entry["exitCode"] = process.returncode
        if process.returncode:
            raise RuntimeError(f"{label} failed; inspect {log}")
        entry["result"] = "PASS"
    finally:
        entry["seconds"] = round(time.perf_counter()-started, 3)
        (OUT / "progress.json").write_text(json.dumps(report, indent=2)+"\n")
    print("PASS", label, entry["seconds"], "s", flush=True)


def environment():
    cpu = next((line.split(":", 1)[1].strip() for line in Path("/proc/cpuinfo").read_text().splitlines() if line.startswith("model name")), "unknown")
    mem = next(line.split(":", 1)[1].strip() for line in Path("/proc/meminfo").read_text().splitlines() if line.startswith("MemTotal:"))
    versions = {}
    for name, args in {"docker": ["docker", "version", "--format", "{{.Server.Version}}"],
                       "compose": ["docker", "compose", "version", "--short"],
                       "go": [str(ROOT / "network/tools/go-1.27.1/bin/go"), "version"],
                       "java": [str(ROOT / "network/tools/jdk-21.0.12.1+1/bin/java"), "-version"],
                       "fabric": [str(ROOT / "network/tools/bin/peer"), "version"]}.items():
        versions[name] = capture(args)
    return {"utc": datetime.datetime.now(datetime.timezone.utc).isoformat(), "os": platform.platform(),
            "cpu": cpu, "logicalCpus": os.cpu_count(), "memory": mem, "python": platform.python_version(),
            "versions": versions, "sourceCommit": capture(["git", "rev-parse", "HEAD"]),
            "workingTreeDirty": bool(capture(["git", "status", "--porcelain"])),
            "topology": "single host, 4 peers / MSPs, 3 Raft orderers, 4 CCAAS services, LevelDB, TLS, Retailer+Regulator endorsement",
            "concurrency": 1, "measurementState": "already running ledger, one Gateway sample and two HTTP routes warmup excluded; JIT/OS caches not flushed",
            "limits": "local pilot observations, no throughput/load/scalability/SLA claim"}


def clean_workspace(destination):
    # Copy current source, including uncommitted Stage 7 files; never clone secrets.
    names = capture(["git", "ls-files", "--cached", "--others", "--exclude-standard", "-z"]).split("\0")
    for name in names:
        src = ROOT / name
        if not name or not src.is_file():
            continue
        dst = destination / name
        dst.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(src, dst)
    (destination / "network/tools").symlink_to(ROOT / "network/tools", target_is_directory=True)
    for cache in ("go-mod", "go-build"):
        src = ROOT / "network/runtime" / cache
        if src.exists():
            shutil.copytree(src, destination / "network/runtime" / cache)
    target = destination / "backend/target"
    target.mkdir(parents=True)
    shutil.copy2(ROOT / "backend/target/agrochain-backend-0.3.0.jar", target)


def reproduce(report):
    # Port/container names are shared, so stop/recreate nodes sequentially. Original
    # identities and seven default ledger volumes are never reset or deleted.
    if os.environ.get("AGROCHAIN_LEDGER_VOLUME_PREFIX", "agrochain") != "agrochain":
        raise RuntimeError("Run reproduction from the default workspace, not an isolated child")
    expected = {"agrochain-"+s for s in SERVICES}
    running = set(capture(["docker", "ps", "--filter", "label=com.docker.compose.project=agrochain", "--format", "{{.Names}}"]).splitlines())
    if running != expected:
        raise RuntimeError("Start the complete existing pilot before the clean reproduction gate")
    original_volumes = json.loads(capture(["docker", "volume", "inspect", *["agrochain-"+s+"-data" for s in SERVICES]]))
    volume_ids = {v["Name"]: v["CreatedAt"] for v in original_volumes}
    manifest_path = ROOT / "network/channel-artifacts/generated-manifest.json"
    original_manifest = manifest_path.read_bytes()
    # Prevent an abrupt TERM/INT from skipping finally restoration.
    import signal
    def interrupted(signum, _frame):
        raise RuntimeError(f"Interrupted by signal {signum}")
    previous = {s: signal.signal(s, interrupted) for s in (signal.SIGTERM, signal.SIGINT)}
    report["cleanRuns"] = []
    try:
        execute("original-stop-preserve", ["make", "network-down"], report)
        for number in (1, 2):
            prefix = "agrochain-stage7-"+secrets.token_hex(6)
            workspace = Path(tempfile.mkdtemp(prefix="agrochain-stage7-"))
            clean_workspace(workspace)
            env = dict(os.environ, AGROCHAIN_LEDGER_VOLUME_PREFIX=prefix, CHAINCODE_SEQUENCE="1")
            # Assert all seven named test volumes are absent before generation.
            for service in SERVICES:
                probe = subprocess.run(["docker", "volume", "inspect", f"{prefix}-{service}-data"], capture_output=True)
                if probe.returncode == 0:
                    raise RuntimeError("Refusing existing test volume")
            try:
                execute(f"clean-{number}-bootstrap", ["make", "bootstrap", "chaincode-deploy", "backend-prepare"], report, workspace, env)
                # Includes real-member/nonmember PDC checks and three demo scenarios.
                execute(f"clean-{number}-privacy", ["make", "privacy-integration"], report, workspace, env)
                execute(f"clean-{number}-http", ["make", "stage6-integration"], report, workspace, env)
                scenarios = json.loads((workspace / "network/runtime/backend-acceptance/stage6-summary.json").read_text())
                outcomes = [{"classification": s["anomaly"]["classification"], "increaseBps": s["anomaly"]["increaseBps"],
                             "thresholdBps": s["anomaly"]["thresholdBps"]} for s in scenarios]
                if outcomes != [{"classification": "NO_SIGNAL", "increaseBps": 4000, "thresholdBps": 5000},
                                {"classification": "REVIEW_REQUIRED", "increaseBps": 8500, "thresholdBps": 5000}]:
                    raise AssertionError("Clean scenarios did not reproduce expected outcomes")
                genesis = workspace / "network/channel-artifacts/agrochannel.block"
                report["cleanRuns"].append({"run": number, "result": "PASS", "emptyLedgerVolumes": True,
                    "freshIdentitiesAndBackendStore": True, "genesisSha256": hashlib.sha256(genesis.read_bytes()).hexdigest(),
                    "outcomes": outcomes, "tamperReplayPdcRestart": "PASS"})
            finally:
                # Both commands use the fresh workspace and its manifest-bound prefix.
                execute(f"clean-{number}-stop", ["make", "network-down"], report, workspace, env)
                execute(f"clean-{number}-reset", ["make", "clean-generated"], report, workspace, env)
                shutil.rmtree(workspace)
        if report["cleanRuns"][0]["genesisSha256"] == report["cleanRuns"][1]["genesisSha256"]:
            raise AssertionError("Expected independently generated clean networks")
    finally:
        try:
            execute("original-restore", ["make", "network-up", "chaincode-check"], report)
            restored = json.loads(capture(["docker", "volume", "inspect", *volume_ids]))
            if {v["Name"]: v["CreatedAt"] for v in restored} != volume_ids or manifest_path.read_bytes() != original_manifest:
                raise AssertionError("Original volumes or identity manifest changed")
            report["originalRestored"] = True
        finally:
            for signum, handler in previous.items():
                signal.signal(signum, handler)


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("mode", choices=("check", "measure", "reproduce"))
    args = parser.parse_args()
    os.umask(0o077)
    OUT.mkdir(parents=True, exist_ok=True)
    with (OUT / "run.lock").open("w") as lock:
        fcntl.flock(lock, fcntl.LOCK_EX | fcntl.LOCK_NB)
        report = {"schemaVersion": "agrochain.stage7.v1", "result": "FAIL", "mode": args.mode, "checks": []}
        try:
            report["environment"] = environment()
            if args.mode == "check":
                execute("unit-and-build", ["make", "test", "chaincode-test", "backend-build", "privacy-vectors", "stage7-unit"], report)
                execute("network-ready", ["make", "network-up", "chaincode-check"], report)
                execute("live-security", ["make", "verify", "chaincode-check", "chaincode-integration", "privacy-integration"], report)
                execute("http-recovery", ["make", "stage6-integration", "backend-gateway-test"], report)
                execute("browser", ["make", "stage6-ui-test"], report)
            if args.mode in ("check", "measure"):
                execute("http-measurement", [sys.executable, "backend/test/stage7_measure.py"], report)
                env = dict(os.environ, AGROCHAIN_ROOT=str(ROOT), AGROCHAIN_MEASURE="true")
                execute("gateway-measurement", ["bash", "backend/mvnw", "-Dtest=GatewayMeasurementTest", "test"], report, env=env)
                gateway = json.loads((OUT / "gateway-measurements.json").read_text())
                report["gateway"] = {key: summarize([s[key] for s in gateway["samples"]]) for key in ("submitValidMs", "sharedQueryMs", "privateQueryMs")}
                http = json.loads((OUT / "http-measurements.json").read_text())
                report["http"] = http["summary"]
            if args.mode in ("check", "reproduce"):
                reproduce(report)
            report["result"] = "PASS"
        finally:
            (OUT / (args.mode+"-report.json")).write_text(json.dumps(report, indent=2)+"\n")
        print("PASS Stage 7", args.mode, "report:", OUT / (args.mode+"-report.json"))


if __name__ == "__main__":
    main()
