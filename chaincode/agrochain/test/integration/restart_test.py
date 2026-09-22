#!/usr/bin/env python3
"""Verify existing ledgers and definition across non-destructive full shutdown."""
import json
import subprocess
from fabric_test import ROOT, ORGS, call, query, BATCH


def snapshot():
    result = {}
    for org in ORGS:
        info = call(org, "info")
        assert info.returncode == 0, info.stderr
        ledger = json.loads(info.stdout.split("Blockchain info: ", 1)[1])
        result[org] = {"ledger": ledger, "health": query(org, "Health"),
                       "batchExists": query(org, "BatchExists", BATCH)}
    return result


before = snapshot()
for target in ("network-down", "network-up", "chaincode-check"):
    subprocess.run(["make", target], cwd=ROOT, check=True)
after = snapshot()
assert before == after, "Ledger tip, contract definition/health or query state changed on restart"
(ROOT / "network/runtime/chaincode/restart-evidence.json").write_text(
    json.dumps({"before": before, "after": after, "preserved": True}, indent=2) + "\n")
print("PASS full network shutdown/start: all four ledger heights/hashes, contract health and query state preserved")
