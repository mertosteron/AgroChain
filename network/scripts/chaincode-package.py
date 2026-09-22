#!/usr/bin/env python3
"""Reproducible CCAAS archives. Only public TLS roots enter lifecycle packages."""
import gzip
import hashlib
import io
import json
import os
from pathlib import Path
import subprocess
import sys
import tarfile

ROOT = Path(__file__).resolve().parents[2]
OUTPUT = ROOT / "network/runtime/chaincode"
ORGS = ("producer", "logistics", "retailer", "regulator")


def archive(files):
    raw = io.BytesIO()
    with tarfile.open(fileobj=raw, mode="w", format=tarfile.USTAR_FORMAT) as tar:
        for name, value in sorted(files.items()):
            item = tarfile.TarInfo(name)
            item.size, item.mode, item.mtime = len(value), 0o644, 0
            tar.addfile(item, io.BytesIO(value))
    return gzip.compress(raw.getvalue(), mtime=0)


def encoded(value):
    return json.dumps(value, sort_keys=True, separators=(",", ":")).encode()


def package(version):
    OUTPUT.mkdir(parents=True, exist_ok=True)
    binary = ROOT / "chaincode/agrochain/build/agrochain"
    digest = hashlib.sha256(binary.read_bytes()).hexdigest()
    manifest = {"version": version, "binarySha256": digest, "packages": {}}
    for org in ORGS:
        directory = OUTPUT / org
        directory.mkdir(exist_ok=True)
        tls = directory / "tls"
        tls.mkdir(exist_ok=True)
        cert, private = tls / "server.crt", tls / "server.key"
        hostname = f"agrochain-cc-{org}"
        if cert.exists() != private.exists():
            raise RuntimeError("Incomplete chaincode TLS pair; restore it before packaging")
        if not cert.exists():
            subprocess.run(["openssl", "req", "-x509", "-newkey", "ec", "-pkeyopt",
                            "ec_paramgen_curve:P-256", "-nodes", "-days", "365",
                            "-subj", f"/CN={hostname}", "-addext", f"subjectAltName=DNS:{hostname}",
                            "-keyout", str(private), "-out", str(cert)], check=True,
                           stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
            private.chmod(0o600)
        subprocess.run(["openssl", "x509", "-in", str(cert), "-noout", "-checkend", "86400"],
                       check=True, stdout=subprocess.DEVNULL)
        connection = {"address": f"{hostname}:9999", "dial_timeout": "10s",
                      "tls_required": True, "client_auth_required": False,
                      "root_cert": cert.read_text()}
        label = f"agrochain_{version}_{org}_{digest[:16]}"
        package_bytes = archive({"metadata.json": encoded({"type": "ccaas", "label": label}),
                                 "code.tar.gz": archive({"connection.json": encoded(connection)})})
        path = directory / "chaincode.tar.gz"
        path.write_bytes(package_bytes)
        manifest["packages"][org] = {"label": label, "sha256": hashlib.sha256(package_bytes).hexdigest()}
    (OUTPUT / "manifest.json").write_bytes(encoded(manifest) + b"\n")
    print(f"Packaged four TLS CCAAS endpoints; binary SHA-256 {digest}")


if __name__ == "__main__":
    package(sys.argv[1])
