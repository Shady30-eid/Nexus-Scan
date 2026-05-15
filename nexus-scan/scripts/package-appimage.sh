#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
FRONTEND_DIR="$ROOT_DIR/frontend"

log()  { echo -e "\033[0;32m[+]\033[0m $*"; }
warn() { echo -e "\033[1;33m[!]\033[0m $*"; }
die()  { echo -e "\033[0;31m[x]\033[0m $*"; exit 1; }

source ~/.cargo/env 2>/dev/null || true
export PATH="$PATH:/usr/local/go/bin:$HOME/.cargo/bin"

command -v pnpm  &>/dev/null || die "pnpm not found. Run: sudo ./scripts/install-deps.sh"
command -v cargo &>/dev/null || die "cargo not found. Run: sudo ./scripts/install-deps.sh"

# ── 1. Install Tauri CLI ─────────────────────────────
log "Installing Tauri CLI..."
cargo install tauri-cli --version "^1.6" --locked 2>&1 | tail -2 || true

# ── 2. Install frontend node_modules ────────────────
log "Installing frontend dependencies..."
cd "$FRONTEND_DIR"
# Use --ignore-workspace so pnpm installs only for this package,
# not for the entire Replit workspace root above it.
pnpm install --ignore-workspace 2>&1 || \
    npm install 2>&1 || \
    die "Failed to install frontend dependencies"

# ── 3. Build the frontend (Vite) ────────────────────
log "Building frontend (Vite)..."
pnpm run build 2>&1 || die "Vite build failed"

# ── 4. Build AppImage (skip beforeBuildCommand since we already built) ──
log "Building AppImage via Tauri..."
cargo tauri build --bundles appimage 2>&1 || die "AppImage build failed"

# ── 5. Copy output ───────────────────────────────────
APPIMAGE_PATH=$(find "$FRONTEND_DIR/src-tauri/target/release/bundle/appimage" \
    -name "*.AppImage" 2>/dev/null | head -1)

if [[ -z "$APPIMAGE_PATH" ]]; then
    die "AppImage not found after build"
fi

cp "$APPIMAGE_PATH" "$ROOT_DIR/nexus-scan.AppImage"
chmod +x "$ROOT_DIR/nexus-scan.AppImage"

log "AppImage created: $ROOT_DIR/nexus-scan.AppImage"
log "Size: $(du -sh "$ROOT_DIR/nexus-scan.AppImage" | cut -f1)"
log ""
log "Run with:  ./nexus-scan.AppImage"
