#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
FRONTEND_DIR="$ROOT_DIR/frontend"

log()  { echo -e "\033[0;32m[+]\033[0m $*"; }
warn() { echo -e "\033[1;33m[!]\033[0m $*"; }
die()  { echo -e "\033[0;31m[x]\033[0m $*"; exit 1; }

cd "$FRONTEND_DIR"

source ~/.cargo/env 2>/dev/null || true
export PATH=$PATH:/usr/local/go/bin

command -v pnpm  &>/dev/null || die "pnpm not found"
command -v cargo &>/dev/null || die "cargo not found. Run: sudo ./scripts/install-deps.sh"

log "Installing Tauri CLI..."
cargo install tauri-cli --version "^1.6" --locked 2>&1 | tail -1 || warn "tauri-cli already up to date"

log "Installing frontend dependencies..."
pnpm install

log "Building AppImage via Tauri..."
cargo tauri build --bundles appimage 2>&1 || die "AppImage build failed"

APPIMAGE_PATH=$(find "$FRONTEND_DIR/src-tauri/target/release/bundle/appimage" -name "*.AppImage" 2>/dev/null | head -1)
if [[ -z "$APPIMAGE_PATH" ]]; then
    die "AppImage not found after build"
fi

cp "$APPIMAGE_PATH" "$ROOT_DIR/nexus-scan.AppImage"
chmod +x "$ROOT_DIR/nexus-scan.AppImage"

log "AppImage created: $ROOT_DIR/nexus-scan.AppImage"
log "Size: $(du -sh "$ROOT_DIR/nexus-scan.AppImage" | cut -f1)"
log ""
log "Run with:  ./nexus-scan.AppImage"
