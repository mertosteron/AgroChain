#!/usr/bin/env python3
"""Repository-local, checksum-pinned Linux x64 build tools; no host installation."""
from pathlib import Path
import hashlib
import tarfile
import urllib.request

ROOT = Path(__file__).resolve().parents[2]
TOOLS = ROOT / "network/tools"
ARTIFACTS = [
    ("apache-maven-3.9.11", "https://repo.maven.apache.org/maven2/org/apache/maven/apache-maven/3.9.11/apache-maven-3.9.11-bin.tar.gz", "sha512", "bcfe4fe305c962ace56ac7b5fc7a08b87d5abd8b7e89027ab251069faebee516b0ded8961445d6d91ec1985dfe30f8153268843c89aa392733d1a3ec956c9978"),
    ("jdk-21.0.12.1+1", "https://github.com/adoptium/temurin21-binaries/releases/download/jdk-21.0.12.1%2B1/OpenJDK21U-jdk_x64_linux_hotspot_21.0.12.1_1.tar.gz", "sha256", "ce79869e1307ed8ee1e2baa86a412b1eb5b75d10a01006d788a6f968bcfaee94"),
]
TOOLS.mkdir(parents=True, exist_ok=True)
for directory, url, algorithm, digest in ARTIFACTS:
    if (TOOLS / directory / "bin").is_dir():
        continue
    archive = TOOLS / (directory + ".tar.gz")
    if not archive.exists():
        print("Downloading", directory, flush=True)
        with urllib.request.urlopen(url, timeout=120) as response, archive.open("wb") as target:
            while chunk := response.read(1024 * 1024):
                target.write(chunk)
    if hashlib.new(algorithm, archive.read_bytes()).hexdigest() != digest:
        raise SystemExit("Checksum mismatch: " + archive.name)
    with tarfile.open(archive) as source:
        source.extractall(TOOLS, filter="data")
    assert (TOOLS / directory / "bin").is_dir()
print("PASS pinned Maven 3.9.11 and Temurin 21.0.12.1+1")
