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
command -v dpkg-deb &>/dev/null || die "dpkg-deb not found. Install: apt-get install dpkg"

log "Installing Tauri CLI..."
cargo install tauri-cli --version "^1.6" 2>/dev/null || warn "tauri-cli already installed"

log "Installing frontend dependencies..."
pnpm install

log "Building .deb package via Tauri..."
pnpm exec cargo-tauri build --bundles deb 2>&1 || die "Debian package build failed"

DEB_PATH=$(find "$FRONTEND_DIR/src-tauri/target/release/bundle/deb" -name "*.deb" 2>/dev/null | head -1)
if [[ -z "$DEB_PATH" ]]; then
    die ".deb package not found after build"
fi

cp "$DEB_PATH" "$ROOT_DIR/nexus-scan.deb"

log ".deb package created: $ROOT_DIR/nexus-scan.deb"
log "Size: $(du -sh "$ROOT_DIR/nexus-scan.deb" | cut -f1)"
log ""
log "Install with:   sudo dpkg -i nexus-scan.deb"
log "Uninstall with: sudo dpkg -r nexus-scan"
