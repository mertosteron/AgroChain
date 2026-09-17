#!/usr/bin/env bash
set -Eeuo pipefail
source "$(dirname -- "${BASH_SOURCE[0]}")/common.sh"
require_docker
require_binaries
require_generated
wait_healthy
helper verify-nodes
helper verify-tls
umask 077
mkdir -p "$NETWORK_DIR/runtime"
for node in 1 2 3; do
  wait_for "orderer$node active consenter" orderer_ready "$node"
done
for org in "${ORGS[@]}"; do
  peer_env "$org"
  peer channel list | grep -Fxq "$CHANNEL_NAME"
  wait_for "$org channel" channel_ready
  peer channel getinfo -c "$CHANNEL_NAME"
  # Fetch the actual peer ledger configuration using this organization's identity.
  peer channel fetch config "$NETWORK_DIR/runtime/$org-config.block" -c "$CHANNEL_NAME"
  configtxlator proto_decode --input "$NETWORK_DIR/runtime/$org-config.block" --type common.Block \
    --output "$NETWORK_DIR/runtime/$org-config.json"
  helper verify-config "$NETWORK_DIR/runtime/$org-config.json"
  peer lifecycle chaincode queryinstalled --output json | jq -e '(.installed_chaincodes // []) | length == 0' >/dev/null
  peer lifecycle chaincode querycommitted -C "$CHANNEL_NAME" --output json | jq -e '(.chaincode_definitions // []) | length == 0' >/dev/null
  # A normal client identity can read channel information too; no business role claim.
  peer_env "$org" User1
  peer channel getinfo -c "$CHANNEL_NAME"
  note "PASS $org admin/client channel queries; no chaincode installed or committed"
done
# Restart every peer in turn to test actual persistence/reconnection, not just status.
for org in "${ORGS[@]}"; do
  peer_env "$org"
  before="$(peer channel getinfo -c "$CHANNEL_NAME" 2>/dev/null)"
  compose restart "$org"
  wait_for "$org restart" healthy "$org"
  wait_for "$org channel after restart" channel_ready
  after="$(peer channel getinfo -c "$CHANNEL_NAME" 2>/dev/null)"
  [[ "$before" == "$after" ]] || die "$org channel state changed unexpectedly across restart"
done
helper verify-nodes
note 'PASS Stage 2 live verification; use make smoke for the two-bootstrap shutdown test'
