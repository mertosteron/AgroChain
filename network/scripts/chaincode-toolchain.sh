#!/usr/bin/env bash
set -Eeuo pipefail
source "$(dirname -- "${BASH_SOURCE[0]}")/common.sh"
GO_VERSION=1.27.1
GO_SHA256=63d339f0da5ab53635a56f2490a7984dfe12dfcff22ad749f63edaf590168445
GO_ROOT="$NETWORK_DIR/tools/go-$GO_VERSION"
export PATH="$GO_ROOT/bin:$PATH"
export GOTOOLCHAIN=local
export GOCACHE="$NETWORK_DIR/runtime/go-build"
export GOMODCACHE="$NETWORK_DIR/runtime/go-mod"
if [[ "${1:-}" == install ]]; then
  [[ "$(uname -s)/$(uname -m)" == Linux/x86_64 ]] || die 'Go toolchain supports Linux amd64 only'
  if [[ ! -x "$GO_ROOT/bin/go" ]]; then
    mkdir -p "$NETWORK_DIR/tools" "$GO_ROOT"
    archive="$NETWORK_DIR/tools/go-$GO_VERSION.tar.gz"
    curl --fail --location --retry 3 "https://go.dev/dl/go$GO_VERSION.linux-amd64.tar.gz" -o "$archive"
    echo "$GO_SHA256  $archive" | sha256sum --check --status
    tar -xzf "$archive" --strip-components=1 -C "$GO_ROOT"
  fi
  [[ "$(go env GOVERSION)" == "go$GO_VERSION" ]] || die 'Unexpected Go toolchain version'
  go version
fi
