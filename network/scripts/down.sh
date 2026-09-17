#!/usr/bin/env bash
set -Eeuo pipefail
source "$(dirname -- "${BASH_SOURCE[0]}")/common.sh"
require_docker
compose down --timeout 15
note 'Stopped; ledger volumes and generated identities preserved'
