#!/usr/bin/env bash
set -Eeuo pipefail
source "$(dirname -- "${BASH_SOURCE[0]}")/common.sh"
require_docker
if [[ -f "$NETWORK_DIR/runtime/chaincode/service.env" ]]; then cc_compose down --timeout 10; fi
compose down --timeout 15
note 'Stopped; ledger volumes and generated identities preserved'
