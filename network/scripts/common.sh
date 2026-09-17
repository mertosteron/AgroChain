#!/usr/bin/env bash
set -Eeuo pipefail

NETWORK_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd -P)"
REPO_ROOT="$(cd -- "$NETWORK_DIR/.." && pwd -P)"
# This version-controlled file is the single source of version/topology pins.
source "$NETWORK_DIR/.env.example"
export FABRIC_VERSION COMPOSE_PROJECT_NAME CHANNEL_NAME
export PATH="${FABRIC_BIN_DIR:-$NETWORK_DIR/tools/bin}:$PATH"
export FABRIC_CFG_PATH="${FABRIC_CLI_CONFIG_DIR:-$NETWORK_DIR/tools/config}"
export FABRIC_LOGGING_SPEC=ERROR
readonly NETWORK_DIR REPO_ROOT
ORGS=(producer logistics retailer regulator)
SERVICES=(orderer1 orderer2 orderer3 producer logistics retailer regulator)
WAIT_SECONDS="${AGROCHAIN_WAIT_SECONDS:-120}"
[[ "$WAIT_SECONDS" =~ ^[1-9][0-9]{0,3}$ ]] || { echo 'ERROR: invalid AGROCHAIN_WAIT_SECONDS' >&2; exit 1; }
trap 'echo "ERROR: ${BASH_SOURCE[0]##*/}:$LINENO failed; see network/README.md" >&2' ERR

die() { echo "ERROR: $*" >&2; exit 1; }
note() { echo "[agrochain] $*"; }
compose() {
  docker compose --project-name agrochain --env-file "$NETWORK_DIR/.env.example" \
    -f "$NETWORK_DIR/compose/compose-agrochain.yaml" "$@"
}
helper() { python3 "$NETWORK_DIR/scripts/network.py" "$@"; }
need() { command -v "$1" >/dev/null || die "Missing $1; see network/README.md prerequisites"; }
require_docker() {
  need docker
  docker info >/dev/null 2>&1 || die 'Docker daemon unavailable or permission denied; check docker info and network/README.md'
  docker compose version >/dev/null 2>&1 || die 'Docker Compose plugin is required'
}
require_binaries() {
  local binary flag output missing=0
  for binary in peer cryptogen configtxgen configtxlator osnadmin; do
    if ! command -v "$binary" >/dev/null; then
      echo "ERROR: missing Fabric binary: $binary" >&2; missing=1; continue
    fi
    if [[ "$binary" == osnadmin ]]; then
      # osnadmin has no version command in 2.5; check official archive binary.
      [[ "$(sha256sum "$(command -v osnadmin)" | cut -d ' ' -f 1)" == "$OSNADMIN_LINUX_AMD64_SHA256" ]] || die 'osnadmin checksum differs from pinned Linux amd64 release'
      continue
    fi
    flag=version
    [[ "$binary" != configtxgen && "$binary" != osnadmin ]] || flag=--version
    output="$("$binary" "$flag")"
    [[ "$(awk '/Version:/ {sub(/^v/, "", $2); print $2; exit}' <<<"$output")" == "$FABRIC_VERSION" ]] || die "$binary must be version $FABRIC_VERSION"
  done
  (( missing == 0 )) || die 'Install the checksum-pinned Fabric CLI archive as documented in network/README.md'
  [[ -f "$FABRIC_CFG_PATH/core.yaml" ]] || die 'Missing Fabric core.yaml; set FABRIC_CLI_CONFIG_DIR to the release config directory'
}
require_generated() { helper validate-generated; }
peer_env() {
  local org="$1" identity="${2:-Admin}" index
  case "$org" in producer) index=0;; logistics) index=1;; retailer) index=2;; regulator) index=3;; *) die 'Unknown organization';; esac
  export CORE_PEER_LOCALMSPID="${org^}MSP"
  export CORE_PEER_MSPCONFIGPATH="$NETWORK_DIR/organizations/peerOrganizations/$org.agrochain.test/users/$identity@$org.agrochain.test/msp"
  export CORE_PEER_ADDRESS="localhost:$((7051 + index * 1000))"
  export CORE_PEER_TLS_ENABLED=true
  export CORE_PEER_TLS_ROOTCERT_FILE="$NETWORK_DIR/organizations/peerOrganizations/$org.agrochain.test/peers/peer0.$org.agrochain.test/tls/ca.crt"
  export CORE_PEER_TLS_SERVERHOSTOVERRIDE="peer0.$org.agrochain.test"
  export CORE_PEER_CLIENT_CONNTIMEOUT=5s
}
orderer_admin() {
  local node="$1"; shift
  local base="$NETWORK_DIR/organizations/ordererOrganizations/orderer.agrochain.test"
  osnadmin channel "$@" --no-status -o "localhost:$((6053 + node * 1000))" \
    --ca-file "$base/orderers/orderer$node.orderer.agrochain.test/tls/ca.crt" \
    --client-cert "$base/users/Admin@orderer.agrochain.test/tls/client.crt" \
    --client-key "$base/users/Admin@orderer.agrochain.test/tls/client.key"
}
wait_for() {
  local description="$1"; shift
  local deadline=$((SECONDS + WAIT_SECONDS))
  until "$@"; do
    (( SECONDS < deadline )) || die "Timed out waiting for $description"
    sleep 1
  done
}
healthy() {
  local status
  status="$(docker inspect --format '{{.State.Status}} {{if .State.Health}}{{.State.Health.Status}}{{end}}' "agrochain-$1" 2>/dev/null)" || return 1
  [[ "$status" == 'running healthy' ]]
}
wait_healthy() { local service; for service in "${SERVICES[@]}"; do wait_for "$service health" healthy "$service"; done; }
channel_ready() { peer channel getinfo -c "$CHANNEL_NAME" >/dev/null 2>&1; }
orderer_ready() { orderer_admin "$1" list --channelID "$CHANNEL_NAME" 2>/dev/null | jq -e '.status == "active" and .consensusRelation == "consenter" and .height >= 1' >/dev/null; }
