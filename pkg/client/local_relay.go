package client

import (
	"context"
	"log"
	"net"
	"net/netip"
	"sync"
	"time"

	"github.com/ThanhNguyxnOrg/lagvex/pkg/relay"
)

var (
	localRelayMu     sync.Mutex
	localRelayServer *relay.Server
	localRelayCancel context.CancelFunc
	localRelayActive bool
)

// EnsureLocalRelayRunning checks if a relay is already listening on listenAddr (e.g. "127.0.0.1:4433").
// If no server is listening, it launches an embedded in-process Relay server using the userspace TUN device.
// This allows gamers to run and test acceleration instantly without needing a remote VPS.
func EnsureLocalRelayRunning(listenAddr string, psk []byte) error {
	localRelayMu.Lock()
	defer localRelayMu.Unlock()

	if localRelayActive {
		return nil
	}

	if listenAddr == "" {
		listenAddr = "127.0.0.1:4433"
	}
	if len(psk) == 0 {
		psk = []byte("lagvex-community-us-free-public-psk-2026")
	}

	// 1. Check if an external relay (or previous instance) is already listening
	laddr, err := net.ResolveUDPAddr("udp4", listenAddr)
	if err == nil {
		testConn, testErr := net.ListenUDP("udp4", laddr)
		if testErr != nil {
			// Port is in use - either another relay or our server is already active!
			log.Printf("[LocalRelay] Port %s is already bound by an active relay. Yielding to existing instance.", listenAddr)
			localRelayActive = true
			return nil
		}
		_ = testConn.Close()
	}

	// 2. Instantiate and launch embedded Relay server
	cfg := relay.Config{
		ListenAddr:  listenAddr,
		TunName:     "lagvex-local",
		Subnet:      netip.MustParsePrefix("10.88.0.0/24"),
		PSK:         psk,
		MTU:         1400,
		IdleTimeout: 90 * time.Second,
		MaxClients:  128,
	}

	server, err := relay.NewServer(cfg)
	if err != nil {
		log.Printf("[LocalRelay] Warning: failed to initialize embedded relay: %v", err)
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	localRelayServer = server
	localRelayCancel = cancel
	localRelayActive = true

	go func() {
		log.Printf("[LocalRelay] Starting embedded low-latency relay on %s (Zero-VPS Gamer Mode)...", listenAddr)
		if err := server.Start(ctx); err != nil && ctx.Err() == nil {
			log.Printf("[LocalRelay] Embedded relay exited: %v", err)
			localRelayMu.Lock()
			localRelayActive = false
			localRelayMu.Unlock()
		}
	}()

	// Wait briefly for socket to bind
	time.Sleep(100 * time.Millisecond)
	return nil
}

// StopLocalRelay stops the embedded relay server if running.
func StopLocalRelay() {
	localRelayMu.Lock()
	defer localRelayMu.Unlock()

	if localRelayCancel != nil {
		localRelayCancel()
		localRelayCancel = nil
	}
	localRelayActive = false
	localRelayServer = nil
}
