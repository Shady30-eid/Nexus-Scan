#!/usr/bin/env bash
# Nexus-Scan — One-shot launcher (desktop app, Kali Linux)
# Usage: ./run.sh

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN="$SCRIPT_DIR/bin/nexus-scan-backend"
LOG="$SCRIPT_DIR/nexus-scan.log"
WS_PORT="${NEXUS_WS_PORT:-9999}"

GREEN="\033[0;32m"; YELLOW="\033[1;33m"; RED="\033[0;31m"; NC="\033[0m"
log()  { echo -e "${GREEN}[+]${NC} $*"; }
warn() { echo -e "${YELLOW}[!]${NC} $*"; }
err()  { echo -e "${RED}[x]${NC} $*"; }

# ── Fix PATH ─────────────────────────────────────────
export PATH="$PATH:/usr/local/go/bin:$HOME/.cargo/bin"
source ~/.cargo/env 2>/dev/null || true

# ── 1. Build Go backend if binary missing ────────────
if [[ ! -f "$BIN" ]]; then
    log "Backend binary not found — building..."

    if ! command -v go &>/dev/null; then
        err "Go not installed. Run:  sudo ./scripts/install-deps.sh"
        err "Or download Go from:   https://go.dev/dl/"
        # Still try to launch the GUI — backend is needed for real scans
        # but the window will open (just WebSocket will fail to connect)
    else
        mkdir -p "$SCRIPT_DIR/bin"
        cd "$SCRIPT_DIR"

        # Make sure gcc + libsqlite3 are present (needed by go-sqlite3)
        if ! command -v gcc &>/dev/null; then
            warn "gcc not found — installing build-essential..."
            sudo apt-get install -y --no-install-recommends build-essential libsqlite3-dev 2>>"$LOG"
        fi

        # Delete old DB so it gets re-seeded with updated malware signatures
        if [[ -f "$SCRIPT_DIR/nexus-scan.db" ]]; then
            log "Removing old database for re-seeding with updated malware signatures..."
            rm -f "$SCRIPT_DIR/nexus-scan.db"
        fi

        log "Compiling backend (this may take a minute)..."
        # Remove stale go.sum so Go regenerates it fresh
        rm -f "$SCRIPT_DIR/go.sum"
        # GONOSUMDB=* + GONOSUMCHECK=* skip checksum DB verification entirely
        GONOSUMDB="*" \
        GONOSUMCHECK="*" \
        GOFLAGS="-mod=mod" \
        GOPATH="${GOPATH:-$HOME/go}" \
        CGO_ENABLED=1 \
        go build \
            -ldflags="-s -w" \
            -o "$BIN" \
            ./cmd/nexus-scan/ \
            2>>"$LOG"
        if [[ $? -eq 0 ]]; then
            log "Backend built: $BIN"
        else
            err "Backend build failed. Check $LOG"
            err "Last error lines:"
            tail -10 "$LOG" 2>/dev/null | sed 's/^/    /'
        fi
    fi
fi

# ── 2. Start ADB daemon ──────────────────────────────
if command -v adb &>/dev/null; then
    log "Starting ADB server..."
    adb start-server 2>/dev/null || warn "ADB start failed (Android won't be detected)"
else
    warn "adb not found — install with: sudo apt-get install adb"
fi

# ── 3. Start usbmuxd (iOS support) ───────────────────
if command -v usbmuxd &>/dev/null; then
    systemctl is-active --quiet usbmuxd 2>/dev/null || \
        usbmuxd --user 2>/dev/null || \
        warn "usbmuxd failed to start (iOS detection disabled)"
fi

# ── 4. Start Go backend ───────────────────────────────
BACKEND_PID=""
if [[ -f "$BIN" ]]; then
    log "Starting Nexus-Scan backend on ws://127.0.0.1:${WS_PORT}..."
    mkdir -p "$(dirname "$LOG")"
    NEXUS_WS_PORT="$WS_PORT" "$BIN" >> "$LOG" 2>&1 &
    BACKEND_PID=$!

    # Wait up to 5 seconds for backend to be ready
    READY=0
    for i in $(seq 1 10); do
        sleep 0.5
        if curl -sf "http://127.0.0.1:${WS_PORT}/health" >/dev/null 2>&1; then
            log "Backend ready (PID: $BACKEND_PID)"
            READY=1
            break
        fi
    done

    if [[ $READY -eq 0 ]]; then
        err "Backend failed to start after 5 seconds."
        err "Check log: $LOG"
        err "Last lines:"
        tail -5 "$LOG" 2>/dev/null | sed 's/^/  /'
    fi
else
    warn "Backend binary not found — GUI will open but scanning won't work."
    warn "Build backend with:  cd nexus-scan && go build -o bin/nexus-scan-backend ./cmd/nexus-scan/"
fi

cleanup() {
    if [[ -n "$BACKEND_PID" ]]; then
        log "Stopping backend (PID: $BACKEND_PID)..."
        kill -SIGTERM "$BACKEND_PID" 2>/dev/null || true
        wait "$BACKEND_PID" 2>/dev/null || true
    fi
}
trap cleanup EXIT INT TERM

# ── 5. Launch desktop app ─────────────────────────────
LAUNCHED=0

# Option A: AppImage (APPIMAGE_EXTRACT_AND_RUN avoids FUSE requirement)
if [[ -f "$SCRIPT_DIR/nexus-scan.AppImage" ]]; then
    log "Launching Nexus-Scan (AppImage)..."
    APPIMAGE_EXTRACT_AND_RUN=1 "$SCRIPT_DIR/nexus-scan.AppImage" &
    LAUNCHED=1

# Option B: system-installed via .deb
elif command -v nexus-scan &>/dev/null; then
    log "Launching Nexus-Scan (system install)..."
    nexus-scan &
    LAUNCHED=1

# Option C: Tauri dev mode
elif command -v cargo &>/dev/null && command -v pnpm &>/dev/null; then
    log "Launching Tauri dev mode..."
    cd "$SCRIPT_DIR/frontend"
    pnpm install --ignore-workspace >/dev/null 2>&1 || true
    pnpm tauri dev &
    LAUNCHED=1
fi

if [[ $LAUNCHED -eq 0 ]]; then
    warn ""
    warn "No Nexus-Scan GUI found. Build it first:"
    warn "  ./scripts/package-appimage.sh   → nexus-scan.AppImage"
    warn "  ./scripts/package-deb.sh        → nexus-scan.deb"
    warn ""
fi

if [[ -n "$BACKEND_PID" ]]; then
    log "Nexus-Scan running. Press Ctrl+C to stop."
    log "Logs: $LOG"
    wait "$BACKEND_PID"
else
    # No backend — just keep GUI alive
    log "Press Ctrl+C to exit."
    wait
fi
