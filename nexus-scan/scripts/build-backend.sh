#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
BIN_DIR="$ROOT_DIR/bin"

log()  { echo -e "\033[0;32m[+]\033[0m $*"; }
die()  { echo -e "\033[0;31m[x]\033[0m $*"; exit 1; }

cd "$ROOT_DIR"

export PATH=$PATH:/usr/local/go/bin
command -v go &>/dev/null || die "Go not found. Run: sudo ./scripts/install-deps.sh"

log "Downloading Go module dependencies..."
go mod download || die "go mod download failed"

log "Verifying Go modules..."
go mod verify || die "go mod verify failed"

log "Running tests before build..."
go test ./tests/... -v -timeout 60s || die "Tests failed — aborting build"

log "Building nexus-scan backend..."
mkdir -p "$BIN_DIR"

CGO_ENABLED=1 go build \
    -ldflags="-s -w -X main.Version=1.0.0 -X main.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    -o "$BIN_DIR/nexus-scan-backend" \
    ./cmd/nexus-scan/

log "Build successful: $BIN_DIR/nexus-scan-backend"
log "Size: $(du -sh "$BIN_DIR/nexus-scan-backend" | cut -f1)"
