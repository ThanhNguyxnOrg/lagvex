package client

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/ThanhNguyxnOrg/lagvex/pkg/profiles"
	"github.com/ThanhNguyxnOrg/lagvex/pkg/protocol"
)

// TestBackendRealEndToEndProof exercises the complete real backend stack over
// live UDP sockets and guarded local HTTP APIs without mocks:
//  1. Two real lagvex-relay servers with two DIFFERENT PSKs
//  2. lagvex:// Share URI import -> live UDP probe sweep
//  3. Live encrypted ChaCha20-Poly1305 connect + authenticated Pong
//  4. Independent-PSK Dual-Path activation against the 2nd live relay
//  5. Real WireGuard (wireguard-go) .conf import -> ConnectWireGuard -> Disconnect
//  6. Anti-DPI UDP Obfuscation toggle -> live obfuscated connect + Pong
func TestBackendRealEndToEndProof(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("live loopback relay integration requires Linux userspace TUN fallback")
	}

	psk1 := []byte("real-e2e-primary-psk-0123456789")
	psk2 := []byte("real-e2e-standby-psk-9876543210")

	srv1 := startE2ERelay(t, psk1)
	srv2 := startE2ERelay(t, psk2)

	pm, err := profiles.NewManager("")
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	eng, err := NewEngine(pm)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}
	t.Cleanup(func() { _ = eng.Disconnect() })

	ui := &UIServer{
		engine:      eng,
		profileMgr:  pm,
		prober:      NewProber(),
		configStore: NewConfigStore(filepath.Join(t.TempDir(), "user_config.json")),
		cpuSampler:  NewProcessCPUSampler(),
	}
	h := withLocalRequestGuard(ui.setupRoutes())

	// 1. Import both live relays via lagvex:// Share URIs
	uri1, err := profiles.EncodeRelayShareURI(profiles.RelayEndpoint{
		ID: "primary-live", Name: "Primary Live", Continent: "Asia", Location: "Singapore",
		Endpoint: srv1.UDPAddr().String(), PSK: string(psk1),
	})
	if err != nil {
		t.Fatalf("EncodeRelayShareURI(1): %v", err)
	}
	uri2, err := profiles.EncodeRelayShareURI(profiles.RelayEndpoint{
		ID: "standby-live", Name: "Standby Live", Continent: "Asia", Location: "Tokyo",
		Endpoint: srv2.UDPAddr().String(), PSK: string(psk2),
	})
	if err != nil {
		t.Fatalf("EncodeRelayShareURI(2): %v", err)
	}

	for _, uri := range []string{uri1, uri2} {
		rec := doGuardedPost(t, h, "/api/relay/import-uri", map[string]string{"uri": uri})
		if rec.Code != http.StatusOK {
			t.Fatalf("import-uri failed: %d %s", rec.Code, rec.Body.String())
		}
	}

	// 2. Live UDP probe sweep against both real relays
	probeReq := httptest.NewRequest(http.MethodGet, "/api/probe-relays?force=1", nil)
	probeRec := httptest.NewRecorder()
	h.ServeHTTP(probeRec, probeReq)
	if probeRec.Code != http.StatusOK {
		t.Fatalf("probe-relays status=%d body=%s", probeRec.Code, probeRec.Body.String())
	}
	var probes []ProbeResult
	if err := json.Unmarshal(probeRec.Body.Bytes(), &probes); err != nil {
		t.Fatalf("decode probes: %v", err)
	}
	if len(probes) != 2 || !probes[0].Reachable || !probes[1].Reachable {
		t.Fatalf("expected 2 reachable live relays, got: %+v", probes)
	}

	// 3. Connect via HTTP API to Primary Live relay
	connRec := doGuardedPost(t, h, "/api/connect", map[string]any{
		"relayEndpoint": srv1.UDPAddr().String(),
		"gameId":        "valorant",
		"regionId":      "asia-sg",
	})
	if connRec.Code != http.StatusOK {
		t.Fatalf("connect failed: %d %s", connRec.Code, connRec.Body.String())
	}
	waitActiveUsers(t, srv1.HTTPAddr(), 1)

	// Wait for real authenticated Pong
	deadline := time.Now().Add(6 * time.Second)
	for time.Now().Before(deadline) && !eng.Stats().PingSampled {
		time.Sleep(25 * time.Millisecond)
	}
	if !eng.Stats().PingSampled {
		t.Fatal("primary relay never returned an authenticated Pong")
	}

	// 4. Seed standby measurement for srv2 (different PSK!) and toggle Dual-Path ON via HTTP API
	eng.standbyMu.Lock()
	eng.standby = &ProbeResult{
		RelayID:   "standby-live",
		Name:      "Standby Live",
		Endpoint:  srv2.UDPAddr().String(),
		Reachable: true,
	}
	eng.standbyAt = time.Now()
	eng.standbyMu.Unlock()

	dualRec := doGuardedPost(t, h, "/api/dual-path/toggle", map[string]bool{"enabled": true})
	if dualRec.Code != http.StatusOK {
		t.Fatalf("dual-path toggle failed: %d %s", dualRec.Code, dualRec.Body.String())
	}
	var dualResp struct {
		Enabled  bool            `json:"enabled"`
		Active   bool            `json:"active"`
		DualPath *DualPathStatus `json:"dualPath"`
	}
	if err := json.Unmarshal(dualRec.Body.Bytes(), &dualResp); err != nil {
		t.Fatalf("decode dual response: %v", err)
	}
	if !dualResp.Active || dualResp.DualPath == nil || dualResp.DualPath.StandbyRelayName != "Standby Live" {
		t.Fatalf("expected independent-PSK dual path active, got: %+v", dualResp)
	}
	waitActiveUsers(t, srv2.HTTPAddr(), 1)

	// Disconnect and verify both relays cleanly drop their sessions
	if err := eng.Disconnect(); err != nil {
		t.Fatalf("disconnect: %v", err)
	}
	waitActiveUsers(t, srv1.HTTPAddr(), 0)
	waitActiveUsers(t, srv2.HTTPAddr(), 0)

	// 5. Real WireGuard (wireguard-go) .conf import -> ConnectWireGuard -> Disconnect
	priv, _, err := generateWireGuardKeypair()
	if err != nil {
		t.Fatalf("generateWireGuardKeypair: %v", err)
	}
	_, peerPub, err := generateWireGuardKeypair()
	if err != nil {
		t.Fatalf("generateWireGuardKeypair peer: %v", err)
	}
	wgConf := "[Interface]\nPrivateKey = " + priv + "\nAddress = 172.16.0.2/32\nMTU = 1280\n[Peer]\nPublicKey = " + peerPub + "\nEndpoint = 127.0.0.1:51899\n"
	wgImp := doGuardedPost(t, h, "/api/relay/import-wireguard", map[string]string{
		"name":   "WG Local E2E",
		"config": wgConf,
	})
	if wgImp.Code != http.StatusOK {
		t.Fatalf("import-wireguard failed: %d %s", wgImp.Code, wgImp.Body.String())
	}
	wgConn := doGuardedPost(t, h, "/api/connect", map[string]any{
		"relayEndpoint": "127.0.0.1:51899",
		"gameId":        "cs2",
		"regionId":      "asia-sg",
	})
	if wgConn.Code != http.StatusOK {
		t.Fatalf("wireguard connect failed: %d %s", wgConn.Code, wgConn.Body.String())
	}
	wgStats := eng.Stats()
	if wgStats.State != StateConnected || wgStats.Transport != "wireguard" {
		t.Fatalf("expected connected wireguard transport, got state=%s transport=%s", wgStats.State, wgStats.Transport)
	}
	if err := eng.Disconnect(); err != nil {
		t.Fatalf("wireguard disconnect: %v", err)
	}

	// 6. Toggle Anti-DPI Obfuscation ON via HTTP API -> connect to srv1 -> verify real Pong
	obfsRec := doGuardedPost(t, h, "/api/obfuscation/toggle", map[string]bool{"enabled": true})
	if obfsRec.Code != http.StatusOK {
		t.Fatalf("obfuscation toggle failed: %d %s", obfsRec.Code, obfsRec.Body.String())
	}
	connObfs := doGuardedPost(t, h, "/api/connect", map[string]any{
		"relayEndpoint": srv1.UDPAddr().String(),
		"gameId":        "valorant",
		"regionId":      "asia-sg",
	})
	if connObfs.Code != http.StatusOK {
		t.Fatalf("obfuscated connect failed: %d %s", connObfs.Code, connObfs.Body.String())
	}
	deadline = time.Now().Add(6 * time.Second)
	for time.Now().Before(deadline) && !eng.Stats().PingSampled {
		time.Sleep(25 * time.Millisecond)
	}
	if !eng.Stats().PingSampled || !eng.Stats().Obfuscated {
		t.Fatalf("expected obfuscated live pong, got %+v", eng.Stats())
	}
	if err := eng.Disconnect(); err != nil {
		t.Fatalf("obfuscated disconnect: %v", err)
	}
	waitActiveUsers(t, srv1.HTTPAddr(), 0)

	// 7. Verify Reed-Solomon RS(4,3) bursts recover 3 dropped real IPv4 UDP packets through the relay's own FECDecoder
	enc := protocol.NewFECEncoder(protocol.FECEncoderConfig{DefaultBlockSize: 4, DefaultParityCount: 3})
	dec := protocol.NewFECDecoder(64)
	var orig [4][]byte
	var parities [][]byte
	for i := 0; i < 4; i++ {
		pkt := make([]byte, 48+i*8)
		pkt[0] = 0x45
		pkt[9] = 17 // UDP
		copy(pkt[12:16], netip.MustParseAddr("10.77.0.2").AsSlice())
		copy(pkt[16:20], netip.MustParseAddr("1.1.1.1").AsSlice())
		pkt[20] = byte(i + 1)
		orig[i] = pkt
		parities = append(parities, enc.AddPacketMulti(uint64(500+i), pkt)...)
	}
	// Keep ONLY packet 0, drop packets 1, 2, 3 (75% consecutive burst loss!)
	dec.RecordPacket(500, orig[0])
	recovered := map[uint64][]byte{}
	for _, rawP := range parities {
		fp, err := protocol.DecodeFECPayload(rawP)
		if err != nil {
			t.Fatalf("DecodeFECPayload: %v", err)
		}
		for _, r := range dec.ProcessParityMulti(fp) {
			recovered[r.Seq] = r.Data
		}
	}
	for i := 1; i <= 3; i++ {
		got, ok := recovered[uint64(500+i)]
		if !ok || !bytes.Equal(got, orig[i]) {
			t.Fatalf("RS(4,3) failed to recover burst-lost packet %d", i)
		}
	}
}
