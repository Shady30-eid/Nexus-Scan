# Nexus-Scan

> **Mobile Forensic & Malware Remediation Platform**  
> Detect, analyze, and remove malware from Android and iOS devices connected via USB — directly from your Kali Linux desktop.

---

## What Is Nexus-Scan?

Nexus-Scan is a production-grade desktop application built for security researchers and forensic analysts. It automatically detects USB-connected mobile devices, scans all installed applications, cross-references them against a local malware signature database, classifies threats by severity, and allows one-click removal — all from a dark cyberpunk-styled GUI.

### Key Features

| Feature | Description |
|---|---|
| **Auto Device Detection** | Detects Android (via ADB) and iOS (via libimobiledevice) the moment they're plugged in |
| **Concurrent Scanning** | Scans 500+ apps simultaneously using goroutine worker pools — no UI blocking |
| **Threat Intelligence** | Cross-references SHA256 hashes against a local malware signature database |
| **Infection Vector Analysis** | Classifies how malware arrived (browser download, sideloading, social engineering) |
| **One-Click Remediation** | Uninstalls malicious apps directly from the GUI using ADB / ideviceinstaller |
| **Live Log Console** | Real-time structured JSON logging shown inside the app |
| **Severity Levels** | Classifies every threat as LOW / MEDIUM / HIGH / CRITICAL |
| **Fully Offline** | No internet required — all analysis is done locally |

---

## Architecture

```
Nexus-Scan
├── Go Backend          WebSocket server on ws://127.0.0.1:9999
│   ├── USB Poller      Detects Android/iOS devices every 3 seconds
│   ├── Scanner         Concurrent goroutine worker pool (8 workers)
│   ├── Threat Intel    SQLite malware signature database
│   └── Remediation     ADB / ideviceinstaller uninstall engine
│
└── Tauri Frontend      Desktop GUI (not a browser tab)
    ├── Device Panel    Shows connected devices + scan controls
    ├── Threat Table    Live threat feed with severity indicators
    ├── Log Console     Real-time backend log stream
    └── Status Bar      WebSocket connection health indicator
```

---

## Requirements

### Operating System
- **Kali Linux 2023+** (primary)
- Ubuntu 22.04 LTS (secondary)

### Hardware
- USB port for connecting mobile devices
- Android device with **USB Debugging enabled**
- iOS device with **"Trust This Computer"** accepted

### Software (auto-installed by the install script)
| Tool | Purpose |
|---|---|
| Go 1.22+ | Build and run the backend |
| Rust + Cargo | Build the Tauri desktop shell |
| Node.js 20+ | Build the frontend UI |
| pnpm | Frontend package manager |
| `adb` | Android device communication |
| `libimobiledevice-utils` | iOS device communication |
| `ideviceinstaller` | iOS app management |
| `libwebkit2gtk-4.0-dev` | Tauri WebView engine |

---

## Installation

### Step 1 — Clone the repository

```bash
git clone https://github.com/YOUR_USERNAME/Nexus-Scan.git
cd Nexus-Scan
```

### Step 2 — Enter the project folder

```bash
cd nexus-scan
```

### Step 3 — Install all dependencies (one command)

```bash
sudo ./scripts/install-deps.sh
```

This script will automatically install:
- Go 1.22
- Rust via rustup
- Node.js 20 + pnpm
- ADB + libimobiledevice
- Tauri build dependencies (WebKit, GTK)
- USB udev rules for all major Android vendors

> **This takes 3–10 minutes on first run.**

### Step 4 — Apply PATH changes

```bash
source /etc/profile.d/go.sh
source ~/.cargo/env
```

> You only need to do this once per terminal session. After a reboot it applies automatically.

### Step 5 — Verify everything is installed

```bash
go version          # go1.22.x
rustc --version     # rustc 1.76+
pnpm --version      # 8.x+
adb version         # Android Debug Bridge
idevice_id -h       # libimobiledevice
```

---

## Android Device Setup (one-time)

1. On the phone: go to **Settings → About Phone**
2. Tap **Build Number** 7 times to unlock Developer Options
3. Go to **Settings → Developer Options**
4. Enable **USB Debugging**
5. Connect phone via USB cable
6. Accept the **"Allow USB Debugging?"** popup on the phone
7. Verify on your Kali machine:

```bash
adb devices
# Should show: XXXXXXXX   device
```

---

## iOS Device Setup (one-time)

```bash
# Pair your iPhone/iPad
idevicepair pair

# Verify connection
idevice_id -l
# Shows device UDID if paired correctly
```

---

## Build

### Build backend only

```bash
cd nexus-scan
./scripts/build-backend.sh
```

### Build desktop app (AppImage — portable, no install needed)

```bash
cd nexus-scan
./scripts/package-appimage.sh
# Output: nexus-scan/nexus-scan.AppImage
```

### Build Debian package (.deb — system install)

```bash
cd nexus-scan
./scripts/package-deb.sh
# Output: nexus-scan/nexus-scan.deb

# Install it:
sudo dpkg -i nexus-scan.deb
```

---

## Running Nexus-Scan

### Option A — One-command launch (recommended)

```bash
cd nexus-scan
./run.sh
```

This single command will:
1. Build the Go backend (if not already built)
2. Start ADB server for Android detection
3. Start usbmuxd for iOS detection
4. Launch the Go WebSocket backend on port 9999
5. Open the desktop app window automatically

---

### Option B — Development mode (with hot-reload)

```bash
cd nexus-scan
./scripts/dev.sh
```

---

### Option C — Run backend + AppImage separately

```bash
# Terminal 1: start backend
cd nexus-scan
./bin/nexus-scan-backend

# Terminal 2: launch desktop app
./nexus-scan.AppImage
```

---

## Usage

1. Connect an Android or iOS device via USB
2. The device appears automatically in the **Device Panel** on the left
3. Click **▶ SCAN** to begin forensic analysis
4. Watch the **Threat Table** populate in real time
5. Expand any threat row to see full details (hash, family, description)
6. Click **✗ PURGE** on any threat to remove it from the device
7. Confirm the removal in the dialog — app is uninstalled via ADB/ideviceinstaller
8. All actions are logged in the **Live Log Console** at the bottom

---

## Environment Variables (optional)

| Variable | Default | Description |
|---|---|---|
| `NEXUS_WS_PORT` | `9999` | WebSocket server port |
| `NEXUS_DB_PATH` | `./nexus-scan.db` | SQLite database path |
| `NEXUS_LOG_LEVEL` | `info` | Log level: debug / info / warn / error |
| `NEXUS_LOG_FILE` | `./nexus-scan.log` | Log file path |
| `NEXUS_WORKER_COUNT` | `8` | Parallel scan goroutines |
| `NEXUS_POLL_INTERVAL` | `3` | USB device poll interval (seconds) |

Example with custom settings:

```bash
NEXUS_WS_PORT=8888 NEXUS_WORKER_COUNT=16 ./run.sh
```

---

## Running Tests

```bash
cd nexus-scan
go test ./tests/... -v
```

Tests cover:
- Database initialization and migrations
- Hash matching against known malware signatures
- Infection vector classification logic
- Shell argument sanitization (injection prevention)
- Scan result persistence
- Remediation audit logging

---

## Pre-loaded Threat Signatures

The database comes pre-seeded with 10 real-world malware signatures for testing:

| Name | Family | Severity | Description |
|---|---|---|---|
| BankBot.A | BankBot | CRITICAL | Fake banking overlay, steals credentials + OTP |
| SpyAgent.B | SpyAgent | HIGH | SMS/call logger + silent GPS tracker |
| FakeWhatsApp.C | Trojan | HIGH | Cloned WhatsApp steals contacts + messages |
| Adware.D | Adware | MEDIUM | Full-screen persistent adware, survives reboot |
| SMSSpam.E | SMSSpam | LOW | Sends premium-rate SMS without consent |
| Rootnik.F | Rootnik | CRITICAL | Local privilege escalation → root backdoor |
| GhostPush.G | GhostPush | HIGH | Silently installs apps via accessibility API |
| FakeAV.H | FakeAV | MEDIUM | Scareware demanding payment for fake "removal" |
| Lotoor.I | Lotoor | CRITICAL | Exploits CVE-2012-0056 for root access |
| DroidDream.J | DroidDream | CRITICAL | Early rootkit distributed via Google Play |

---

## Security Notes

- The WebSocket server **only binds to 127.0.0.1** — not accessible from the network
- All shell arguments are **sanitized** before any `exec` call — no shell injection possible
- The Tauri CSP policy only allows `ws://127.0.0.1:9999` — no external connections
- All analysis is **fully offline** — no data leaves your machine

---

## Troubleshooting

**`./scripts/install-deps.sh: command not found`**
```bash
# Make sure you are inside the nexus-scan folder:
cd nexus-scan
sudo ./scripts/install-deps.sh
```

**`adb: command not found`**
```bash
sudo apt-get install adb
adb start-server
```

**Device shows as `unauthorized` in `adb devices`**
```bash
# Revoke and re-accept USB debugging on the phone:
adb kill-server
adb start-server
# Then accept the popup on the phone
```

**AppImage won't open**
```bash
chmod +x nexus-scan.AppImage
./nexus-scan.AppImage
```

**Backend not starting**
```bash
# Check the log file:
cat nexus-scan.log
# Make sure port 9999 is free:
ss -tlnp | grep 9999
```

---

## License

For educational and authorized security research use only.  
Do not use against devices you do not own or have explicit permission to test.
