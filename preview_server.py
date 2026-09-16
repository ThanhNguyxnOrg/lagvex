#!/usr/bin/env python3
"""
Lagvex Live UI Preview & Interactive Demo Server
Runs without requiring Windows UAC administrator elevation.
"""

import http.server
import json
import os
import socketserver
import sys

PORT = 18888
BASE_DIR = os.path.dirname(os.path.abspath(__file__))
WEB_DIR = os.path.join(BASE_DIR, "web")
PROFILES_FILE = os.path.join(BASE_DIR, "configs", "profiles.json")

# Simulated state
is_connected = False
current_ping = 24

class LagvexDemoHandler(http.server.SimpleHTTPRequestHandler):
    def __init__(self, *args, **kwargs):
        super().__init__(*args, directory=WEB_DIR, **kwargs)

    def do_GET(self):
        global is_connected, current_ping
        if self.path == "/api/games":
            self.send_json(self.load_games())
        elif self.path == "/api/relays":
            self.send_json(self.load_relays())
        elif self.path.startswith("/api/best-relay"):
            self.send_json({
                "name": "Singapore #1",
                "endpoint": "128.199.200.10:4433",
                "location": "Singapore",
                "continent": "Asia",
                "rttMedianMs": 18,
                "jitterMs": 1.1,
                "packetLoss": 0.0,
                "score": 99.2
            })
        elif self.path.startswith("/api/probe-relays"):
            relays = self.load_relays()
            probed = []
            for r in relays:
                probed.append({
                    "name": r.get("name", "Node"),
                    "endpoint": r.get("endpoint", ""),
                    "location": r.get("location", ""),
                    "continent": r.get("continent", "Global"),
                    "rttMedianMs": 18 if "Singapore" in r.get("name","") else 32 if "Tokyo" in r.get("name","") else 65,
                    "jitterMs": 1.2,
                    "packetLoss": 0.0,
                    "score": 95.0
                })
            self.send_json(probed)
        elif self.path == "/api/status":
            self.send_json({
                "state": "accelerating" if is_connected else "standby",
                "pingMs": current_ping if is_connected else 0,
                "upRateBps": 184320 if is_connected else 0,
                "downRateBps": 524288 if is_connected else 0,
                "routeCount": 24 if is_connected else 0,
                "gameRunning": True
            })
        else:
            # Fallback to serving static files from web/
            super().do_GET()

    def do_POST(self):
        global is_connected, current_ping
        if self.path == "/api/connect":
            is_connected = True
            current_ping = 22
            self.send_json({"status": "connecting", "endpoint": "128.199.200.10:4433"})
        elif self.path == "/api/disconnect":
            is_connected = False
            self.send_json({"status": "disconnected"})
        elif self.path == "/api/test-relay" or self.path == "/api/ping-relay":
            self.send_json({
                "reachable": True,
                "name": "Singapore #1 (Tier-1 Subsea)",
                "latencyMs": 18,
                "jitterMs": 1.2,
                "packetLoss": 0.0
            })
        elif self.path == "/api/add-game":
            self.send_json({"status": "ok"})
        else:
            self.send_error(404, "Endpoint Not Found")

    def send_json(self, data, status=200):
        body = json.dumps(data).encode("utf-8")
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.send_header("Access-Control-Allow-Origin", "*")
        self.end_headers()
        self.wfile.write(body)

    def load_games(self):
        if os.path.exists(PROFILES_FILE):
            try:
                with open(PROFILES_FILE, "r", encoding="utf-8") as f:
                    return json.load(f).get("games", [])
            except Exception:
                pass
        return []

    def load_relays(self):
        if os.path.exists(PROFILES_FILE):
            try:
                with open(PROFILES_FILE, "r", encoding="utf-8") as f:
                    return json.load(f).get("relays", [])
            except Exception:
                pass
        return []

if __name__ == "__main__":
    socketserver.TCPServer.allow_reuse_address = True
    with socketserver.TCPServer(("127.0.0.1", PORT), LagvexDemoHandler) as httpd:
        print(f"[Lagvex Demo Server] Active at http://127.0.0.1:{PORT}")
        sys.stdout.flush()
        try:
            httpd.serve_forever()
        except KeyboardInterrupt:
            pass
