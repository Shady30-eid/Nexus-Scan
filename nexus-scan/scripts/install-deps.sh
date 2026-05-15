#!/usr/bin/env bash
set -euo pipefail

# Nexus-Scan — Dependency Installer for Kali Linux / Ubuntu
# Run as root or with sudo

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LOG_FILE="/tmp/nexus-scan-install.log"

log()  { echo -e "\033[0;32m[+]\033[0m $*" | tee -a "$LOG_FILE"; }
warn() { echo -e "\033[1;33m[!]\033[0m $*" | tee -a "$LOG_FILE"; }
die()  { echo -e "\033[0;31m[x]\033[0m $*" | tee -a "$LOG_FILE"; exit 1; }

require_root() {
    if [[ $EUID -ne 0 ]]; then
        die "This script must be run as root (use: sudo $0)"
    fi
}

detect_distro() {
    if [[ -f /etc/os-release ]]; then
        . /etc/os-release
        echo "$ID"
    else
        die "Cannot detect Linux distribution"
    fi
}

install_system_packages() {
    log "Updating package lists..."
    apt-get update -qq 2>>"$LOG_FILE"

    log "Installing system build dependencies..."
    apt-get install -y --no-install-recommends \
        build-essential \
        curl \
        wget \
        git \
        pkg-config \
        libssl-dev \
        libsqlite3-dev \
        usbutils \
        2>>"$LOG_FILE"

    log "Installing Android forensics tools (adb)..."
    apt-get install -y --no-install-recommends \
        adb \
        android-tools-adb \
        2>>"$LOG_FILE" || warn "adb install failed — install manually from https://developer.android.com/tools/adb"

    log "Installing iOS forensics tools (libimobiledevice)..."
    apt-get install -y --no-install-recommends \
        libimobiledevice-utils \
        ideviceinstaller \
        ifuse \
        usbmuxd \
        2>>"$LOG_FILE" || warn "iOS tools install failed — continuing without iOS support"

    log "Installing Tauri/WebKit dependencies..."
    apt-get install -y --no-install-recommends \
        libwebkit2gtk-4.0-dev \
        libgtk-3-dev \
        libayatana-appindicator3-dev \
        librsvg2-dev \
        patchelf \
        file \
        2>>"$LOG_FILE"

    log "System packages installed."
}

install_go() {
    local GO_VERSION="1.22.2"
    local GO_TARBALL="go${GO_VERSION}.linux-amd64.tar.gz"
    local GO_URL="https://go.dev/dl/${GO_TARBALL}"

    if command -v go &>/dev/null; then
        CURRENT_GO=$(go version | awk '{print $3}' | sed 's/go//')
        log "Go already installed: v${CURRENT_GO}"
        return 0
    fi

    log "Downloading Go ${GO_VERSION}..."
    wget -q "$GO_URL" -O "/tmp/${GO_TARBALL}" 2>>"$LOG_FILE" || die "Failed to download Go"

    log "Installing Go to /usr/local/go..."
    rm -rf /usr/local/go
    tar -C /usr/local -xzf "/tmp/${GO_TARBALL}" 2>>"$LOG_FILE"
    rm "/tmp/${GO_TARBALL}"

    if ! grep -q '/usr/local/go/bin' /etc/profile.d/go.sh 2>/dev/null; then
        echo 'export PATH=$PATH:/usr/local/go/bin' > /etc/profile.d/go.sh
        chmod +x /etc/profile.d/go.sh
    fi

    export PATH=$PATH:/usr/local/go/bin
    log "Go ${GO_VERSION} installed at /usr/local/go"
}

install_rust() {
    if command -v rustc &>/dev/null; then
        log "Rust already installed: $(rustc --version)"
        return 0
    fi

    log "Installing Rust via rustup..."
    curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | \
        sh -s -- -y --no-modify-path 2>>"$LOG_FILE" || die "Rust installation failed"

    export PATH="$HOME/.cargo/bin:$PATH"
    log "Rust installed: $(rustc --version)"
}

install_node_pnpm() {
    if ! command -v node &>/dev/null; then
        log "Installing Node.js 20 LTS..."
        curl -fsSL https://deb.nodesource.com/setup_20.x | bash - 2>>"$LOG_FILE"
        apt-get install -y nodejs 2>>"$LOG_FILE" || die "Node.js installation failed"
    else
        log "Node.js already installed: $(node --version)"
    fi

    if ! command -v pnpm &>/dev/null; then
        log "Installing pnpm..."
        npm install -g pnpm 2>>"$LOG_FILE" || die "pnpm installation failed"
    else
        log "pnpm already installed: $(pnpm --version)"
    fi
}

setup_udev_rules() {
    log "Setting up udev rules for Android USB devices..."
    UDEV_RULES_FILE="/etc/udev/rules.d/51-android.rules"
    cat > "$UDEV_RULES_FILE" << 'UDEV'
# Common Android USB vendor IDs
SUBSYSTEM=="usb", ATTR{idVendor}=="0502", MODE="0666", GROUP="plugdev"   # Acer
SUBSYSTEM=="usb", ATTR{idVendor}=="0b05", MODE="0666", GROUP="plugdev"   # ASUS
SUBSYSTEM=="usb", ATTR{idVendor}=="413c", MODE="0666", GROUP="plugdev"   # Dell
SUBSYSTEM=="usb", ATTR{idVendor}=="0489", MODE="0666", GROUP="plugdev"   # Foxconn
SUBSYSTEM=="usb", ATTR{idVendor}=="18d1", MODE="0666", GROUP="plugdev"   # Google/Nexus
SUBSYSTEM=="usb", ATTR{idVendor}=="091e", MODE="0666", GROUP="plugdev"   # Garmin
SUBSYSTEM=="usb", ATTR{idVendor}=="0bb4", MODE="0666", GROUP="plugdev"   # HTC
SUBSYSTEM=="usb", ATTR{idVendor}=="12d1", MODE="0666", GROUP="plugdev"   # Huawei
SUBSYSTEM=="usb", ATTR{idVendor}=="2314", MODE="0666", GROUP="plugdev"   # INQ
SUBSYSTEM=="usb", ATTR{idVendor}=="0482", MODE="0666", GROUP="plugdev"   # Kyocera
SUBSYSTEM=="usb", ATTR{idVendor}=="1004", MODE="0666", GROUP="plugdev"   # LG
SUBSYSTEM=="usb", ATTR{idVendor}=="22b8", MODE="0666", GROUP="plugdev"   # Motorola
SUBSYSTEM=="usb", ATTR{idVendor}=="0409", MODE="0666", GROUP="plugdev"   # NEC
SUBSYSTEM=="usb", ATTR{idVendor}=="2080", MODE="0666", GROUP="plugdev"   # Nook
SUBSYSTEM=="usb", ATTR{idVendor}=="0955", MODE="0666", GROUP="plugdev"   # Nvidia
SUBSYSTEM=="usb", ATTR{idVendor}=="2257", MODE="0666", GROUP="plugdev"   # OTGV
SUBSYSTEM=="usb", ATTR{idVendor}=="10a9", MODE="0666", GROUP="plugdev"   # Pantech
SUBSYSTEM=="usb", ATTR{idVendor}=="1d4d", MODE="0666", GROUP="plugdev"   # Pegatron
SUBSYSTEM=="usb", ATTR{idVendor}=="0471", MODE="0666", GROUP="plugdev"   # Philips
SUBSYSTEM=="usb", ATTR{idVendor}=="04da", MODE="0666", GROUP="plugdev"   # PMC-Sierra
SUBSYSTEM=="usb", ATTR{idVendor}=="05c6", MODE="0666", GROUP="plugdev"   # Qualcomm
SUBSYSTEM=="usb", ATTR{idVendor}=="1f53", MODE="0666", GROUP="plugdev"   # SK Telesys
SUBSYSTEM=="usb", ATTR{idVendor}=="04e8", MODE="0666", GROUP="plugdev"   # Samsung
SUBSYSTEM=="usb", ATTR{idVendor}=="04dd", MODE="0666", GROUP="plugdev"   # Sharp
SUBSYSTEM=="usb", ATTR{idVendor}=="054c", MODE="0666", GROUP="plugdev"   # Sony
SUBSYSTEM=="usb", ATTR{idVendor}=="0fce", MODE="0666", GROUP="plugdev"   # Sony Ericsson
SUBSYSTEM=="usb", ATTR{idVendor}=="2340", MODE="0666", GROUP="plugdev"   # Teleepoch
SUBSYSTEM=="usb", ATTR{idVendor}=="0930", MODE="0666", GROUP="plugdev"   # Toshiba
SUBSYSTEM=="usb", ATTR{idVendor}=="19d2", MODE="0666", GROUP="plugdev"   # ZTE
SUBSYSTEM=="usb", ATTR{idVendor}=="2a70", MODE="0666", GROUP="plugdev"   # OnePlus
SUBSYSTEM=="usb", ATTR{idVendor}=="2d95", MODE="0666", GROUP="plugdev"   # vivo
SUBSYSTEM=="usb", ATTR{idVendor}=="1ebf", MODE="0666", GROUP="plugdev"   # Yulong/Coolpad
UDEV
    chmod 644 "$UDEV_RULES_FILE"
    udevadm control --reload-rules 2>>"$LOG_FILE" || warn "udevadm reload failed — reboot may be required"
    log "udev rules installed at $UDEV_RULES_FILE"

    if ! groups "$SUDO_USER" 2>/dev/null | grep -q plugdev; then
        usermod -aG plugdev "$SUDO_USER" 2>/dev/null || true
        log "Added $SUDO_USER to plugdev group (re-login required)"
    fi
}

main() {
    require_root

    log "=== Nexus-Scan Dependency Installer ==="
    log "Log file: $LOG_FILE"

    DISTRO=$(detect_distro)
    log "Detected distro: $DISTRO"

    case "$DISTRO" in
        kali|ubuntu|debian) ;;
        *) warn "Unsupported distro: $DISTRO — proceeding anyway (may fail)" ;;
    esac

    install_system_packages
    install_go
    install_rust
    install_node_pnpm
    setup_udev_rules

    log ""
    log "=== All dependencies installed successfully ==="
    log ""
    log "Next steps:"
    log "  1. Log out and back in (or reboot) for group changes to take effect"
    log "  2. Source Go PATH:  source /etc/profile.d/go.sh"
    log "  3. Source Rust PATH: source ~/.cargo/env"
    log "  4. Enable adb server: adb start-server"
    log "  5. Build and run:     ./scripts/build-backend.sh && ./scripts/start.sh"
}

main "$@"
