#!/usr/bin/env python3
"""Live domain rejection and endorsement regression tests, updated for Stage 4."""
import base64
import json
import os
from pathlib import Path
import re
import subprocess
import unittest

ROOT = Path(__file__).resolve().parents[4]
CLIENT = ROOT / "network/scripts/chaincode-client.sh"
ORGS = ("producer", "logistics", "retailer", "regulator")
BATCH = "BAT-LIVE0001"
VERSION = re.search(r"^CHAINCODE_VERSION=(.+)$", (ROOT / "network/config/chaincode.env").read_text(), re.M)[1]


def endorser_msp(encoded):
    binary = Path(os.environ.get("FABRIC_BIN_DIR", ROOT / "network/tools/bin")) / "configtxlator"
    decoded = subprocess.run([str(binary), "proto_decode", "--type", "msp.SerializedIdentity", "--input", "/dev/stdin"],
                             input=base64.b64decode(encoded), capture_output=True, timeout=10)
    assert decoded.returncode == 0, decoded.stderr.decode()
    return json.loads(decoded.stdout)["mspid"]


def call(org, action, value=None):
    args = ["bash", str(CLIENT), org, action]
    if value is not None:
        args.append(json.dumps(value, separators=(",", ":")) if not isinstance(value, str) else value)
    role = dict(producer="producer", logistics="carrier", retailer="retailer", regulator="auditor")[org]
    return subprocess.run(args, env=dict(os.environ, AGROCHAIN_ROLE=role), capture_output=True, text=True, timeout=90)


def query(org, fn, arg=None):
    response = call(org, "query", {"Args": [fn] + ([] if arg is None else [arg])})
    if response.returncode:
        raise AssertionError(f"{fn} exit={response.returncode}: {response.stderr}")
    return json.loads(response.stdout)


def height(org="regulator"):
    result = call(org, "info")
    assert result.returncode == 0, result.stderr
    return json.loads(result.stdout.split("Blockchain info: ", 1)[1])["height"]


def command(fn, payload, index=0):
    return {"schemaVersion": "agrochain.command.v1", "operationId": f"OP-LIVE{index:04d}",
            "batchId": BATCH, "expectedVersion": index, "command": fn, "payload": payload}


CREATE = {"productCode": "TOMATO", "gradeCode": "STANDARD", "quantityGrams": 100000,
          "originRegionCode": "07", "harvestDate": "2026-09-15",
          "logisticsMsp": "LogisticsMSP", "intendedRetailerMsp": "RetailerMSP"}
FLOWS = [("producer", "CreateBatch", CREATE),
         ("producer", "OfferPickup", {"transferId": "TRF-PICKUP01", "quantityGrams": 100000}),
         ("logistics", "AcceptPickup", {"transferId": "TRF-PICKUP01", "quantityGrams": 100000}),
         ("logistics", "RecordFreightCost", {"costId": "CST-FREIGHT1"}),
         ("logistics", "OfferDelivery", {"transferId": "TRF-DELIVER1", "quantityGrams": 100000}),
         ("retailer", "AcceptDelivery", {"transferId": "TRF-DELIVER1", "quantityGrams": 100000}),
         ("retailer", "ReportRetailPrice", {"reportId": "RPT-RETAIL01", "retailLotId": "LOT-RETAIL01", "policyId": "CFG-PRICE001"})]


class FabricTests(unittest.TestCase):
    def assert_rejected(self, org, fn, payload, expected, index=0):
        before = height()
        envelope = command(fn, payload, index)
        result = call(org, "invoke", {"Args": [fn, json.dumps(envelope, separators=(",", ":"))]})
        self.assertNotEqual(result.returncode, 0, result.stdout)
        # CLI escapes the chaincode JSON when rendering an error response.
        normalized = (result.stdout + result.stderr).replace('\\"', '"')
        self.assertIn(f'"code":"{expected}"', normalized)
        self.assertEqual(height(), before, "rejected proposal was ordered")
        for member in ORGS:
            self.assertFalse(query(member, "BatchExists", BATCH))
        receipt = call(org, "query", {"Args": ["GetOperation", envelope["operationId"]]})
        self.assertNotEqual(receipt.returncode, 0)
        self.assertIn("OPERATION_NOT_FOUND", receipt.stderr)

    def test_01_public_queries_all_four_identities(self):
        for org in ORGS:
            with self.subTest(org=org):
                health = query(org, "Health")
                self.assertEqual(health["version"], VERSION)
                self.assertEqual(health["evidenceVerification"], "REQUIRED_STAGE_4")
                self.assertFalse(query(org, "BatchExists", BATCH))
                for fn, filt in (("GetBatchHistory", BATCH), ("QueryBatchesByOwner", "ProducerMSP"), ("QueryBatchesByState", "CREATED")):
                    page = query(org, fn, json.dumps({"schemaVersion": "agrochain.page-request.v1", "filter": filt, "pageSize": 10, "bookmark": ""}))
                    self.assertEqual(page["schemaVersion"], "agrochain.page.v1")
                    self.assertLessEqual(len(page["records"]), 10)
                    if fn == "GetBatchHistory": self.assertEqual(page["records"], [])

    def test_02_unverified_business_writes_fail_closed(self):
        for i, (org, fn, payload) in enumerate(FLOWS):
            with self.subTest(fn=fn):
                configured = query(org, "Health")["writesEnabled"]
                expected = ("MISSING_EVIDENCE" if i == 0 else "BATCH_NOT_FOUND") if configured else "CONFIGURATION_REQUIRED"
                self.assert_rejected(org, fn, payload, expected, i)

    def test_03_wrong_organizations(self):
        for org, fn, payload, error in (
            ("logistics", "CreateBatch", CREATE, "UNAUTHORIZED_ORGANIZATION"),
            ("retailer", "AcceptPickup", FLOWS[2][2], "WRONG_TRANSFER_RECIPIENT"),
            ("producer", "ReportRetailPrice", FLOWS[6][2], "UNAUTHORIZED_ORGANIZATION"),
            ("regulator", "OfferPickup", FLOWS[1][2], "UNAUTHORIZED_ORGANIZATION"),
            ("retailer", "OfferDelivery", FLOWS[4][2], "UNAUTHORIZED_ORGANIZATION"),
        ):
            with self.subTest(org=org, fn=fn):
                self.assert_rejected(org, fn, payload, error)

    def test_04_malformed_input_and_public_price_rejected(self):
        for quantity in (-1, 0, 100000001):
            self.assert_rejected("producer", "CreateBatch", dict(CREATE, quantityGrams=quantity), "INVALID_QUANTITY")
        self.assert_rejected("producer", "CreateBatch", dict(CREATE, quantityGrams="100000"), "INVALID_SCHEMA")
        self.assert_rejected("producer", "CreateBatch", dict(CREATE, unit="kg"), "INVALID_SCHEMA")
        self.assert_rejected("retailer", "ReportRetailPrice", dict(FLOWS[6][2], offeredPriceKurusPerKg=-1), "INVALID_SCHEMA")

    def test_05_endorsed_health_commits_have_no_domain_writes_or_events(self):
        evidence = []
        for org in ORGS:
            before = height()
            result = call(org, "invoke", {"Args": ["Health"]})
            self.assertEqual(result.returncode, 0, result.stderr)
            txids = re.findall(r"txid \[([0-9a-f]{64})\]", result.stderr)
            self.assertTrue(txids, "CLI did not report committed tx ID")
            self.assertIn("committed with status (VALID)", result.stderr)
            self.assertEqual(height(), before + 1)
            block = call(org, "block", str(before))
            self.assertEqual(block.returncode, 0, block.stderr)
            data = json.loads(block.stdout)
            # Fabric transaction validation flags: zero is VALID.
            self.assertEqual(base64.b64decode(data["metadata"]["metadata"][2]), b"\x00")
            envelopes = data["data"]["data"]
            self.assertEqual(len(envelopes), 1)
            payload = envelopes[0]["payload"]
            txid = payload["header"]["channel_header"]["tx_id"]
            self.assertIn(txid, txids)
            action = payload["data"]["actions"][0]["payload"]["action"]
            extension = action["proposal_response_payload"]["extension"]
            self.assertFalse(extension.get("events"), "health must not emit business events")
            for ns in extension["results"]["ns_rwset"]:
                if ns["namespace"] == "agrochain":
                    self.assertFalse(ns["rwset"].get("writes"))
                    self.assertFalse(ns.get("collection_hashed_rwset"))
            self.assertEqual(len(action["endorsements"]), 2)
            self.assertEqual({endorser_msp(e["endorser"]) for e in action["endorsements"]}, {"RetailerMSP", "RegulatorMSP"})
            evidence.append({"actor": org, "txId": txid, "block": before, "valid": True})
        output = ROOT / "network/runtime/chaincode/integration-commits.json"
        output.write_text(json.dumps(evidence, indent=2) + "\n")

    def test_06_single_regulator_endorsement_cannot_commit_validly(self):
        before = height()
        result = call("regulator", "invoke-one", {"Args": ["Health"]})
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("ENDORSEMENT_POLICY_FAILURE", result.stderr)
        self.assertEqual(height(), before + 1)
        block = call("regulator", "block", str(before))
        self.assertEqual(block.returncode, 0, block.stderr)
        data = json.loads(block.stdout)
        self.assertEqual(base64.b64decode(data["metadata"]["metadata"][2]), b"\x0a")
        for org in ORGS:
            self.assertFalse(query(org, "BatchExists", BATCH))


if __name__ == "__main__":
    unittest.main(verbosity=2)
