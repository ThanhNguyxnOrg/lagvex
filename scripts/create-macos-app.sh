#!/bin/sh
set -e

SRC_BIN="$1"
DEST_DIR="$2"
VER="${3:-1.0.0}"

if [ -z "$SRC_BIN" ] || [ -z "$DEST_DIR" ]; then
    echo "Usage: $0 <src_binary> <dest_dir> [version]"
    exit 1
fi

APP_DIR="${DEST_DIR}/Lagvex.app"
mkdir -p "${APP_DIR}/Contents/MacOS"
mkdir -p "${APP_DIR}/Contents/Resources"

cp "$SRC_BIN" "${APP_DIR}/Contents/MacOS/lagvex-client"
chmod +x "${APP_DIR}/Contents/MacOS/lagvex-client"

cat << EOF > "${APP_DIR}/Contents/Info.plist"
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleExecutable</key>
    <string>lagvex-client</string>
    <key>CFBundleIdentifier</key>
    <string>org.thanhnguyen.lagvex</string>
    <key>CFBundleName</key>
    <string>Lagvex</string>
    <key>CFBundleDisplayName</key>
    <string>Lagvex</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>CFBundleShortVersionString</key>
    <string>${VER}</string>
    <key>CFBundleVersion</key>
    <string>${VER}</string>
    <key>LSMinimumSystemVersion</key>
    <string>11.0</string>
    <key>NSHighResolutionCapable</key>
    <true/>
    <key>LSUIElement</key>
    <false/>
</dict>
</plist>
EOF

echo "Created ${APP_DIR} successfully."
