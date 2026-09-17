#!/usr/bin/env bash
set -Eeuo pipefail
source "$(dirname -- "${BASH_SOURCE[0]}")/common.sh"
# Stop/recreate containers while keeping credentials and persistent ledgers.
bash "$NETWORK_DIR/scripts/down.sh"
for cycle in 1 2; do
  note "Bootstrap/verify cycle $cycle of 2"
  make -C "$REPO_ROOT" bootstrap
  bash "$NETWORK_DIR/scripts/verify.sh"
  if (( cycle == 1 )); then bash "$NETWORK_DIR/scripts/down.sh"; fi
done
note 'PASS two bootstraps, shutdown/recreation and peer restarts; network left running'
