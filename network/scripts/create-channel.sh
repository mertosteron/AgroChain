#!/usr/bin/env bash
set -Eeuo pipefail
source "$(dirname -- "${BASH_SOURCE[0]}")/common.sh"
require_docker
require_binaries
require_generated
wait_healthy
for node in 1 2 3; do
  channels="$(orderer_admin "$node" list)"
  if ! jq -e --arg channel "$CHANNEL_NAME" 'any(.channels[]?; .name == $channel)' <<<"$channels" >/dev/null; then
    orderer_admin "$node" join --channelID "$CHANNEL_NAME" --config-block "$NETWORK_DIR/channel-artifacts/$CHANNEL_NAME.block"
  fi
done
for node in 1 2 3; do wait_for "orderer$node channel activation" orderer_ready "$node"; done
for org in "${ORGS[@]}"; do
  peer_env "$org"
  channels="$(peer channel list)"
  if ! grep -Fxq "$CHANNEL_NAME" <<<"$channels"; then
    peer channel join -b "$NETWORK_DIR/channel-artifacts/$CHANNEL_NAME.block"
  fi
  wait_for "$org channel ledger" channel_ready
done
note 'agrochannel active on three orderers and joined by four peers'
