#!/usr/bin/env python3
"""Sequential HTTP pilot observations; no throughput or SLA inference."""
import json
import math
import os
from pathlib import Path
import statistics
import time


def summarize(samples):
    if not samples or any(not math.isfinite(x) or x < 0 for x in samples):
        raise ValueError("Expected nonempty finite nonnegative durations")
    ordered = sorted(samples)
    return {"n": len(samples), "medianMs": round(statistics.median(samples), 3),
            "p95Ms": round(ordered[math.ceil(.95 * len(ordered))-1], 3),
            "minMs": round(ordered[0], 3), "maxMs": round(ordered[-1], 3)}


def main():
    import integration_test as base
    from stage6_test import Stage6Integration
    os.umask(0o077)
    result = {"schemaVersion": "agrochain.measurements.v1", "result": "FAIL", "concurrency": 1,
              "clock": "perf_counter_ns", "p95Method": "nearest-rank ceil(0.95*n)",
              "scope": "HTTP client through VALID commit; full route includes analysis and public projection polling",
              "warmup": "one normal and one suspicious full route excluded", "failures": 0,
              "commands": [], "routes": []}
    base.DATA = base.ROOT / "network/runtime/stage7/measurement-backend"
    output = base.ROOT / "network/runtime/stage7/http-measurements.json"
    output.parent.mkdir(parents=True, exist_ok=True)

    class Measurement(Stage6Integration):
        recording = False

        def submit(self, actor, path, request, expected=201):
            start = time.perf_counter_ns()
            response = super().submit(actor, path, request, expected)
            elapsed = (time.perf_counter_ns()-start)/1e6
            if self.recording:
                result["commands"].append({"command": request["command"]["command"], "ms": elapsed,
                    "requestBytes": len(json.dumps(request, separators=(",", ":")).encode()),
                    "status": response["status"], "txId": response["txId"]})
            return response

        def measured_route(self, name, price):
            start = time.perf_counter_ns()
            s = self.run_scenario(name, price)
            deadline = time.monotonic()+45
            while True:
                status, a = base.call("GET", f'/batches/{s["batch"]}/anomaly', "regulator")
                public_status, _ = base.call("GET", f'/public/lots/{s["lot"]}')
                if status == 200 and a.get("classification") and public_status == 200:
                    break
                if time.monotonic() >= deadline:
                    raise AssertionError("Analysis or consumer projection did not converge")
                time.sleep(.1)
            self.assertEqual(a["classification"], "NO_SIGNAL" if price == 2800 else "REVIEW_REQUIRED")
            self.assertEqual(a["increaseBps"], 4000 if price == 2800 else 8500)
            if self.recording:
                result["routes"].append({"scenario": name, "ms": (time.perf_counter_ns()-start)/1e6,
                    "classification": a["classification"], "increaseBps": a["increaseBps"],
                    "batchId": s["batch"], "analysisTxId": a["txId"]})
            return s

    case = Measurement()
    try:
        Measurement.start()
        for _ in range(4):
            case.scenarios = [case.measured_route("NORMAL", 2800), case.measured_route("SUSPICIOUS", 3700)]
            case.recording = True
        # Third demo scenario: original bytes pass, modified bytes fail.
        case.recording = False
        case.test_05_original_modified_document_and_archive_authorization()
        case.test_03_retry_conflict_and_caller_scoped_status()
        result["tamperAndReplay"] = "PASS"
        result["summary"] = {"httpCommand": summarize([x["ms"] for x in result["commands"]]),
                             "fullRoute": summarize([x["ms"] for x in result["routes"]])}
        result["perCommand"] = {name: summarize([x["ms"] for x in result["commands"] if x["command"] == name])
                                for name in sorted({x["command"] for x in result["commands"]})}
        result["result"] = "PASS"
    except Exception:
        result["failures"] += 1
        raise
    finally:
        Measurement.stop()
        output.write_text(json.dumps(result, indent=2)+"\n")
    print("PASS HTTP observations:", json.dumps(result["summary"]))


if __name__ == "__main__":
    main()
