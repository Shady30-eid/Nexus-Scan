#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
FRONTEND_DIR="$ROOT_DIR/frontend"

log()  { echo -e "\033[0;32m[+]\033[0m $*"; }
die()  { echo -e "\033[0;31m[x]\033[0m $*"; exit 1; }

cd "$FRONTEND_DIR"

command -v pnpm &>/dev/null || die "pnpm not found. Run: sudo ./scripts/install-deps.sh"
command -v rustc &>/dev/null || { source ~/.cargo/env 2>/dev/null || true; }
command -v rustc &>/dev/null || die "Rust not found. Run: sudo ./scripts/install-deps.sh"

log "Installing frontend npm dependencies..."
pnpm install || die "pnpm install failed"

log "Type-checking TypeScript..."
pnpm exec tsc --noEmit || die "TypeScript type-check failed"

log "Building frontend (Vite)..."
pnpm build || die "Vite build failed"

log "Frontend built successfully at: $FRONTEND_DIR/dist"
log ""
log "To build the full Tauri AppImage/deb, run:"
log "  ./scripts/package-appimage.sh"
log "  ./scripts/package-deb.sh"
