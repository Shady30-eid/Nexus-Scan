#!/usr/bin/env bash
# Generate Tauri icon files using ImageMagick or Python (whichever is available).
set -euo pipefail

ICONS_DIR="$(dirname "${BASH_SOURCE[0]}")/../frontend/src-tauri/icons"
ICONS_DIR="$(mkdir -p "$ICONS_DIR" && cd "$ICONS_DIR" && pwd)"

log()  { echo -e "\033[0;32m[+]\033[0m $*"; }
warn() { echo -e "\033[1;33m[!]\033[0m $*"; }

# ── Draw a simple 512x512 "N" logo in cyberpunk colors ──────────────────────
generate_png_python() {
    local size="$1"
    local out="$2"
    python3 - "$size" "$out" << 'PYEOF'
import sys, struct, zlib

size = int(sys.argv[1])
path = sys.argv[2]

def make_png_rgba(w, h, pixels_rgba):
    """Build a minimal valid RGBA PNG from a flat list of (R,G,B,A) tuples."""
    def chunk(tag, data):
        crc = zlib.crc32(tag + data) & 0xFFFFFFFF
        return struct.pack(">I", len(data)) + tag + data + struct.pack(">I", crc)
    # color type 6 = RGBA (8 bits per channel)
    ihdr_data = struct.pack(">IIBBBBB", w, h, 8, 6, 0, 0, 0)
    raw = b""
    for y in range(h):
        raw += b"\x00"          # filter type None for each row
        for x in range(w):
            r, g, b, a = pixels_rgba[y * w + x]
            raw += bytes([r, g, b, a])
    idat_data = zlib.compress(raw, 9)
    return (b"\x89PNG\r\n\x1a\n"
            + chunk(b"IHDR", ihdr_data)
            + chunk(b"IDAT", idat_data)
            + chunk(b"IEND", b""))

bg     = (10,  10,  20,  255)   # dark navy
fg     = (0,  255, 200, 255)    # cyan-green
border = (0,  180, 255, 255)    # blue border

pixels = []
for y in range(size):
    for x in range(size):
        nx = x / size
        ny = y / size
        b  = max(4, size // 32) / size
        if nx < b or nx > 1 - b or ny < b or ny > 1 - b:
            pixels.append(border)
            continue
        cx = nx - 0.5
        in_left  = abs(cx + 0.22) < 0.06
        in_right = abs(cx - 0.22) < 0.06
        diag     = abs((ny - 0.15) - (cx + 0.22) / 0.44 * 0.70) < 0.06
        in_n     = (abs(ny - 0.5) < 0.35) and (in_left or in_right or diag)
        pixels.append(fg if in_n else bg)

with open(path, "wb") as f:
    f.write(make_png_rgba(size, size, pixels))
print(f"[python] wrote RGBA PNG {size}x{size} → {path}")
PYEOF
}

# ── Try ImageMagick first, fall back to Python ───────────────────────────────
make_icon() {
    local size="$1"
    local out="$2"

    if command -v convert &>/dev/null; then
        # -type TrueColorAlpha forces RGBA output required by Tauri
        convert -size "${size}x${size}" \
            xc:'#0a0a14' \
            -fill '#00ffc8' \
            -stroke '#00b4ff' \
            -strokewidth $((size/32 + 1)) \
            -font DejaVu-Sans-Bold \
            -pointsize $((size * 55 / 100)) \
            -gravity Center \
            -annotate 0 "N" \
            -type TrueColorAlpha \
            PNG32:"$out" 2>/dev/null && return 0
    fi

    # Fallback: pure Python (no deps, writes RGBA PNG directly)
    if command -v python3 &>/dev/null; then
        generate_png_python "$size" "$out" && return 0
    fi

    warn "Cannot generate icon $out (install imagemagick or python3)"
    return 1
}

# ── Generate all required sizes ──────────────────────────────────────────────
log "Generating Tauri icons in $ICONS_DIR ..."

make_icon 32  "$ICONS_DIR/32x32.png"
make_icon 128 "$ICONS_DIR/128x128.png"
make_icon 256 "$ICONS_DIR/128x128@2x.png"

# icon.ico  — multi-size Windows icon (just copy 32px PNG if no convert)
if command -v convert &>/dev/null; then
    convert "$ICONS_DIR/32x32.png" "$ICONS_DIR/128x128.png" \
        "$ICONS_DIR/icon.ico" 2>/dev/null || \
        cp "$ICONS_DIR/32x32.png" "$ICONS_DIR/icon.ico"
else
    cp "$ICONS_DIR/32x32.png" "$ICONS_DIR/icon.ico"
fi

# icon.icns — macOS icon (just copy 128px PNG; only needed on macOS builds)
cp "$ICONS_DIR/128x128.png" "$ICONS_DIR/icon.icns"

log "Icons ready:"
ls -lh "$ICONS_DIR/"
