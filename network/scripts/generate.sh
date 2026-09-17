#!/usr/bin/env bash
set -Eeuo pipefail
source "$(dirname -- "${BASH_SOURCE[0]}")/common.sh"
need python3
require_binaries
umask 077
if [[ -f "$NETWORK_DIR/channel-artifacts/generated-manifest.json" ]]; then
  require_generated
  note 'Existing generated artifacts validated; identities unchanged'
  exit 0
fi
helper require-empty
cryptogen generate --config="$NETWORK_DIR/config/crypto-config.yaml" --output="$NETWORK_DIR/organizations"
FABRIC_CFG_PATH="$NETWORK_DIR/config" configtxgen -profile AgroChannel -channelID "$CHANNEL_NAME" \
  -outputBlock "$NETWORK_DIR/channel-artifacts/$CHANNEL_NAME.block"
helper profiles
helper record-generated
require_generated
note 'Development identities, channel genesis and four connection profiles generated'
