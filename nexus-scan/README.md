# Nexus-Scan

**Mobile Forensic & Malware Remediation Platform**

Production-grade desktop application for detecting and remediating malware on Android and iOS devices connected via USB. Built for Kali Linux.

---

## Stack

| Layer     | Technology                          |
|-----------|-------------------------------------|
| Backend   | Go 1.21+ (goroutines, worker pools) |
| Frontend  | Tauri 1.x + React 18 + TypeScript   |
| IPC       | WebSockets over localhost (port 9999)|
| Database  | SQLite (WAL mode, Drizzle-style schema)|
| Packaging | AppImage + Debian (.deb)             |
| Platform  | Kali Linux (primary), Ubuntu (secondary)|

---

## Architecture

```
nexus-scan/
├── cmd/nexus-scan/         # Entry point
├── internal/
│   ├── config/             # Environment-based config
│   └── server/             # WebSocket HTTP server + command router
├── pkg/
│   ├── logger/             # Structured JSON logging (zap)
│   ├── models/             # Shared data types
│   ├── database/           # SQLite DB + migrations + seed data
│   ├── ipc/                # WebSocket broadcaster + client pump
│   ├── device/             # USB device polling (adb + libimobiledevice)
│   ├── scanner/            # Concurrent scan engine (goroutine worker pool)
│   └── remediation/        # Uninstall engine (adb / ideviceinstaller)
├── frontend/
│   ├── src/
│   │   ├── App.tsx         # Root component + state reducer
│   │   ├── components/     # DevicePanel, ThreatTable, LogConsole, StatusBar
│   │   ├── hooks/          # useWebSocket (reconnect logic)
│   │   ├── types/          # Shared TypeScript interfaces
│   │   └── styles/         # Cyberpunk dark theme (CSS vars)
│   └── src-tauri/          # Tauri Rust shell + tauri.conf.json
├── tests/                  # Go unit tests
└── scripts/                # Shell scripts for build/install/package
```

---

## Prerequisites

- Kali Linux 2023+ or Ubuntu 22.04+
- USB debugging enabled on Android device (Developer Options)
- libimobiledevice pairing for iOS

---

## Step-by-Step Setup (Fresh Kali Linux)

### 1. Clone or copy the project

```bash
git clone <repo-url> nexus-scan
cd nexus-scan
```

### 2. Install all system dependencies

```bash
sudo ./scripts/install-deps.sh
```

This installs:
- Go 1.22
- Rust + cargo
- Node.js 20 LTS + pnpm
- `adb` (Android Debug Bridge)
- `libimobiledevice-utils` + `ideviceinstaller` (iOS)
- `libwebkit2gtk-4.0-dev` (Tauri dependency)
- USB udev rules for all major Android vendors

**After install:**
```bash
# Apply PATH changes
source /etc/profile.d/go.sh
source ~/.cargo/env
# Log out and back in for plugdev group changes
```

### 3. Verify dependencies

```bash
go version          # should show go1.22+
rustc --version     # should show rustc 1.76+
pnpm --version      # should show 8.x+
adb version         # should show Android Debug Bridge
idevice_id --version  # should show libimobiledevice version
```

### 4. Build the backend

```bash
./scripts/build-backend.sh
```

Binary output: `./bin/nexus-scan-backend`

### 5. Build the frontend

```bash
./scripts/build-frontend.sh
```

### 6. Package (choose one)

**AppImage** (portable, no install):
```bash
./scripts/package-appimage.sh
# Output: ./nexus-scan.AppImage
```

**Debian package** (system install):
```bash
./scripts/package-deb.sh
# Output: ./nexus-scan.deb
sudo dpkg -i nexus-scan.deb
```

### 7. Start the application

```bash
./scripts/start.sh
```

This:
1. Starts `adb start-server`
2. Starts `usbmuxd` for iOS
3. Launches the Go backend on `ws://127.0.0.1:9999`
4. Launches the Tauri frontend

---

## Development Mode

```bash
./scripts/dev.sh
```

Starts the Go backend + Tauri devserver with hot-reload on port 1420.

---

## Running Tests

```bash
go test ./tests/... -v
```

Tests cover:
- Database initialization and migrations
- Hash matching (known malware vs benign)
- Infection vector classification
- Command sanitization (injection prevention)
- Scan result persistence
- Remediation logging
- Threat simulation dataset

---

## WebSocket API

The backend listens on `ws://127.0.0.1:9999/ws`.

### Commands (client → server)

| Command         | Payload                                      | Description                    |
|-----------------|----------------------------------------------|--------------------------------|
| `start_scan`    | `{ "device_id": "emulator-5554" }`           | Start scanning a device        |
| `stop_scan`     | `{ "device_id": "emulator-5554" }`           | Stop an in-progress scan       |
| `remediate`     | `{ "device_id": "...", "package_name": "..." }` | Uninstall a package         |
| `list_devices`  | `{}`                                         | Get currently connected devices|
| `get_scan_history` | `{ "device_id": "..." }`                  | Retrieve past scan results     |
| `ping`          | `{}`                                         | Heartbeat check                |

### Events (server → client)

| Event                  | Description                                   |
|------------------------|-----------------------------------------------|
| `connected`            | Backend handshake                             |
| `device_detected`      | New device connected                          |
| `device_disconnected`  | Device removed                                |
| `device_list`          | Response to `list_devices`                    |
| `scan_started`         | Scan began                                    |
| `scan_progress`        | Total packages counted                        |
| `package_scanned`      | One package analyzed                          |
| `threat_detected`      | Malware/suspicious package found              |
| `scan_finished`        | Scan complete with summary                    |
| `scan_history`         | Response to `get_scan_history`                |
| `remediation_started`  | Uninstall initiated                           |
| `remediation_completed`| Uninstall result (success/fail)               |
| `remediation_failed`   | Uninstall failed with reason                  |
| `error`                | Backend error with code and message           |

---

## Database Schema

SQLite database at `./nexus-scan.db` (configurable via `NEXUS_DB_PATH`).

### `scan_history`
| Column           | Type    | Description                              |
|------------------|---------|------------------------------------------|
| `id`             | TEXT PK | UUID                                     |
| `device_id`      | TEXT    | Device serial number                     |
| `package_name`   | TEXT    | App package identifier                   |
| `file_path`      | TEXT    | Path on device                           |
| `sha256_hash`    | TEXT    | Package hash                             |
| `threat_name`    | TEXT    | Matched malware name (empty if clean)    |
| `severity`       | TEXT    | LOW / MEDIUM / HIGH / CRITICAL           |
| `infection_vector` | TEXT  | How the threat arrived                   |
| `malware_family` | TEXT    | Malware family                           |
| `scan_timestamp` | DATETIME| When the scan occurred                   |

### `malware_signatures`
| Column           | Type    | Description                              |
|------------------|---------|------------------------------------------|
| `id`             | INTEGER PK | Auto-increment                        |
| `malware_name`   | TEXT    | Signature display name                   |
| `sha256_hash`    | TEXT    | Hash to match (UNIQUE)                   |
| `malware_family` | TEXT    | Malware family group                     |
| `severity`       | TEXT    | LOW / MEDIUM / HIGH / CRITICAL           |
| `description`    | TEXT    | Human-readable threat description        |

### `remediation_log`
Append-only audit log of all uninstall actions.

---

## Environment Variables

| Variable              | Default              | Description                    |
|-----------------------|----------------------|--------------------------------|
| `NEXUS_WS_PORT`       | `9999`               | WebSocket server port          |
| `NEXUS_DB_PATH`       | `./nexus-scan.db`    | SQLite database path           |
| `NEXUS_LOG_LEVEL`     | `info`               | Log level (debug/info/warn/error)|
| `NEXUS_LOG_FILE`      | `./nexus-scan.log`   | JSON log file path             |
| `NEXUS_WORKER_COUNT`  | `8`                  | Concurrent scan goroutines     |
| `NEXUS_POLL_INTERVAL` | `3`                  | Device poll interval (seconds) |

---

## Android Device Setup

1. Enable **Developer Options**: Settings → About Phone → tap "Build Number" 7 times
2. Enable **USB Debugging**: Developer Options → USB Debugging → ON
3. Connect device via USB
4. Accept the RSA fingerprint prompt on the device
5. Verify: `adb devices` — device should appear as "device" (not "unauthorized")

## iOS Device Setup

1. Connect device via USB and trust the computer
2. Install `idevicepair` and pair: `idevicepair pair`
3. Verify: `idevice_id -l` — UDID should appear

---

## Threat Simulation Dataset (pre-seeded)

| Malware Name  | Family    | Severity | Description                              |
|---------------|-----------|----------|------------------------------------------|
| BankBot.A     | BankBot   | CRITICAL | Fake banking overlay + SMS OTP intercept |
| SpyAgent.B    | SpyAgent  | HIGH     | SMS/call logger + GPS tracker            |
| FakeWhatsApp.C| Trojan    | HIGH     | Cloned WhatsApp credential harvester     |
| Adware.D      | Adware    | MEDIUM   | Persistent full-screen adware            |
| SMSSpam.E     | SMSSpam   | LOW      | Unauthorized premium SMS sender          |
| Rootnik.F     | Rootnik   | CRITICAL | Local privilege escalation + root backdoor|
| GhostPush.G   | GhostPush | HIGH     | Silent APK installer via accessibility   |
| FakeAV.H      | FakeAV    | MEDIUM   | Scareware fake antivirus                 |
| Lotoor.I      | Lotoor    | CRITICAL | CVE-2012-0056 root exploit               |
| DroidDream.J  | DroidDream| CRITICAL | Early Google Play rootkit                |

---

## Security Notes

- WebSocket server binds to `127.0.0.1` only — not accessible from network
- All shell arguments sanitized before `os/exec` calls (no shell expansion)
- CSP policy set in `tauri.conf.json` — only `ws://127.0.0.1:9999` is allowed
- No external network calls — fully offline operation
- SQLite uses WAL mode for safe concurrent access
