#!/usr/bin/env python3
"""Create a portable source + presentation archive; never include local identities."""
import gzip
import hashlib
import io
import json
import re
from pathlib import Path, PurePosixPath
import subprocess
import tarfile

ROOT = Path(__file__).resolve().parents[1]
ALLOWED = {".github", "backend", "chaincode", "docs", "network", "scripts"}
ROOT_FILES = {".gitignore", "AGENTS.md", "ARCHITECTURE.md", "Makefile", "README.md"}
EXCLUDED = ("network/runtime/", "network/tools/", "network/organizations/", "network/channel-artifacts/", "network/connection-profiles/generated/", "backend/target/", "chaincode/agrochain/build/")


def contains_private_key(data):
    # A PEM block begins on its own line; source code can legitimately mention
    # a marker while parsing or rejecting it, without embedding a private key.
    return re.search(rb"(?m)^[ \t]*-----BEGIN (?:[A-Z]+ )?PRIVATE KEY-----\r?\n", data) is not None


def permitted(name):
    p = PurePosixPath(name)
    if p.is_absolute() or ".." in p.parts:
        return False
    if name not in ROOT_FILES and (not p.parts or p.parts[0] not in ALLOWED):
        return False
    if any(name.startswith(prefix) for prefix in EXCLUDED):
        return name.endswith("/.gitkeep")
    return not (p.suffix.lower() in {".pem", ".key", ".sqlite", ".p12", ".pyc"} or p.name in {".env", "tokens.json"} or p.name.endswith("_sk") or "__pycache__" in p.parts)


def main():
    names = subprocess.check_output(["git", "ls-files", "--cached", "--others", "--exclude-standard", "-z"], cwd=ROOT).decode().split("\0")
    names = sorted(set(names + ["AGENTS.md"]))
    entries = {}
    for name in names:
        src = ROOT / name
        if not name or not permitted(name) or not src.is_file():
            continue
        if src.is_symlink():
            raise RuntimeError("Source symlinks are not allowed in the deliverable")
        data = src.read_bytes()
        if contains_private_key(data):
            raise RuntimeError("Private key marker in selected source")
        entries[name] = data
    for required in ["docs/competition/AgroChain.pptx", "docs/competition/offline.html", "docs/STAGE8_REPORT.md", "scripts/demo.py"]:
        if required not in entries:
            raise RuntimeError("Missing deliverable: " + required)
    manifest = {"schemaVersion": "agrochain.competition-package.v1", "sourceMode": "SIMULATED", "files": {n: hashlib.sha256(b).hexdigest() for n, b in entries.items()}}
    dest = ROOT / "dist"
    dest.mkdir(exist_ok=True)
    archive = dest / "AgroChain-competition.tar.gz"
    with archive.open("wb") as file, gzip.GzipFile(filename="", mode="wb", fileobj=file, mtime=0) as zipped, tarfile.open(fileobj=zipped, mode="w") as tar:
        for name, data in {**entries, "PACKAGE_MANIFEST.json": (json.dumps(manifest, indent=2) + "\n").encode()}.items():
            info = tarfile.TarInfo("AgroChain/" + name)
            info.size = len(data)
            info.mode = 0o755 if name in entries and (ROOT / name).stat().st_mode & 0o111 else 0o644
            tar.addfile(info, io.BytesIO(data))
    (dest / "AgroChain-competition.sha256").write_text(hashlib.sha256(archive.read_bytes()).hexdigest() + "  AgroChain-competition.tar.gz\n")
    print("PASS source/presentation package with", len(entries), "files; credentials and caches excluded:", archive)


if __name__ == "__main__":
    main()
