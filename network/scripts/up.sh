#!/usr/bin/env bash
set -Eeuo pipefail
source "$(dirname -- "${BASH_SOURCE[0]}")/common.sh"
bash "$NETWORK_DIR/scripts/prerequisites.sh"
require_generated
compose up -d --wait --wait-timeout "$WAIT_SECONDS"
wait_healthy
note 'All seven nodes are healthy; run make channel-create'
