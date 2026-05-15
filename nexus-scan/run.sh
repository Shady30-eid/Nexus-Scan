#!/usr/bin/env bash
set -euo pipefail

# =====================================================
# Nexus-Scan — One-shot launcher (desktop app, Kali)
# Usage: ./run.sh
# =====================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN="$SCRIPT_DIR/bin/nexus-scan-backend"
LOG="$SCRIPT_DIR/nexus-scan.log"
WS_PORT="${NEXUS_WS_PORT:-9999}"

GREEN="\033[0;32m"; YELLOW="\033[1;33m"; RED="\033[0;31m"; NC="\033[0m"
log()  { echo -e "${GREEN}[+]${NC} $*"; }
warn() { echo -e "${YELLOW}[!]${NC} $*"; }
die()  { echo -e "${RED}[x]${NC} $*"; exit 1; }

export PATH="$PATH:/usr/local/go/bin:$HOME/.cargo/bin"

# ── 1. Build backend if binary is missing ───────────
if [[ ! -f "$BIN" ]]; then
    log "Binary not found — building Go backend..."
    command -v go &>/dev/null || die "Go not installed. Run: sudo ./scripts/install-deps.sh"
    cd "$SCRIPT_DIR"
    go mod download 2>>"$LOG"
    CGO_ENABLED=1 go build \
        -ldflags="-s -w" \
        -o "$BIN" \
        ./cmd/nexus-scan/ \
        2>>"$LOG" || die "Build failed. Check $LOG"
    log "Backend built: $BIN"
fi

# ── 2. Start ADB daemon (Android support) ───────────
if command -v adb &>/dev/null; then
    log "Starting ADB server..."
    adb start-server 2>/dev/null || warn "ADB start failed (Android won't be detected)"
else
    warn "adb not found — install with: sudo apt-get install adb"
fi

# ── 3. Start usbmuxd (iOS support) ──────────────────
if command -v usbmuxd &>/dev/null; then
    systemctl is-active --quiet usbmuxd 2>/dev/null || \
        usbmuxd --user 2>/dev/null || \
        warn "usbmuxd failed to start (iOS detection disabled)"
fi

# ── 4. Start Go backend ──────────────────────────────
log "Starting Nexus-Scan backend on ws://127.0.0.1:${WS_PORT}..."
NEXUS_WS_PORT="$WS_PORT" "$BIN" >> "$LOG" 2>&1 &
BACKEND_PID=$!

# Wait for backend to be ready (up to 5 seconds)
for i in $(seq 1 10); do
    if curl -sf "http://127.0.0.1:${WS_PORT}/health" >/dev/null 2>&1; then
        log "Backend ready (PID: $BACKEND_PID)"
        break
    fi
    sleep 0.5
    if [[ $i -eq 10 ]]; then
        kill "$BACKEND_PID" 2>/dev/null || true
        die "Backend failed to start. Check $LOG"
    fi
done

cleanup() {
    log "Stopping Nexus-Scan backend..."
    kill -SIGTERM "$BACKEND_PID" 2>/dev/null || true
    wait "$BACKEND_PID" 2>/dev/null || true
}
trap cleanup EXIT INT TERM

# ── 5. Launch desktop app (Tauri) ────────────────────
LAUNCHED=0

# Option A: pre-built AppImage
if [[ -f "$SCRIPT_DIR/nexus-scan.AppImage" ]]; then
    log "Launching desktop app (AppImage)..."
    "$SCRIPT_DIR/nexus-scan.AppImage" &
    LAUNCHED=1

# Option B: system-installed .deb
elif command -v nexus-scan &>/dev/null; then
    log "Launching desktop app (system install)..."
    nexus-scan &
    LAUNCHED=1

# Option C: Tauri dev mode (requires Rust + pnpm)
elif command -v cargo &>/dev/null && command -v pnpm &>/dev/null; then
    log "Launching Tauri in dev mode (building frontend)..."
    cd "$SCRIPT_DIR/frontend"
    pnpm install --frozen-lockfile >/dev/null 2>&1 || pnpm install >/dev/null 2>&1
    pnpm tauri dev &
    LAUNCHED=1
fi

if [[ $LAUNCHED -eq 0 ]]; then
    warn "No Tauri app found. Build first with one of:"
    warn "  ./scripts/package-appimage.sh   → produces nexus-scan.AppImage"
    warn "  ./scripts/package-deb.sh        → produces nexus-scan.deb"
    warn "  cd frontend && pnpm tauri dev   → dev mode (needs Rust + pnpm)"
    warn ""
    warn "Backend is running at ws://127.0.0.1:${WS_PORT} — waiting..."
fi

log "Nexus-Scan running. Press Ctrl+C to stop."
wait "$BACKEND_PID"
