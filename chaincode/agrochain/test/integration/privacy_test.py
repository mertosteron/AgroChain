#!/usr/bin/env python3
"""Real Stage 4 lifecycle, PDC exclusion, integrity and restart acceptance."""
import base64
import copy
import hashlib
import json
import os
import subprocess
import tempfile
import unittest
from concurrent.futures import ThreadPoolExecutor
from threading import Barrier
from evidence_support import ROOT, canonical, bundle, identifier, configuration, sign_header, verify_original

CLIENT = ROOT / "network/scripts/chaincode-client.sh"
ORGS = ("producer", "logistics", "retailer", "regulator")
ROLES = dict(zip(ORGS, ("producer", "carrier", "retailer", "auditor")))


def call(org, action, args=None, transient=None, role=None, peer=None):
    env = dict(os.environ, AGROCHAIN_ROLE=role or ROLES[org])
    if peer:
        env["AGROCHAIN_QUERY_PEER"] = peer
    with tempfile.TemporaryDirectory(prefix="agrochain-transient-") as tmp:
        if transient is not None:
            f = os.path.join(tmp, "transient.json")
            with open(f, "w") as out:
                json.dump({k: base64.b64encode(canonical(v)).decode() for k, v in transient.items()}, out)
            os.chmod(f, 0o600)
            env["AGROCHAIN_TRANSIENT_FILE"] = f
        cmd = ["bash", str(CLIENT), org, action]
        if args is not None:
            cmd.append(json.dumps({"Args": args}) if isinstance(args, list) else str(args))
        return subprocess.run(cmd, env=env, capture_output=True, text=True, timeout=90)


def query(org, fn, arg=None, **kw):
    r = call(org, "query", [fn] + ([] if arg is None else [arg]), **kw)
    if r.returncode:
        raise AssertionError(fn + ": " + r.stderr)
    return json.loads(r.stdout)


def height():
    r = call("regulator", "info")
    assert r.returncode == 0
    return json.loads(r.stdout.split("Blockchain info: ", 1)[1])["height"]


def command(fn, payload, batch, version):
    return {"schemaVersion": "agrochain.command.v1", "operationId": identifier("OP"), "batchId": batch,
            "expectedVersion": version, "command": fn, "payload": payload}


class PrivacyTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.config = configuration()
        existing = call("regulator", "query", ["GetConfiguration"])
        if existing.returncode:
            assert "CONFIGURATION_REQUIRED" in existing.stderr
            r = call("regulator", "invoke", ["Bootstrap", canonical(cls.config).decode()], role="admin")
            assert r.returncode == 0, r.stderr
        else:
            assert json.loads(existing.stdout) == cls.config, "Local signing keys differ from immutable ledger trust"
        cls.batch = identifier("BAT")
        cls.original = b'<invoice id="STAGE4-SYNTHETIC">200000</invoice>'
        create = dict(productCode="TOMATO", gradeCode="STANDARD", quantityGrams=100000, originRegionCode="07",
                      harvestDate="2026-09-15", logisticsMsp="LogisticsMSP", intendedRetailerMsp="RetailerMSP")
        pickup, delivery = identifier("TRF"), identifier("TRF")
        cls.flows = [("producer", "CreateBatch", create), ("producer", "OfferPickup", {"transferId": pickup, "quantityGrams": 100000}),
                     ("logistics", "AcceptPickup", {"transferId": pickup, "quantityGrams": 100000}),
                     ("logistics", "RecordFreightCost", {"costId": identifier("CST")}),
                     ("logistics", "OfferDelivery", {"transferId": delivery, "quantityGrams": 100000}),
                     ("retailer", "AcceptDelivery", {"transferId": delivery, "quantityGrams": 100000}),
                     ("retailer", "ReportRetailPrice", {"reportId": identifier("RPT"), "retailLotId": identifier("LOT"), "policyId": "CFG-PRICE001"})]
        cls.commands = [command(fn, p, cls.batch, i) for i, (_, fn, p) in enumerate(cls.flows)]
        cls.transients = [{} for _ in cls.flows]
        cls.transients[0] = {"evidence": [bundle("CKS", "PRODUCER_ELIGIBILITY", dict(producerMsp="ProducerMSP", productCode="TOMATO",
                            gradeCode="STANDARD", quantityGrams=100000, originRegionCode="07", harvestDate="2026-09-15", eligible=True), cls.commands[0])]}
        cls.purchase = bundle("EFATURA", "PURCHASE_INVOICE", dict(sellerMsp="ProducerMSP", buyerMsp="RetailerMSP", quantityGrams=100000,
                     priceKurusPerKg=2000, totalKurus=200000, currency="TRY", taxBasis="EXCLUDING_TAX",
                     attachmentDigest=hashlib.sha256(cls.original).hexdigest(), attachmentMediaType="application/xml"), cls.commands[1])
        docs = [cls.purchase, bundle("HKS", "TRADE_NOTIFICATION", dict(producerMsp="ProducerMSP", retailerMsp="RetailerMSP", productCode="TOMATO", quantityGrams=100000, notified=True), cls.commands[1]),
                bundle("UETDS", "TRANSPORT_MANIFEST", dict(carrierMsp="LogisticsMSP", consignorMsp="ProducerMSP", consigneeMsp="RetailerMSP", productCode="TOMATO", quantityGrams=100000, declaredDepartureAt="2026-09-21T09:00:00.000Z"), cls.commands[1])]
        cls.transients[1] = {"evidence": sorted(docs, key=lambda d: d["envelope"]["header"]["documentId"])}
        cls.transients[3] = {"evidence": [bundle("EFATURA", "FREIGHT_INVOICE", dict(carrierMsp="LogisticsMSP", payerMsp="RetailerMSP", quantityGrams=100000, totalKurus=20000,
                             currency="TRY", taxBasis="EXCLUDING_TAX", attachmentDigest=hashlib.sha256(b"freight").hexdigest(), attachmentMediaType="application/xml"), cls.commands[3])], "recordSaltHex": os.urandom(32).hex()}
        cls.transients[6] = {"retailReportInput": dict(offeredPriceKurusPerKg=2800, currency="TRY", taxBasis="EXCLUDING_TAX", reportedAt="2026-09-21T09:30:00.000Z", recordSaltHex=os.urandom(32).hex())}
        cls.blocks = []

    def rejected(self, org, args, code, **kw):
        before = height()
        r = call(org, "invoke", args, **kw)
        self.assertNotEqual(r.returncode, 0)
        self.assertIn(code, r.stderr)
        self.assertEqual(before, height(), "rejected proposal was ordered")

    def test_01_bootstrap_authority_and_immutability(self):
        self.rejected("producer", ["Bootstrap", canonical(self.config).decode()], "UNAUTHORIZED_ORGANIZATION", role="admin")
        self.rejected("regulator", ["Bootstrap", canonical(self.config).decode()], "UNAUTHORIZED_ROLE")
        self.rejected("regulator", ["Bootstrap", canonical(self.config).decode()], "CONFIGURATION_ALREADY_EXISTS", role="admin")

    def test_02_invalid_evidence_cannot_create(self):
        for kind, expected in [("signature", "INVALID_SOURCE_SIGNATURE"), ("salt", "DOCUMENT_HASH_MISMATCH"), ("binding", "SOURCE_BINDING_MISMATCH"), ("key", "UNTRUSTED_SOURCE_KEY")]:
            with self.subTest(kind=kind):
                t = copy.deepcopy(self.transients[0]); doc = t["evidence"][0]
                if kind == "signature": doc["envelope"]["signature"] = base64.urlsafe_b64encode(bytes(64)).decode().rstrip("=")
                if kind == "salt": doc["saltHex"] = os.urandom(32).hex()
                if kind == "key": doc["envelope"]["header"]["keyId"] = "UNKNOWN"
                if kind == "binding":
                    doc["envelope"]["header"]["boundOperationId"] = identifier("OP"); sign_header(doc)
                self.rejected("producer", ["CreateBatch", canonical(self.commands[0]).decode()], expected, transient=t)
                self.assertFalse(query("producer", "BatchExists", self.batch))
        self.rejected("producer", ["CreateBatch", canonical(self.commands[0]).decode()], "UNAUTHORIZED_ROLE", transient=self.transients[0], role="none")
        self.rejected("producer", ["CreateBatch", canonical(self.commands[0]).decode()], "UNAUTHORIZED_ROLE", transient=self.transients[0], role="admin")

    def test_03_live_product_lifecycle(self):
        states = ["CREATED", "PICKUP_PENDING", "IN_TRANSPORT", "IN_TRANSPORT", "DELIVERY_PENDING", "RECEIVED", "RETAIL_REPORTED"]
        for i, (org, fn, _) in enumerate(self.flows):
            before = height()
            r = call(org, "invoke", [fn, canonical(self.commands[i]).decode()], transient=self.transients[i])
            self.assertEqual(r.returncode, 0, fn + r.stderr)
            self.assertIn("committed with status (VALID)", r.stderr)
            self.blocks.append(before)
            b = query(org, "GetBatch", self.batch)
            self.assertEqual((b["state"], b["version"], b["quantityGrams"]), (states[i], i+1, 100000))
            self.assertEqual(b["ownerMsp"], "ProducerMSP" if i < 5 else "RetailerMSP")
            receipt = query(org, "GetOperation", self.commands[i]["operationId"])
            self.assertEqual(receipt["txId"], b["updatedTxId"])

    def test_04_real_pdc_visibility(self):
        allowed = {"GetPurchase": {"producer", "retailer", "regulator"}, "GetFreightCost": {"logistics", "retailer", "regulator"}, "GetRetailReport": {"retailer", "regulator"}}
        for fn, members in allowed.items():
            for org in ORGS:
                r = call(org, "query", [fn, self.batch])
                if org in members: self.assertEqual(r.returncode, 0, fn+r.stderr)
                else:
                    self.assertNotEqual(r.returncode, 0); self.assertIn("PRIVATE_DATA_ACCESS_DENIED", r.stderr)
                self.assertEqual(query(org, "GetBatch", self.batch)["state"], "RETAIL_REPORTED")
            for peer in set(ORGS)-members:
                r = call("regulator", "query", [fn, self.batch], peer=peer)
                self.assertNotEqual(r.returncode, 0); self.assertIn("PRIVATE_DATA_UNAVAILABLE", r.stderr)
        r = call("regulator", "query", ["GetRetailReport", self.batch], role="public-reader")
        self.assertNotEqual(r.returncode, 0); self.assertIn("PRIVATE_DATA_ACCESS_DENIED", r.stderr)

    def test_05_original_and_tampered_document(self):
        committed = query("producer", "GetPurchase", self.batch)
        self.assertTrue(verify_original(committed, self.original, query("regulator", "GetConfiguration")))
        with self.assertRaisesRegex(ValueError, "DOCUMENT_HASH_MISMATCH"):
            verify_original(committed, self.original+b"\n", self.config)
        opening = {k: committed[k] for k in ["body", "saltHex"]}
        docid = committed["envelope"]["header"]["documentId"]
        self.assertTrue(query("producer", "VerifyDocument", docid, transient={"opening": opening})["commitmentMatched"])
        opening["body"]["totalKurus"] += 1
        r = call("producer", "query", ["VerifyDocument", docid], transient={"opening": opening})
        self.assertNotEqual(r.returncode, 0); self.assertIn("DOCUMENT_HASH_MISMATCH", r.stderr)

    def test_06_operation_and_source_replay(self):
        for i, (org, fn, _) in enumerate(self.flows):
            self.rejected(org, [fn, canonical(self.commands[i]).decode()], "DUPLICATE_TRANSACTION", transient=self.transients[i])
        for kind, expected in [("sourceDocumentId", "DOCUMENT_ALREADY_USED"), ("nonce", "SOURCE_NONCE_ALREADY_USED")]:
            c = command("CreateBatch", self.flows[0][2], identifier("BAT"), 0)
            doc = bundle("CKS", "PRODUCER_ELIGIBILITY", self.transients[0]["evidence"][0]["body"], c)
            doc["envelope"]["header"][kind] = self.transients[0]["evidence"][0]["envelope"]["header"][kind]; sign_header(doc)
            self.rejected("producer", ["CreateBatch", canonical(c).decode()], expected, transient={"evidence": [doc]})
            self.assertFalse(query("producer", "BatchExists", c["batchId"]))

    def test_07_blocks_events_and_private_hashes(self):
        private_blocks = 0
        for i, block in enumerate(self.blocks):
            r = call("regulator", "block", block); self.assertEqual(r.returncode, 0)
            data = json.loads(r.stdout)
            self.assertEqual(base64.b64decode(data["metadata"]["metadata"][2]), b"\0")
            extension = data["data"]["data"][0]["payload"]["data"]["actions"][0]["payload"]["action"]["proposal_response_payload"]["extension"]
            self.assertTrue(extension["events"])
            event = json.loads(base64.b64decode(extension["events"]["payload"]))
            self.assertEqual(set(event), {"schemaVersion", "eventType", "operationId", "batchId", "actorMsp", "txId", "txTime", "resultVersion", "objectIds", "previousState", "newState"})
            self.assertEqual(event["operationId"], self.commands[i]["operationId"])
            self.assertEqual(event["batchId"], self.batch)
            self.assertEqual(event["resultVersion"], i+1)
            self.assertEqual(event["txId"], extension["events"]["tx_id"])
            for ns in extension["results"]["ns_rwset"]:
                if ns["namespace"] != "agrochain": continue
                if ns.get("collection_hashed_rwset"): private_blocks += 1
                for w in ns["rwset"].get("writes", []):
                    if w.get("is_delete"): continue
                    raw = base64.b64decode(w["value"]).decode()
                    for field in ["priceKurusPerKg", "totalKurus", "recordSaltHex", "saltHex", "attachmentDigest", "offeredPriceKurusPerKg"]:
                        self.assertNotIn('"'+field+'"', raw)
        self.assertEqual(private_blocks, 3)
        # Inspect actual service logs without printing them or confidential needles.
        needles = [self.purchase["saltHex"], self.transients[3]["recordSaltHex"], self.transients[6]["retailReportInput"]["recordSaltHex"]]
        for org in ORGS:
            for container in ["agrochain-"+org, "agrochain-cc-"+org]:
                result = subprocess.run(["docker", "logs", container], capture_output=True, text=True, timeout=20)
                self.assertEqual(result.returncode, 0)
                for needle in needles:
                    self.assertNotIn(needle, result.stdout+result.stderr, "private salt leaked to service log")

    def test_08_mvcc_duplicate_concurrent_proposals(self):
        # Both signed proposals read the same absent batch/operation/source keys.
        c = command("CreateBatch", self.flows[0][2], identifier("BAT"), 0)
        doc = bundle("CKS", "PRODUCER_ELIGIBILITY", self.transients[0]["evidence"][0]["body"], c)
        barrier = Barrier(2)
        def submit():
            barrier.wait()
            return call("producer", "invoke", ["CreateBatch", canonical(c).decode()], transient={"evidence": [doc]})
        before = height()
        with ThreadPoolExecutor(max_workers=2) as pool:
            results = list(pool.map(lambda _: submit(), range(2)))
        self.assertEqual(sum(r.returncode == 0 for r in results), 1, str([r.stderr for r in results]))
        self.assertTrue(any("MVCC_READ_CONFLICT" in r.stderr for r in results), "Both proposals must reach ordering for MVCC proof")
        flags = []
        for block in range(before, height()):
            r = call("regulator", "block", block); self.assertEqual(r.returncode, 0)
            flags.extend(base64.b64decode(json.loads(r.stdout)["metadata"]["metadata"][2]))
        self.assertEqual(sorted(flags), [0, 11])
        self.assertEqual(query("producer", "GetBatch", c["batchId"])["version"], 1)
        self.assertEqual(query("producer", "GetOperation", c["operationId"])["txId"], query("producer", "GetBatch", c["batchId"])["updatedTxId"])

    def test_09_populated_restart(self):
        def snapshot():
            return {"height": height(), "batch": query("producer", "GetBatch", self.batch),
                    "trade": query("producer", "GetPurchase", self.batch), "freight": query("logistics", "GetFreightCost", self.batch),
                    "retail": query("retailer", "GetRetailReport", self.batch)}
        before = snapshot()
        for target in ["network-down", "network-up", "chaincode-check"]:
            r = subprocess.run(["make", target], cwd=ROOT, capture_output=True, text=True, timeout=180)
            self.assertEqual(r.returncode, 0, r.stderr[-1200:])
        self.assertEqual(before, snapshot())
        dest = ROOT / "network/runtime/chaincode/privacy-evidence.json"
        dest.write_text(json.dumps({"batchId": self.batch, "blocks": self.blocks, "height": height(), "populatedRestartPreserved": True,
                                    "sourceMode": "SIMULATED", "privateValuesOmitted": True}, indent=2)+"\n")


if __name__ == "__main__":
    unittest.main(verbosity=2)
