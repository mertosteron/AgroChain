#!/usr/bin/env bash
# PDC endorsement needs discovered collection members, not just healthy processes.
set -Eeuo pipefail
source "$(dirname -- "${BASH_SOURCE[0]}")/../network/scripts/common.sh"
require_binaries
need discover
ready() {
  local org="$1" result
  peer_env "$org" User1
  local keys=("$CORE_PEER_MSPCONFIGPATH"/keystore/*)
  local certs=("$CORE_PEER_MSPCONFIGPATH"/signcerts/*)
  [[ ${#keys[@]} == 1 && ${#certs[@]} == 1 ]] || return 1
  result="$(discover --peerTLSCA "$CORE_PEER_TLS_ROOTCERT_FILE" \
    --userKey "${keys[0]}" --userCert "${certs[0]}" --MSP "$CORE_PEER_LOCALMSPID" \
    peers --channel "$CHANNEL_NAME" --server "$CORE_PEER_ADDRESS" 2>/dev/null)" || return 1
  jq -e '([.[].MSPID] | sort) == ["LogisticsMSP","ProducerMSP","RegulatorMSP","RetailerMSP"] and all(.[]; .Chaincodes | index("agrochain") != null)' <<<"$result" >/dev/null
}
for org in "${ORGS[@]}"; do
  wait_for "$org discovery of four installed chaincode peers" ready "$org"
  note "PASS $org sees all four agrochain peers"
done
