#!/usr/bin/env bash
set -Eeuo pipefail
source "$(dirname -- "${BASH_SOURCE[0]}")/chaincode-toolchain.sh"
requested_version="${CHAINCODE_VERSION:-}"
source "$NETWORK_DIR/config/chaincode.env"
CHAINCODE_VERSION="${requested_version:-$CHAINCODE_VERSION}"
[[ "$CHAINCODE_VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]] || die 'Invalid release version'
[[ -x "$GO_ROOT/bin/go" ]] || die 'Run make chaincode-prerequisites first'
cd "$REPO_ROOT/chaincode/agrochain"
mkdir -p build
case "${1:-test}" in
  deps) go mod tidy; go mod verify ;;
  format) gofmt -w cmd internal ;;
  test|build)
    [[ -z "$(gofmt -l cmd internal)" ]] || die 'Run make chaincode-format'
    go mod verify
    go vet ./...
    go test -race -count=1 -coverprofile=build/coverage.out ./...
    go tool cover -func=build/coverage.out
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=readonly -trimpath -buildvcs=false -ldflags="-s -w -buildid= -X agrochain/chaincode/internal/agrochain.ReleaseVersion=$CHAINCODE_VERSION" -o build/agrochain ./cmd/server
    ;;
  *) die 'Unknown build command' ;;
esac
