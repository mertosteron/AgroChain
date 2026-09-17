#!/usr/bin/env bash
set -Eeuo pipefail
source "$(dirname -- "${BASH_SOURCE[0]}")/common.sh"
require_docker
# Validate every filesystem and Docker target before deleting anything.
helper cleanup-preflight
helper clean-generated
note 'Removed only named AgroChain demo volumes and generated artifacts; tools and source preserved'
