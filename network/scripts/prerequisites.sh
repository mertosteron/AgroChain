#!/usr/bin/env bash
set -Eeuo pipefail
source "$(dirname -- "${BASH_SOURCE[0]}")/common.sh"
for tool in bash python3 jq openssl timeout sha256sum make awk grep curl tar; do need "$tool"; done
(( BASH_VERSINFO[0] >= 4 )) || die 'Bash >=4 required'
python3 -c 'import sys; sys.exit(0 if sys.version_info >= (3, 10) else "Python >=3.10 required")'
[[ "$(uname -s)" == Linux && "$(uname -m)" == x86_64 ]] || die 'This tested/pinned installation targets Linux x86-64'
require_docker
helper check-docker
require_binaries
compose config --quiet
helper check-ports
note "PASS prerequisites: Fabric $FABRIC_VERSION, Docker/Compose, configuration and ports"
