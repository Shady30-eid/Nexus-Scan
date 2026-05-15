#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
FRONTEND_DIR="$ROOT_DIR/frontend"

log()  { echo -e "\033[0;32m[+]\033[0m $*"; }
warn() { echo -e "\033[1;33m[!]\033[0m $*"; }
die()  { echo -e "\033[0;31m[x]\033[0m $*"; exit 1; }

export PATH=$PATH:/usr/local/go/bin
source ~/.cargo/env 2>/dev/null || true

command -v go   &>/dev/null || die "Go not found. Run: sudo ./scripts/install-deps.sh"
command -v pnpm &>/dev/null || die "pnpm not found. Run: sudo ./scripts/install-deps.sh"

log "Starting adb daemon..."
adb start-server 2>/dev/null || warn "adb not available"

log "Starting Nexus-Scan backend in dev mode..."
cd "$ROOT_DIR"
CGO_ENABLED=1 go run ./cmd/nexus-scan/ &
BACKEND_PID=$!
log "Backend started (PID: $BACKEND_PID)"

sleep 1

if ! kill -0 "$BACKEND_PID" 2>/dev/null; then
    die "Backend crashed at startup. Check logs."
fi

cleanup() {
    log "Stopping backend (PID: $BACKEND_PID)..."
    kill -SIGTERM "$BACKEND_PID" 2>/dev/null || true
    wait "$BACKEND_PID" 2>/dev/null || true
}
trap cleanup EXIT INT TERM

log "Starting Tauri dev mode..."
cd "$FRONTEND_DIR"
pnpm install --frozen-lockfile 2>/dev/null || pnpm install
pnpm tauri dev
