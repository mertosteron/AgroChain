#!/usr/bin/env python3
"""Real HTTP -> Java Fabric Gateway -> signed simulator -> PDC acceptance."""
import base64
import copy
import json
import os
from pathlib import Path
import secrets
import sqlite3
import subprocess
import time
import unittest
import urllib.error
import urllib.request

ROOT = Path(__file__).resolve().parents[2]
DATA = ROOT / "network/runtime/backend-acceptance"
PORT = int(os.environ.get("AGROCHAIN_TEST_API_PORT", "18080"))
TOKENS = json.loads((ROOT / "network/runtime/backend/tokens.json").read_text())
ACTORS = {"producer": "producer:producer", "logistics": "logistics:carrier", "retailer": "retailer:retailer", "regulator": "regulator:auditor", "public-reader": "regulator:public-reader"}


def ident(prefix):
    return prefix + "-" + secrets.token_hex(16).upper()


def call(method, path, actor=None, body=None, headers=None):
    h = {"Content-Type": "application/json"}
    if actor:
        h["Authorization"] = "Bearer " + TOKENS[ACTORS[actor]]
    h.update(headers or {})
    data = None if body is None else json.dumps(body, separators=(",", ":")).encode()
    request = urllib.request.Request(f"http://127.0.0.1:{PORT}/api/v1" + path, data=data, headers=h, method=method)
    try:
        response = urllib.request.urlopen(request, timeout=45)
    except urllib.error.HTTPError as error:
        response = error
    with response:
        raw = response.read()
        return response.status, json.loads(raw) if "application/json" in response.headers.get("Content-Type", "") else raw


def create_request():
    return {"command": {"schemaVersion": "agrochain.command.v1", "operationId": ident("OP"), "batchId": ident("BAT"), "expectedVersion": 0, "command": "CreateBatch",
                        "payload": {"productCode": "TOMATO", "gradeCode": "STANDARD", "quantityGrams": 100000, "originRegionCode": "07", "harvestDate": "2026-09-15", "logisticsMsp": "LogisticsMSP", "intendedRetailerMsp": "RetailerMSP"}},
            "scenario": "NORMAL", "privateInput": {}}


class BackendIntegration(unittest.TestCase):
    process = None
    log = None

    @classmethod
    def start(cls, disabled="", endpoint=None):
        DATA.mkdir(parents=True, exist_ok=True)
        cls.log = (DATA / "acceptance.log").open("a")
        env = dict(os.environ, AGROCHAIN_ROOT=str(ROOT), AGROCHAIN_BACKEND_DATA=str(DATA), AGROCHAIN_API_PORT=str(PORT),
                   AGROCHAIN_TOKEN_FILE=str(ROOT / "network/runtime/backend/tokens.json"), AGROCHAIN_DISABLED_SIMULATORS=disabled)
        if endpoint:
            env["AGROCHAIN_GATEWAY_ENDPOINT"] = endpoint
        cls.process = subprocess.Popen([str(ROOT / "network/tools/jdk-21.0.12.1+1/bin/java"), "-jar", str(ROOT / "backend/target/agrochain-backend-0.3.0.jar")], env=env, cwd=ROOT, stdout=cls.log, stderr=cls.log)
        for _ in range(80):
            if cls.process.poll() is not None:
                raise AssertionError("Backend exited at startup; inspect ignored acceptance.log")
            try:
                if call("GET", "/health")[0] == 200:
                    return
            except OSError:
                pass
            time.sleep(0.25)
        raise AssertionError("Backend did not start")

    @classmethod
    def stop(cls):
        if cls.process is not None:
            cls.process.terminate()
            try:
                cls.process.wait(timeout=15)
            except subprocess.TimeoutExpired:
                cls.process.kill(); cls.process.wait(timeout=5)
            cls.process = None
        if cls.log:
            cls.log.close(); cls.log = None

    @classmethod
    def setUpClass(cls):
        cls.addClassCleanup(cls.stop)
        cls.start()

    @classmethod
    def tearDownClass(cls):
        cls.stop()

    def submit(self, actor, path, request, expected=201):
        status, body = call("POST", path, actor, request, {"Idempotency-Key": request["command"]["operationId"]})
        self.assertEqual(status, expected, str(body))
        return body

    def test_01_authentication_and_forged_role(self):
        r = create_request()
        status, body = call("POST", "/batches", body=r, headers={"X-MSP": "ProducerMSP", "X-Role": "producer"})
        self.assertEqual((status, body["code"]), (401, "AUTHENTICATION_REQUIRED"))
        body = self.submit("retailer", "/batches", r, 403)
        self.assertEqual(body["code"], "UNAUTHORIZED_ORGANIZATION")
        status, body = call("POST", "/batches", "producer", r)
        self.assertEqual(status, 400)
        self.assertFalse(any(x in json.dumps(body) for x in ["stackTrace", "certificate", "PRIVATE KEY"]))
        negative = create_request()
        negative["command"]["payload"]["quantityGrams"] = -1
        self.assertEqual(self.submit("producer", "/batches", negative, 400)["code"], "INVALID_QUANTITY")

    def test_02_normal_and_suspicious_full_workflows(self):
        self.__class__.scenarios = []
        for scenario, price in [("NORMAL", 2800), ("SUSPICIOUS", 3700)]:
            create = create_request(); create["scenario"] = scenario
            batch = create["command"]["batchId"]
            pickup, delivery, lot = ident("TRF"), ident("TRF"), ident("LOT")
            steps = [("producer", "CreateBatch", "/batches", create["command"]["payload"]),
                     ("producer", "OfferPickup", "pickup-offers", {"transferId": pickup, "quantityGrams": 100000}),
                     ("logistics", "AcceptPickup", "pickup-acceptances", {"transferId": pickup, "quantityGrams": 100000}),
                     ("logistics", "RecordFreightCost", "freight-costs", {"costId": ident("CST")}),
                     ("logistics", "OfferDelivery", "delivery-offers", {"transferId": delivery, "quantityGrams": 100000}),
                     ("retailer", "AcceptDelivery", "delivery-acceptances", {"transferId": delivery, "quantityGrams": 100000}),
                     ("retailer", "ReportRetailPrice", "retail-reports", {"reportId": ident("RPT"), "retailLotId": lot, "policyId": "CFG-PRICE001"})]
            requests = []
            for i, (actor, fn, route, payload) in enumerate(steps):
                r = copy.deepcopy(create)
                r["command"].update(command=fn, operationId=ident("OP"), expectedVersion=i, payload=payload)
                if i == 6:
                    r["privateInput"] = {"offeredPriceKurusPerKg": price, "currency": "TRY", "taxBasis": "EXCLUDING_TAX", "reportedAt": "2026-09-22T09:00:00.000Z"}
                path = route if i == 0 else f"/batches/{batch}/{route}"
                response = self.submit(actor, path, r)
                self.assertEqual(response["status"], "COMMITTED")
                self.assertEqual(response["receipt"]["resultVersion"], i+1)
                self.assertEqual(response["receipt"]["actorMsp"], {"producer": "ProducerMSP", "logistics": "LogisticsMSP", "retailer": "RetailerMSP"}[actor])
                requests.append((actor, path, r, response))
            status, b = call("GET", f"/batches/{batch}", "producer")
            self.assertEqual((status, b["state"], b["ownerMsp"]), (200, "RETAIL_REPORTED", "RetailerMSP"))
            self.scenarios.append({"batch": batch, "lot": lot, "price": price, "requests": requests})

    def test_03_retry_conflict_and_caller_scoped_status(self):
        for actor, path, r, original in self.scenarios[0]["requests"]:
            again = self.submit(actor, path, r, 200)
            self.assertEqual(again["receipt"], original["receipt"])
            changed = copy.deepcopy(r); changed["scenario"] = "SUSPICIOUS"
            self.assertEqual(self.submit(actor, path, changed, 409)["code"], "IDEMPOTENCY_CONFLICT")
        op = self.scenarios[0]["requests"][0][2]["command"]["operationId"]
        status, body = call("GET", "/operations/"+op, "retailer")
        self.assertEqual((status, body["code"]), (404, "OPERATION_NOT_FOUND"))

    def test_04_private_access_and_no_generic_submit(self):
        batch = self.scenarios[0]["batch"]
        for kind, members in {"purchase": {"producer", "retailer", "regulator"}, "freight": {"logistics", "retailer", "regulator"}, "retail": {"retailer", "regulator"}}.items():
            for actor in ACTORS:
                status, body = call("GET", f"/batches/{batch}/commercial?kind={kind}", actor)
                self.assertEqual(status, 200 if actor in members else 403)
        for scenario in self.scenarios:
            status, price = call("GET", f'/batches/{scenario["batch"]}/commercial?kind=retail', "retailer")
            self.assertEqual(price["offeredPriceKurusPerKg"], scenario["price"])
        self.assertEqual(call("POST", f"/batches/{batch}/GetPurchase", "producer", {})[0], 400)
        invalid = create_request()
        invalid["command"].update(batchId=batch, operationId=ident("OP"), expectedVersion=7, command="OfferDelivery", payload={"transferId": ident("TRF"), "quantityGrams": 100000})
        self.assertEqual(self.submit("logistics", f"/batches/{batch}/delivery-offers", invalid, 409)["code"], "INVALID_STATE_TRANSITION")

    def test_05_original_modified_document_and_archive_authorization(self):
        batch = self.scenarios[0]["batch"]
        _, purchase = call("GET", f"/batches/{batch}/commercial?kind=purchase", "producer")
        docid = purchase["envelope"]["header"]["documentId"]
        status, original = call("GET", f"/documents/{docid}", "producer")
        self.assertEqual(status, 200)
        self.assertEqual(call("GET", f"/documents/{docid}", "logistics")[0], 403)
        self.assertEqual(call("POST", f"/documents/{docid}/verify", "producer", {"originalBase64": base64.b64encode(original).decode()})[1]["attachmentMatched"], True)
        status, error = call("POST", f"/documents/{docid}/verify", "producer", {"originalBase64": base64.b64encode(original+b"\n").decode()})
        self.assertEqual((status, error["code"]), (422, "DOCUMENT_HASH_MISMATCH"))

    def test_06_public_projection_allowlist_and_replay(self):
        allowed = {"schemaVersion", "retailLotId", "batchId", "productCode", "gradeCode", "quantityGrams", "originRegionCode", "harvestDate", "state", "version", "organizations", "history", "sources", "asOfBlock", "updatedTxId"}
        for s in self.scenarios:
            for _ in range(40):
                status, result = call("GET", "/public/lots/"+s["lot"])
                if status == 200:
                    break
                time.sleep(0.25)
            self.assertEqual(status, 200)
            self.assertEqual(set(result), allowed)
            self.assertEqual(len(result["history"]), 3)
            self.assertEqual({x["sourceSystem"] for x in result["sources"]}, {"CKS", "EFATURA", "HKS", "UETDS"})
            self.assertTrue(all(x["sourceMode"] == "SIMULATED" for x in result["sources"]))
            for forbidden in ["Kurus", "saltHex", "signature", "commitment", "attachmentDigest", "anomaly"]:
                self.assertNotIn(forbidden, json.dumps(result))

    def test_07_simulator_failure_and_same_operation_retry(self):
        self.stop(); self.start(disabled="CKS")
        r = create_request()
        self.assertEqual(self.submit("producer", "/batches", r, 503)["code"], "SIMULATOR_UNAVAILABLE")
        self.stop(); self.start()
        response = self.submit("producer", "/batches", r, 200)
        self.assertEqual(response["status"], "COMMITTED")

    def test_08_fabric_failure_and_recovery(self):
        self.stop(); self.start(endpoint="localhost:1")
        r = create_request()
        self.assertEqual(self.submit("producer", "/batches", r, 503)["code"], "FABRIC_UNAVAILABLE")
        self.stop(); self.start()
        self.assertEqual(self.submit("producer", "/batches", r, 200)["status"], "COMMITTED")

    def test_09_restart_committed_but_unpromoted_reconciliation(self):
        # Reconstruct the durable crash window AFTER a real VALID commit, BEFORE
        # off-chain promotion/checkpoint. This is explicit fault injection, not a
        # claim that the network itself was mocked or an actual process crash timed.
        actor, path, r, committed = self.scenarios[0]["requests"][1]
        op = r["command"]["operationId"]
        self.stop()
        with sqlite3.connect(DATA / "operations.sqlite") as db:
            db.execute("UPDATE operations SET state='SUBMITTED_UNKNOWN',receipt=NULL WHERE actor=? AND id=?", (ACTORS[actor], op))
            db.execute("UPDATE evidence SET committed=0 WHERE actor=? AND operation=?", (ACTORS[actor], op))
            db.execute("DELETE FROM projections")
            db.execute("DELETE FROM events")
        self.start()
        for _ in range(60):
            _, status = call("GET", "/operations/"+op, actor)
            if status["status"] == "COMMITTED":
                break
            time.sleep(0.25)
        self.assertEqual(status["receipt"], committed["receipt"])
        self.test_06_public_projection_allowlist_and_replay()
        self.assertEqual(self.submit(actor, path, r, 200)["receipt"], committed["receipt"])
        with sqlite3.connect(DATA / "operations.sqlite") as db:
            self.assertEqual(db.execute("SELECT COUNT(*) FROM evidence WHERE actor=? AND operation=? AND committed=0", (ACTORS[actor], op)).fetchone()[0], 0)

    def test_10_no_private_credentials_in_logs(self):
        log = (DATA / "acceptance.log").read_text()
        for token in TOKENS.values():
            self.assertNotIn(token, log)
        self.assertNotIn("PRIVATE KEY", log)
        with sqlite3.connect(DATA / "operations.sqlite") as db:
            for (bundle,) in db.execute("SELECT bundle FROM evidence"):
                self.assertNotIn(json.loads(bundle)["saltHex"], log)
        (DATA / "acceptance-summary.json").write_text(json.dumps({"sourceMode": "SIMULATED", "blockchainMode": "FABRIC", "batches": [{"batchId": s["batch"], "lotId": s["lot"]} for s in self.scenarios], "privateValuesOmitted": True}, indent=2)+"\n")


if __name__ == "__main__":
    os.umask(0o077)
    unittest.main(verbosity=2, failfast=True)
