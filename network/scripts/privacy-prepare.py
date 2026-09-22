#!/usr/bin/env python3
"""Development-only role certificates and signed-fixture keys. Never a government CA."""
import datetime
import json
import os
from pathlib import Path
import shutil

from cryptography import x509
from cryptography.hazmat.primitives import hashes, serialization
from cryptography.hazmat.primitives.asymmetric import ec, ed25519
from cryptography.x509.oid import NameOID, ObjectIdentifier

ROOT = Path(__file__).resolve().parents[2]
RUNTIME = ROOT / "network/runtime"
ROLES = {"producer": ["producer", "admin"], "logistics": ["carrier", "admin"],
         "retailer": ["retailer", "admin"], "regulator": ["admin", "auditor", "reviewer", "oracle", "public-reader"]}


def prepare():
    os.umask(0o077)
    for org, roles in ROLES.items():
        base = ROOT / f"network/organizations/peerOrganizations/{org}.agrochain.test"
        ca_cert = x509.load_pem_x509_certificate(next((base / "ca").glob("*.pem")).read_bytes())
        ca_key = serialization.load_pem_private_key(next((base / "ca").glob("*_sk")).read_bytes(), None)
        for role in roles:
            dest = RUNTIME / f"identities/{org}/{role}/msp"
            cert_path, key_path = dest / "signcerts/cert.pem", dest / "keystore/key.pem"
            if cert_path.exists() != key_path.exists():
                raise RuntimeError("Incomplete role identity; restore it before continuing")
            if cert_path.exists():
                cert = x509.load_pem_x509_certificate(cert_path.read_bytes())
                cert.verify_directly_issued_by(ca_cert)
                if cert.not_valid_after_utc <= datetime.datetime.now(datetime.timezone.utc):
                    raise RuntimeError("Expired demo role certificate")
                continue
            for d in ["signcerts", "keystore", "cacerts"]:
                (dest / d).mkdir(parents=True, exist_ok=True)
            k = ec.generate_private_key(ec.SECP256R1())
            now = datetime.datetime.now(datetime.timezone.utc)
            subject = x509.Name([x509.NameAttribute(NameOID.COMMON_NAME, f"DEVELOPMENT-{org}-{role}"),
                                 x509.NameAttribute(NameOID.ORGANIZATIONAL_UNIT_NAME, "admin" if role == "admin" else "client")])
            attrs = json.dumps({"attrs": {"agrochain.role": role}}, separators=(",", ":")).encode()
            cert = (x509.CertificateBuilder().subject_name(subject).issuer_name(ca_cert.subject)
                    .public_key(k.public_key()).serial_number(x509.random_serial_number())
                    .not_valid_before(now - datetime.timedelta(minutes=5)).not_valid_after(now + datetime.timedelta(days=365))
                    .add_extension(x509.BasicConstraints(ca=False, path_length=None), critical=True)
                    .add_extension(x509.KeyUsage(True, False, False, False, False, False, False, False, False), critical=True)
                    .add_extension(x509.SubjectKeyIdentifier.from_public_key(k.public_key()), critical=False)
                    .add_extension(x509.AuthorityKeyIdentifier.from_issuer_public_key(ca_key.public_key()), critical=False)
                    .add_extension(x509.UnrecognizedExtension(ObjectIdentifier("1.2.3.4.5.6.7.8.1"), attrs), critical=False)
                    .sign(ca_key, hashes.SHA256()))
            key_path.write_bytes(k.private_bytes(serialization.Encoding.PEM, serialization.PrivateFormat.PKCS8, serialization.NoEncryption()))
            cert_path.write_bytes(cert.public_bytes(serialization.Encoding.PEM))
            for f in (base / "users" / f"User1@{org}.agrochain.test" / "msp/cacerts").iterdir():
                shutil.copyfile(f, dest / "cacerts" / f.name)
            shutil.copyfile(base / "users" / f"User1@{org}.agrochain.test" / "msp/config.yaml", dest / "config.yaml")
    sources = RUNTIME / "source-keys"
    sources.mkdir(exist_ok=True)
    for source in ["CKS", "EFATURA", "HKS", "UETDS"]:
        f = sources / (source + ".pem")
        if not f.exists():
            k = ed25519.Ed25519PrivateKey.generate()
            f.write_bytes(k.private_bytes(serialization.Encoding.PEM, serialization.PrivateFormat.PKCS8, serialization.NoEncryption()))
    print("PASS development role identities and four local fixture signing keys ready; no credentials printed")


if __name__ == "__main__":
    prepare()
