#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
BIN_DIR="$ROOT_DIR/bin"
BACKEND_BIN="$BIN_DIR/nexus-scan-backend"

log()  { echo -e "\033[0;32m[+]\033[0m $*"; }
warn() { echo -e "\033[1;33m[!]\033[0m $*"; }
die()  { echo -e "\033[0;31m[x]\033[0m $*"; exit 1; }

cd "$ROOT_DIR"

if [[ ! -f "$BACKEND_BIN" ]]; then
    log "Backend binary not found — building now..."
    ./scripts/build-backend.sh || die "Backend build failed"
fi

log "Starting adb server..."
if command -v adb &>/dev/null; then
    adb start-server 2>/dev/null || warn "adb start-server failed (continuing)"
else
    warn "adb not found — Android device detection will be unavailable"
fi

log "Starting usbmuxd for iOS support..."
if command -v usbmuxd &>/dev/null; then
    systemctl start usbmuxd 2>/dev/null || usbmuxd 2>/dev/null || warn "usbmuxd start failed (continuing)"
else
    warn "usbmuxd not found — iOS device detection will be unavailable"
fi

log "Starting Nexus-Scan backend on ws://127.0.0.1:9999..."
"$BACKEND_BIN" &
BACKEND_PID=$!

sleep 0.5

if ! kill -0 "$BACKEND_PID" 2>/dev/null; then
    die "Backend failed to start. Check nexus-scan.log for details."
fi

log "Backend running (PID: $BACKEND_PID)"

cleanup() {
    log "Shutting down Nexus-Scan backend (PID: $BACKEND_PID)..."
    kill -SIGTERM "$BACKEND_PID" 2>/dev/null || true
    wait "$BACKEND_PID" 2>/dev/null || true
    log "Backend stopped."
}
trap cleanup EXIT INT TERM

if [[ -f "$ROOT_DIR/nexus-scan.AppImage" ]]; then
    log "Launching Nexus-Scan AppImage..."
    "$ROOT_DIR/nexus-scan.AppImage" &
elif command -v nexus-scan &>/dev/null; then
    log "Launching Nexus-Scan (system-installed)..."
    nexus-scan &
else
    log ""
    warn "Tauri frontend not found. Options:"
    warn "  1. Build AppImage:  ./scripts/package-appimage.sh"
    warn "  2. Build .deb:      ./scripts/package-deb.sh"
    warn "  3. Dev mode:        cd frontend && pnpm tauri dev"
    warn ""
    warn "Backend is running at ws://127.0.0.1:9999 — waiting for frontend connection."
fi

log "Nexus-Scan is running. Press Ctrl+C to stop."
wait "$BACKEND_PID"
