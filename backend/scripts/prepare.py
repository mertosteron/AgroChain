#!/usr/bin/env python3
"""Create local demo bearer credentials without displaying them."""
import json
import os
from pathlib import Path
import secrets

root = Path(__file__).resolve().parents[2]
os.umask(0o077)
dest = root / "network/runtime/backend"
dest.mkdir(parents=True, exist_ok=True)
os.chmod(dest, 0o700)
path = dest / "tokens.json"
if not path.exists():
    actors = ["producer:producer", "logistics:carrier", "retailer:retailer", "regulator:auditor", "regulator:reviewer", "regulator:oracle", "regulator:public-reader"]
    path.write_text(json.dumps({a: secrets.token_hex(32) for a in actors}, indent=2) + "\n")
os.chmod(path, 0o600)
print("PASS development API credentials prepared in ignored backend runtime")
