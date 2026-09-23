#!/usr/bin/env python3
"""Stable synthetic jury fixtures through the real authenticated HTTP workflow."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import time
import urllib.error
import urllib.parse
import urllib.request

ROOT = Path(__file__).resolve().parents[1]
ACTORS = {"producer": "producer:producer", "logistics": "logistics:carrier", "retailer": "retailer:retailer", "auditor": "regulator:auditor"}
SCENARIOS = [("NORMAL", 2800, "NO_SIGNAL", 4000), ("SUSPICIOUS", 3700, "REVIEW_REQUIRED", 8500)]


def fixture(scenario, price):
    suffix = "DEMO" + scenario + "01"
    batch, lot = "BAT-" + suffix, "LOT-" + suffix
    pickup, delivery = "TRF-" + suffix + "P", "TRF-" + suffix + "D"
    steps = [
        ("producer", "CreateBatch", "/batches", dict(productCode="TOMATO", gradeCode="STANDARD", quantityGrams=100000, originRegionCode="07", harvestDate="2026-09-15", logisticsMsp="LogisticsMSP", intendedRetailerMsp="RetailerMSP")),
        ("producer", "OfferPickup", "pickup-offers", dict(transferId=pickup, quantityGrams=100000)),
        ("logistics", "AcceptPickup", "pickup-acceptances", dict(transferId=pickup, quantityGrams=100000)),
        ("logistics", "RecordFreightCost", "freight-costs", dict(costId="CST-" + suffix)),
        ("logistics", "OfferDelivery", "delivery-offers", dict(transferId=delivery, quantityGrams=100000)),
        ("retailer", "AcceptDelivery", "delivery-acceptances", dict(transferId=delivery, quantityGrams=100000)),
        ("retailer", "ReportRetailPrice", "retail-reports", dict(reportId="RPT-" + suffix, retailLotId=lot, policyId="CFG-PRICE001")),
    ]
    commands = []
    for version, (actor, fn, route, payload) in enumerate(steps):
        operation = "OP-" + hashlib.sha256(("agrochain-demo-v1:" + scenario + ":" + fn).encode()).hexdigest()[:24].upper()
        request = {"scenario": scenario, "command": dict(schemaVersion="agrochain.command.v1", command=fn, operationId=operation, batchId=batch, expectedVersion=version, payload=payload), "privateInput": {}}
        if version == 6:
            request["privateInput"] = dict(offeredPriceKurusPerKg=price, currency="TRY", taxBasis="EXCLUDING_TAX", reportedAt="2026-09-22T09:00:00.000Z")
        commands.append((actor, route if version == 0 else f"/batches/{batch}/{route}", request))
    return batch, lot, commands


class Client:
    def __init__(self, base, token_file):
        parsed = urllib.parse.urlsplit(base)
        # Tokens may only travel to this machine; reject redirects and proxies too.
        if parsed.scheme != "http" or parsed.hostname not in ("localhost", "127.0.0.1") or parsed.username or parsed.password or parsed.path not in ("", "/") or parsed.query or parsed.fragment:
            raise ValueError("Demo URL must be a loopback http origin")
        self.base = base.rstrip("/") + "/api/v1"
        self.tokens = json.loads(Path(token_file).read_text())
        class NoRedirect(urllib.request.HTTPRedirectHandler):
            def redirect_request(self, *args, **kwargs):
                return None
        self.http = urllib.request.build_opener(urllib.request.ProxyHandler({}), NoRedirect())

    def call(self, method, path, actor=None, body=None):
        headers = {"Content-Type": "application/json"}
        if actor:
            headers["Authorization"] = "Bearer " + self.tokens[ACTORS[actor]]
        if body:
            headers["Idempotency-Key"] = body["command"]["operationId"]
        req = urllib.request.Request(self.base + path, data=json.dumps(body).encode() if body else None, headers=headers, method=method)
        try:
            response = self.http.open(req, timeout=45)
        except urllib.error.HTTPError as error:
            response = error
        with response:
            raw = response.read()
            data = json.loads(raw) if "application/json" in response.headers.get("Content-Type", "") else raw
            return response.status, data

    def submit(self, actor, path, request):
        # Retry the exact same request/operation after timeout or 202; never mint a new ID.
        deadline = time.monotonic() + 150
        while time.monotonic() < deadline:
            try:
                status, data = self.call("POST", path, actor, request)
                if status in (200, 201) and data.get("status") == "COMMITTED":
                    return data["txId"]
                if status not in (202, 503):
                    raise RuntimeError(f"{request['command']['command']}: HTTP {status}, {data.get('code', 'UNEXPECTED_RESPONSE')}")
            except (OSError, TimeoutError):
                pass
            time.sleep(1)
        raise RuntimeError("Commit not confirmed; preserve backend data and rerun the same seed command")


def verify(client):
    results = []
    for scenario, price, classification, increase in SCENARIOS:
        batch, lot, commands = fixture(scenario, price)
        for attempt in range(120):
            status, analysis = client.call("GET", f"/batches/{batch}/anomaly", "auditor")
            public_status, public = client.call("GET", f"/public/lots/{lot}")
            if status == 200 and analysis.get("classification") and public_status == 200:
                break
            if status not in (200, 404, 503) or public_status not in (200, 404, 503):
                raise RuntimeError("Demo query failed; check identities and network")
            time.sleep(.5)
        expected = {"classification": classification, "increaseBps": increase, "thresholdBps": 5000, "purchasePriceKurusPerKg": 2000, "retailPriceKurusPerKg": price}
        if status != 200 or any(analysis.get(k) != v for k, v in expected.items()):
            raise RuntimeError("Demo analysis differs from the fixed fixture or is pending")
        if public_status != 200 or public.get("batchId") != batch or len(public.get("history", [])) != 3:
            raise RuntimeError("Public provenance incomplete")
        for secret in ("Kurus", "classification", "anomaly", "saltHex", "explanation", "reviewState"):
            if secret in json.dumps(public):
                raise RuntimeError("Private field in consumer projection")
        for actor in ("producer", "logistics"):
            if client.call("GET", f"/batches/{batch}/anomaly", actor)[0] != 403:
                raise RuntimeError("Private analysis authorization failed")
        transaction_ids = []
        for actor, _, request in commands:
            code, operation = client.call("GET", "/operations/" + request["command"]["operationId"], actor)
            if code != 200 or operation.get("status") != "COMMITTED":
                raise RuntimeError("Demo operation journal is missing; restore the original backend store")
            transaction_ids.append(operation["txId"])
        # Deliberate allowlist: synthetic demo data only, never PDC salts or source bundles.
        results.append(dict(scenario=scenario, batchId=batch, lotId=lot, **expected, evaluationTxId=analysis["txId"], transactionIds=transaction_ids))
    return dict(schemaVersion="agrochain.demo-evidence.v1", sourceMode="SIMULATED", blockchainMode="FABRIC", syntheticDataOnly=True, scenarios=results)


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("mode", choices=("seed", "check"))
    parser.add_argument("--url", default="http://127.0.0.1:8080")
    parser.add_argument("--tokens", type=Path, default=ROOT / "network/runtime/backend/tokens.json")
    parser.add_argument("--evidence", type=Path, default=ROOT / "network/runtime/demo/evidence.json")
    args = parser.parse_args()
    os.umask(0o077)
    client = Client(args.url, args.tokens)
    if client.call("GET", "/health")[0] != 200:
        raise RuntimeError("Start the backend before running the demo")
    if args.mode == "seed":
        for scenario, price, _, _ in SCENARIOS:
            _, _, commands = fixture(scenario, price)
            for actor, path, request in commands:
                client.submit(actor, path, request)
            print("Committed synthetic demo:", scenario)
    result = verify(client)
    args.evidence.parent.mkdir(parents=True, exist_ok=True)
    args.evidence.write_text(json.dumps(result, indent=2) + "\n")
    print("PASS fixed demo, 40% / 85%, private access denied, consumer allowlist, 14 committed operations")


if __name__ == "__main__":
    main()
