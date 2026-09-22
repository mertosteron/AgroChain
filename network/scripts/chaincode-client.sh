#!/usr/bin/env bash
set -Eeuo pipefail
source "$(dirname -- "${BASH_SOURCE[0]}")/common.sh"
source "$NETWORK_DIR/config/chaincode.env"
[[ $# -ge 2 ]] || die 'Expected organization and query/invoke/info/block'
org="$1"; action="$2"
peer_env "$org" User1
signer="$CORE_PEER_MSPCONFIGPATH"
role="${AGROCHAIN_ROLE:-none}"
if [[ "$role" != none ]]; then
  [[ "$role" =~ ^(producer|carrier|retailer|auditor|reviewer|oracle|public-reader|admin)$ ]] || die 'Invalid business role'
  signer="$NETWORK_DIR/runtime/identities/$org/$role/msp"
  [[ -d "$signer" ]] || die 'Run make privacy-prepare to provision development role identities'
fi
if [[ -n "${AGROCHAIN_QUERY_PEER:-}" ]]; then peer_env "$AGROCHAIN_QUERY_PEER" User1; fi
export CORE_PEER_MSPCONFIGPATH="$signer"
export CORE_PEER_LOCALMSPID="${org^}MSP"
transient_args=()
if [[ -n "${AGROCHAIN_TRANSIENT_FILE:-}" ]]; then transient_args=(--transient "$(cat -- "$AGROCHAIN_TRANSIENT_FILE")"); fi
case "$action" in
  query) peer chaincode query -C "$CHANNEL_NAME" -n "$CHAINCODE_NAME" -c "$3" "${transient_args[@]}" ;;
  invoke|invoke-one)
    peer_targets=()
    for endorser in retailer regulator; do
      [[ "$action" != invoke-one || "$endorser" == regulator ]] || continue
      peer_env "$endorser" User1
      peer_targets+=(--peerAddresses "$CORE_PEER_ADDRESS" --tlsRootCertFiles "$CORE_PEER_TLS_ROOTCERT_FILE")
    done
    peer_env "$org" User1
    export CORE_PEER_MSPCONFIGPATH="$signer"
    export CORE_PEER_TLS_SERVERHOSTOVERRIDE=localhost
    FABRIC_LOGGING_SPEC=INFO peer chaincode invoke -C "$CHANNEL_NAME" -n "$CHAINCODE_NAME" -c "$3" \
      -o localhost:7050 --tls --cafile "$NETWORK_DIR/organizations/ordererOrganizations/orderer.agrochain.test/orderers/orderer1.orderer.agrochain.test/tls/ca.crt" \
      --ordererTLSHostnameOverride orderer1.orderer.agrochain.test "${peer_targets[@]}" "${transient_args[@]}" --waitForEvent --waitForEventTimeout 30s
    ;;
  info) peer channel getinfo -c "$CHANNEL_NAME" ;;
  block)
    [[ "$3" =~ ^[0-9]+$ ]] || die 'Expected numeric block number'
    mkdir -p "$NETWORK_DIR/runtime/chaincode/blocks"
    dest="$NETWORK_DIR/runtime/chaincode/blocks/$org-$3"
    peer channel fetch "$3" "$dest.block" -c "$CHANNEL_NAME"
    configtxlator proto_decode --input "$dest.block" --type common.Block --output "$dest.json"
    cat "$dest.json"
    ;;
  *) die 'Unknown client action' ;;
esac
