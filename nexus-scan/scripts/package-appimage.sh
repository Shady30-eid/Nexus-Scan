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

# ── 0. Install missing system libraries (GTK/WebKit/GLib) ──
install_system_libs() {
    local missing=()

    pkg-config --exists glib-2.0   2>/dev/null || missing+=("libglib2.0-dev")
    pkg-config --exists gtk+-3.0   2>/dev/null || missing+=("libgtk-3-dev")
    pkg-config --exists openssl    2>/dev/null || missing+=("libssl-dev")

    # WebKit: try 4.0 first, then 4.1 (Kali/Debian ≥2023 only ships 4.1)
    if ! pkg-config --exists webkit2gtk-4.0 2>/dev/null; then
        if apt-cache show libwebkit2gtk-4.0-dev &>/dev/null 2>&1; then
            missing+=("libwebkit2gtk-4.0-dev")
        else
            missing+=("libwebkit2gtk-4.1-dev")
        fi
    fi

    # libsoup-2.4 (needed by Tauri 1.x WebKit)
    if ! pkg-config --exists libsoup-2.4 2>/dev/null; then
        if apt-cache show libsoup2.4-dev &>/dev/null 2>&1; then
            missing+=("libsoup2.4-dev")
        fi
    fi

    if [[ ${#missing[@]} -gt 0 ]]; then
        warn "Missing system libraries: ${missing[*]}"
        log "Installing missing libraries (requires sudo)..."
        sudo apt-get update -qq
        sudo apt-get install -y --no-install-recommends \
            "${missing[@]}" \
            libayatana-appindicator3-dev \
            librsvg2-dev \
            patchelf \
            file \
            2>&1 || die "Failed to install system libraries. Run: sudo apt-get install ${missing[*]}"
        log "System libraries installed."
    fi

    SHIM_DIR="/usr/local/lib/pkgconfig"
    sudo mkdir -p "$SHIM_DIR"
    export PKG_CONFIG_PATH="$SHIM_DIR:${PKG_CONFIG_PATH:-}"

    # ── webkit2gtk-4.0 shim (Kali/Debian ≥2023 only has 4.1) ───────────
    if ! pkg-config --exists webkit2gtk-4.0 2>/dev/null && \
         pkg-config --exists webkit2gtk-4.1 2>/dev/null; then
        warn "Creating webkit2gtk-4.0 shim → 4.1"
        sudo tee "$SHIM_DIR/webkit2gtk-4.0.pc" > /dev/null << 'EOF'
Name: webkit2gtk-4.0
Description: WebKit2 GTK+ 4.0 (shim → 4.1)
Version: 2.42.0
Requires: webkit2gtk-4.1
Libs:
Cflags:
EOF
        sudo tee "$SHIM_DIR/javascriptcoregtk-4.0.pc" > /dev/null << 'EOF'
Name: javascriptcoregtk-4.0
Description: JavaScriptCore GTK 4.0 (shim → 4.1)
Version: 2.42.0
Requires: javascriptcoregtk-4.1
Libs:
Cflags:
EOF
    fi

    # ── libsoup-2.4 shim (Kali/Debian ≥2023 only has soup3) ────────────
    if ! pkg-config --exists libsoup-2.4 2>/dev/null && \
         pkg-config --exists libsoup-3.0 2>/dev/null; then
        warn "Creating libsoup-2.4 shim → 3.0"
        sudo tee "$SHIM_DIR/libsoup-2.4.pc" > /dev/null << 'EOF'
Name: libsoup-2.4
Description: libsoup 2.4 (shim → 3.0)
Version: 2.74.0
Requires: libsoup-3.0
Libs:
Cflags:
EOF
    fi

    log "System libraries OK."
}

install_system_libs

# ── 1. Install Tauri CLI ─────────────────────────────
log "Installing Tauri CLI..."
cargo install tauri-cli --version "^1.6" --locked 2>&1 | tail -2 || true

# ── 2. Install frontend node_modules ────────────────
log "Installing frontend dependencies..."
cd "$FRONTEND_DIR"

# Fix permissions on dist/ if a previous sudo run left root-owned files
if [[ -d "$FRONTEND_DIR/dist" ]]; then
    log "Cleaning old dist/ folder..."
    sudo rm -rf "$FRONTEND_DIR/dist" || rm -rf "$FRONTEND_DIR/dist" || true
fi

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
