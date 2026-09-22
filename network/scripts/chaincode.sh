#!/usr/bin/env bash
set -Eeuo pipefail
source "$(dirname -- "${BASH_SOURCE[0]}")/common.sh"
requested_version="${CHAINCODE_VERSION:-}"
requested_sequence="${CHAINCODE_SEQUENCE:-}"
source "$NETWORK_DIR/config/chaincode.env"
CHAINCODE_VERSION="${requested_version:-$CHAINCODE_VERSION}"
CHAINCODE_SEQUENCE="${requested_sequence:-$CHAINCODE_SEQUENCE}"
[[ "$CHAINCODE_VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ && "$CHAINCODE_SEQUENCE" =~ ^[1-9][0-9]*$ ]] || die 'Explicit semantic version and positive sequence required'
CC_DIR="$NETWORK_DIR/runtime/chaincode"
require_docker
require_binaries
require_generated
if [[ -z "$requested_sequence" ]]; then
  peer_env regulator
  existing="$(peer lifecycle chaincode querycommitted -C "$CHANNEL_NAME" --output json)"
  matching_sequence="$(jq -r --arg n "$CHAINCODE_NAME" --arg v "$CHAINCODE_VERSION" '[.chaincode_definitions[]? | select(.name == $n and .version == $v) | .sequence][0] // 0' <<<"$existing")"
  if [[ "$matching_sequence" -gt 0 ]]; then CHAINCODE_SEQUENCE="$matching_sequence"; fi
fi
orderer=(--orderer localhost:7050 --tls --cafile "$NETWORK_DIR/organizations/ordererOrganizations/orderer.agrochain.test/orderers/orderer1.orderer.agrochain.test/tls/ca.crt" --ordererTLSHostnameOverride orderer1.orderer.agrochain.test)
definition=(--channelID "$CHANNEL_NAME" --name "$CHAINCODE_NAME" --version "$CHAINCODE_VERSION" --sequence "$CHAINCODE_SEQUENCE" --signature-policy "$CHAINCODE_POLICY" --collections-config "$NETWORK_DIR/config/collections.json")
all_peers=()
for org in "${ORGS[@]}"; do
  peer_env "$org"
  all_peers+=(--peerAddresses "$CORE_PEER_ADDRESS" --tlsRootCertFiles "$CORE_PEER_TLS_ROOTCERT_FILE")
done

package_cc() {
  export CHAINCODE_VERSION
  bash "$NETWORK_DIR/scripts/chaincode-build.sh" build
  python3 "$NETWORK_DIR/scripts/chaincode-package.py" "$CHAINCODE_VERSION"
  local digest image org package_id
  digest="$(jq -r .binarySha256 "$CC_DIR/manifest.json")"
  image="agrochain-chaincode:$CHAINCODE_VERSION-${digest:0:16}"
  docker build --network=none --tag "$image" "$REPO_ROOT/chaincode/agrochain"
  # Never hard-code package IDs: the pinned peer CLI calculates their exact bytes.
  {
    echo "CHAINCODE_IMAGE=$image"
    echo "CHAINCODE_UID=$(id -u)"
    echo "CHAINCODE_GID=$(id -g)"
    for org in "${ORGS[@]}"; do
      package_id="$(peer lifecycle chaincode calculatepackageid "$CC_DIR/$org/chaincode.tar.gz")"
      echo "${org^^}_PACKAGE_ID=$package_id"
    done
  } > "$CC_DIR/candidate.env"
}
install_cc() {
  [[ -f "$CC_DIR/candidate.env" ]] || die 'Run make chaincode-package'
  local org package_id
  for org in "${ORGS[@]}"; do
    peer_env "$org"
    package_id="$(peer lifecycle chaincode calculatepackageid "$CC_DIR/$org/chaincode.tar.gz")"
    local committed approved
    committed="$(peer lifecycle chaincode querycommitted -C "$CHANNEL_NAME" --output json)"
    if jq -e --arg n "$CHAINCODE_NAME" --argjson s "$CHAINCODE_SEQUENCE" 'any(.chaincode_definitions[]?; .name == $n and .sequence == $s)' <<<"$committed" >/dev/null; then
      approved="$(peer lifecycle chaincode queryapproved -C "$CHANNEL_NAME" -n "$CHAINCODE_NAME" --sequence "$CHAINCODE_SEQUENCE" --output json)"
      jq -e --arg id "$package_id" '.source.Type.LocalPackage.package_id == $id' <<<"$approved" >/dev/null || die 'Release bytes differ from committed approval; use a new explicit version and sequence'
    fi
    if ! peer lifecycle chaincode queryinstalled --output json | jq -e --arg id "$package_id" 'any(.installed_chaincodes[]?; .package_id == $id)' >/dev/null; then
      peer lifecycle chaincode install "$CC_DIR/$org/chaincode.tar.gz"
    fi
    peer lifecycle chaincode queryinstalled --output json | jq -e --arg id "$package_id" 'any(.installed_chaincodes[]?; .package_id == $id)' >/dev/null
  done
  cp "$CC_DIR/candidate.env" "$CC_DIR/service.env"
  cc_compose up -d
}
approve_cc() {
  local org package_id
  for org in "${ORGS[@]}"; do
    peer_env "$org"
    package_id="$(peer lifecycle chaincode calculatepackageid "$CC_DIR/$org/chaincode.tar.gz")"
    peer lifecycle chaincode approveformyorg "${orderer[@]}" "${definition[@]}" --package-id "$package_id" --waitForEvent --waitForEventTimeout 30s
  done
}
readiness() {
  peer_env regulator
  peer lifecycle chaincode checkcommitreadiness "${definition[@]}" --output json | jq -e '.approvals == {"ProducerMSP":true,"LogisticsMSP":true,"RetailerMSP":true,"RegulatorMSP":true}' >/dev/null
}
commit_cc() {
  readiness
  peer_env regulator
  # Each peer certificate includes localhost. One organization-specific override
  # must not be reused across a multi-peer TLS request.
  export CORE_PEER_TLS_SERVERHOSTOVERRIDE=localhost
  peer lifecycle chaincode commit "${orderer[@]}" "${definition[@]}" "${all_peers[@]}" --waitForEvent --waitForEventTimeout 30s
}
check_cc() {
  local org
  for org in "${ORGS[@]}"; do
    peer_env "$org"
    peer lifecycle chaincode querycommitted -C "$CHANNEL_NAME" -n "$CHAINCODE_NAME" --output json | jq -e --arg v "$CHAINCODE_VERSION" --argjson s "$CHAINCODE_SEQUENCE" '.version == $v and .sequence == $s and .approvals == {"ProducerMSP":true,"LogisticsMSP":true,"RetailerMSP":true,"RegulatorMSP":true}' >/dev/null
    peer_env "$org" User1
    peer chaincode query -C "$CHANNEL_NAME" -n "$CHAINCODE_NAME" -c '{"Args":["Health"]}' | jq -e --arg v "$CHAINCODE_VERSION" '.schemaVersion == "agrochain.health.v1" and .version == $v and (.writesEnabled | type) == "boolean" and .evidenceVerification == "REQUIRED_STAGE_4"' >/dev/null
    note "PASS $org committed definition and evidence-required health"
  done
}
case "${1:-}" in
  package) package_cc ;;
  install) install_cc ;;
  approve) approve_cc ;;
  commit) commit_cc ;;
  check) check_cc ;;
  deploy|upgrade)
    peer_env regulator
    committed="$(peer lifecycle chaincode querycommitted -C "$CHANNEL_NAME" --output json)"
    current="$(jq -r --arg n "$CHAINCODE_NAME" '[.chaincode_definitions[]? | select(.name == $n) | .sequence][0] // 0' <<<"$committed")"
    if [[ "${1:-}" == upgrade ]]; then
      [[ -n "$requested_version" && -n "$requested_sequence" && "$CHAINCODE_SEQUENCE" -eq $((current+1)) ]] || die 'Upgrade requires explicit version and next sequence'
      old_version="$(jq -r --arg n "$CHAINCODE_NAME" '[.chaincode_definitions[]? | select(.name == $n) | .version][0] // ""' <<<"$committed")"
      [[ "$old_version" != "$CHAINCODE_VERSION" ]] || die 'Upgrade requires a new semantic version'
    elif [[ "$current" -ne 0 && "$current" -ne "$CHAINCODE_SEQUENCE" ]]; then
      die 'Use explicit chaincode-upgrade for a new sequence'
    fi
    package_cc
    install_cc
    if [[ "$current" -lt "$CHAINCODE_SEQUENCE" ]]; then approve_cc; commit_cc; fi
    check_cc
    ;;
  *) die 'Use package/install/approve/commit/check/deploy/upgrade' ;;
esac
