#!/usr/bin/env bash
# Generate Tauri icon files using ImageMagick or Python (whichever is available).
set -euo pipefail

ICONS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../frontend/src-tauri/icons" && pwd)"
mkdir -p "$ICONS_DIR"

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

def make_png(w, h, pixels_rgba):
    def chunk(tag, data):
        c = zlib.crc32(tag + data) & 0xFFFFFFFF
        return struct.pack(">I", len(data)) + tag + data + struct.pack(">I", c)
    ihdr = struct.pack(">IIBBBBB", w, h, 8, 2, 0, 0, 0)
    raw = b""
    for y in range(h):
        raw += b"\x00"
        for x in range(w):
            raw += bytes(pixels_rgba[y * w + x][:3])
    idat = zlib.compress(raw)
    return b"\x89PNG\r\n\x1a\n" + chunk(b"IHDR", ihdr) + chunk(b"IDAT", idat) + chunk(b"IEND", b"")

pixels = []
bg = (10, 10, 20, 255)
fg = (0, 255, 200, 255)
border = (0, 180, 255, 255)

for y in range(size):
    for x in range(size):
        nx = x / size
        ny = y / size
        b = 4 / size
        # border
        if nx < b or nx > 1-b or ny < b or ny > 1-b:
            pixels.append(border)
            continue
        # letter N
        cx = nx - 0.5
        cy = ny - 0.5
        in_left  = abs(cx + 0.22) < 0.06
        in_right = abs(cx - 0.22) < 0.06
        diag = abs((cy + 0.35) - (cx + 0.22) / 0.44 * 0.70) < 0.06
        in_n = (abs(ny - 0.5) < 0.35) and (in_left or in_right or diag)
        pixels.append(fg if in_n else bg)

with open(path, "wb") as f:
    f.write(make_png(size, size, pixels))
PYEOF
}

# ── Try ImageMagick first, fall back to Python ───────────────────────────────
make_icon() {
    local size="$1"
    local out="$2"

    if command -v convert &>/dev/null; then
        convert -size "${size}x${size}" \
            xc:'#0a0a14' \
            -fill '#00ffc8' \
            -stroke '#00b4ff' \
            -strokewidth $((size/32 + 1)) \
            -font DejaVu-Sans-Bold \
            -pointsize $((size * 55 / 100)) \
            -gravity Center \
            -annotate 0 "N" \
            -stroke '#00b4ff' \
            -strokewidth $((size/32 + 1)) \
            -draw "rectangle 0,0 $((size-1)),$((size/16)) \
                   rectangle 0,$((size - size/16)) $((size-1)),$((size-1)) \
                   rectangle 0,0 $((size/16)),$((size-1)) \
                   rectangle $((size - size/16)),0 $((size-1)),$((size-1))" \
            "$out" 2>/dev/null && return 0
    fi

    # Fallback: pure Python (no deps)
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
