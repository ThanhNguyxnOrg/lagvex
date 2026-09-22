#!/bin/sh
# Start Lagvex Client in background without keeping terminal open
cd "$(dirname "$0")"
chmod +x ./lagvex-client 2>/dev/null || true
if command -v nohup >/dev/null 2>&1; then
    nohup ./lagvex-client "$@" >/dev/null 2>&1 &
else
    ./lagvex-client -daemon "$@"
fi
echo "Lagvex Client started in background (Web Dashboard: http://127.0.0.1:18888)"
