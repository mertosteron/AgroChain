"""Signed local fixtures and original-byte checks; not institutional adapters."""
import base64
import hashlib
import json
import os
from pathlib import Path
import struct
from cryptography.hazmat.primitives import serialization
from cryptography.hazmat.primitives.asymmetric.ed25519 import Ed25519PublicKey

ROOT = Path(__file__).resolve().parents[4]
TYPES = {"CKS": ["PRODUCER_ELIGIBILITY"], "EFATURA": ["FREIGHT_INVOICE", "PURCHASE_INVOICE"],
         "HKS": ["TRADE_NOTIFICATION"], "UETDS": ["TRANSPORT_MANIFEST"]}


def canonical(value):
    return json.dumps(value, sort_keys=True, separators=(",", ":"), ensure_ascii=False).encode()


def b64(value):
    return base64.urlsafe_b64encode(value).decode().rstrip("=")


def identifier(prefix):
    return prefix + "-" + base64.b32encode(os.urandom(16)).decode().rstrip("=")


def key(system):
    return serialization.load_pem_private_key((ROOT / f"network/runtime/source-keys/{system}.pem").read_bytes(), None)


def configuration():
    return {"schemaVersion": "agrochain.bootstrap.v1", "policyId": "CFG-PRICE001", "thresholdBps": 5000,
            "sources": [{"schemaVersion": "agrochain.source-key.v1", "issuerId": "SIM_" + s, "sourceSystem": s,
                         "documentTypes": ts, "keyId": "SIM_" + s + "_KEY01", "enabled": True,
                         "publicKeyRawBase64url": b64(key(s).public_key().public_bytes_raw())} for s, ts in TYPES.items()]}


def commitment(body, salt):
    raw = canonical(body)
    return hashlib.sha256(b"AgroChain/document/v1\0" + bytes.fromhex(salt) + struct.pack(">Q", len(raw)) + raw).hexdigest()


def sign_header(doc):
    h = doc["envelope"]["header"]
    doc["envelope"]["signature"] = b64(key(h["sourceSystem"]).sign(b"AgroChain/source-signature/v1\0" + canonical(h)))


def bundle(system, kind, body, command):
    salt = os.urandom(32).hex()
    h = {"schemaVersion": "agrochain.source-document.v1", "documentId": identifier("DOC"), "sourceSystem": system,
         "sourceMode": "SIMULATED", "issuerId": "SIM_" + system, "keyId": "SIM_" + system + "_KEY01",
         "sourceDocumentId": identifier("DOC"), "documentType": kind, "batchId": command["batchId"],
         "boundOperationId": command["operationId"], "issuedAt": "2026-09-21T09:00:00.000Z", "nonce": os.urandom(32).hex(),
         "commitmentAlgorithm": "SHA256_SALTED_JCS_V1", "commitment": commitment(body, salt), "signatureAlgorithm": "Ed25519"}
    doc = {"envelope": {"header": h}, "body": body, "saltHex": salt}
    sign_header(doc)
    return doc


def verify_original(committed_bundle, original, trusted_config):
    """Read-only off-chain boundary: verify against a bundle queried from Fabric."""
    if len(original) > 5 * 1024 * 1024:
        raise ValueError("INVALID_SCHEMA")
    h = committed_bundle["envelope"]["header"]
    candidates = [s for s in trusted_config["sources"] if s["issuerId"] == h["issuerId"] and s["keyId"] == h["keyId"]
                  and s["sourceSystem"] == h["sourceSystem"] and h["documentType"] in s["documentTypes"] and s["enabled"]]
    if len(candidates) != 1 or h["sourceMode"] != "SIMULATED":
        raise ValueError("UNTRUSTED_SOURCE_KEY")
    Ed25519PublicKey.from_public_bytes(base64.urlsafe_b64decode(candidates[0]["publicKeyRawBase64url"] + "==")).verify(
        base64.urlsafe_b64decode(committed_bundle["envelope"]["signature"] + "=="), b"AgroChain/source-signature/v1\0" + canonical(h))
    if commitment(committed_bundle["body"], committed_bundle["saltHex"]) != h["commitment"]:
        raise ValueError("DOCUMENT_HASH_MISMATCH")
    if hashlib.sha256(original).hexdigest() != committed_bundle["body"]["attachmentDigest"]:
        raise ValueError("DOCUMENT_HASH_MISMATCH")
    return True
