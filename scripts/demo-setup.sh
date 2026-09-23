#!/usr/bin/env bash
set -Eeuo pipefail
cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.."
if [[ -e network/channel-artifacts/generated-manifest.json ]]; then
  echo 'Existing network detected. Use make network-up chaincode-check backend-build, then make backend-run.' >&2
  exit 1
fi
make backend-prerequisites chaincode-prerequisites chaincode-deps
make bootstrap chaincode-deploy backend-prepare
make demo-ready
# Registers/checks immutable source keys and runs the existing real privacy gate.
make privacy-integration backend-build
echo 'PASS setup. Terminal 1: make backend-run. Terminal 2: make demo-seed demo-check.'
