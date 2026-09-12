package client

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/netip"
	"sync"
	"sync/atomic"
	"time"

	"github.com/lagvex/lagvex/pkg/profiles"
	"github.com/lagvex/lagvex/pkg/protocol"
)

// TunnelState represents the operational status of the tunnel.
type TunnelState string

const (
	StateDisconnected TunnelState = "disconnected"
	StateConnecting   TunnelState = "connecting"
	StateConnected    TunnelState = "connected"
	StateDisconnecting TunnelState = "disconnecting"
)

// TunnelStats holds real-time telemetry.
type TunnelStats struct {
	State        TunnelState `json:"state"`
	RelayAddr    string      `json:"relayAddr"`
	ClientIP     string      `json:"clientIP"`
	GatewayIP    string      `json:"gatewayIP"`
	PingMs       int64       `json:"pingMs"`
	BytesUp      uint64      `json:"bytesUp"`
	BytesDown    uint64      `json:"bytesDown"`
	UpRateBps    int64       `json:"upRateBps"`
	DownRateBps  int64       `json:"downRateBps"`
	ActiveGame   string      `json:"activeGame"`
	ActiveRegion string      `json:"activeRegion"`
	RouteCount   int         `json:"routeCount"`
	GameRunning  bool        `json:"gameRunning"`
}

// Engine orchestrates the client VPN tunnel and routing lifecycle.
type Engine struct {
	mu           sync.Mutex
	state        TunnelState
	clientID     uint64
	sessionID    uint64
	clientIP     netip.Addr
	gatewayIP    netip.Addr
	mtu          int

	relayAddr    netip.AddrPort
	relayIP      netip.Addr
	psk          []byte
	udpConn      *net.UDPConn
	wintun       *WintunAdapter
	routeManager *RouteManager
	profileMgr   *profiles.Manager
	watcher      *ProcessWatcher

	activeGame   string
	activeRegion string
	gameCIDRs    []string
	isGameActive bool

	cancelFunc   context.CancelFunc
	wg           sync.WaitGroup

	bytesUp      atomic.Uint64
	bytesDown    atomic.Uint64
	pingMs       atomic.Int64
	upRate       atomic.Int64
	downRate     atomic.Int64
}

// NewEngine initializes the Lagvex Client Engine.
func NewEngine(pm *profiles.Manager) (*Engine, error) {
	clientID, err := protocol.RandomUint64()
	if err != nil {
		return nil, err
	}

	return &Engine{
		state:        StateDisconnected,
		clientID:     clientID,
		routeManager: NewRouteManager(),
		profileMgr:   pm,
	}, nil
}

// Stats returns a snapshot of current telemetry.
func (e *Engine) Stats() TunnelStats {
	e.mu.Lock()
	defer e.mu.Unlock()

	return TunnelStats{
		State:        e.state,
		RelayAddr:    e.relayAddr.String(),
		ClientIP:     e.clientIP.String(),
		GatewayIP:    e.gatewayIP.String(),
		PingMs:       e.pingMs.Load(),
		BytesUp:      e.bytesUp.Load(),
		BytesDown:    e.bytesDown.Load(),
		UpRateBps:    e.upRate.Load(),
		DownRateBps:  e.downRate.Load(),
		ActiveGame:   e.activeGame,
		ActiveRegion: e.activeRegion,
		RouteCount:   e.routeManager.ActiveRouteCount(),
		GameRunning:  e.isGameActive,
	}
}

// Connect initiates connection to the relay and configures adapter & routing.
func (e *Engine) Connect(relayEndpoint string, psk []byte, gameID, regionID string, forceRoutesNow bool) error {
	e.mu.Lock()
	if e.state != StateDisconnected {
		e.mu.Unlock()
		return fmt.Errorf("cannot connect in state %s", e.state)
	}
	e.state = StateConnecting
	e.mu.Unlock()

	log.Printf("[Engine] Connecting to relay %s (Game: %s, Region: %s)...", relayEndpoint, gameID, regionID)

	// 1. Resolve Relay endpoint
	udpAddr, err := net.ResolveUDPAddr("udp4", relayEndpoint)
	if err != nil {
		e.resetState()
		return fmt.Errorf("resolve relay endpoint %s: %w", relayEndpoint, err)
	}
	relayAddrPort := udpAddr.AddrPort()
	relayIP := relayAddrPort.Addr()

	// 2. Open UDP socket
	conn, err := net.ListenUDP("udp4", nil)
	if err != nil {
		e.resetState()
		return fmt.Errorf("create udp socket: %w", err)
	}
	_ = conn.SetReadBuffer(8 * 1024 * 1024)
	_ = conn.SetWriteBuffer(8 * 1024 * 1024)

	// 3. Handshake
	nonce, _ := protocol.RandomUint64()
	req := protocol.HandshakeRequest{
		Nonce:     nonce,
		Timestamp: time.Now().Unix(),
		ClientID:  e.clientID,
	}
	reqBuf := protocol.EncodeHandshakeRequest(psk, req)

	// Retry handshake up to 3 times
	var resp *protocol.HandshakeResponse
	recvBuf := make([]byte, 256)

	for attempt := 1; attempt <= 3; attempt++ {
		_, _ = conn.WriteToUDP(reqBuf, udpAddr)
		_ = conn.SetReadDeadline(time.Now().Add(2500 * time.Millisecond))

		n, _, readErr := conn.ReadFromUDP(recvBuf)
		if readErr == nil {
			var decErr error
			resp, decErr = protocol.DecodeHandshakeResponse(psk, recvBuf[:n], nonce)
			if decErr == nil && resp.Status == protocol.StatusOK {
				break
			}
		}
		log.Printf("[Engine] Handshake attempt %d failed, retrying...", attempt)
	}

	_ = conn.SetReadDeadline(time.Time{})

	if resp == nil || resp.Status != protocol.StatusOK {
		conn.Close()
		e.resetState()
		return fmt.Errorf("handshake failed with relay (check PSK or server availability)")
	}

	log.Printf("[Engine] Handshake OK: session=%016x assignedIP=%s gw=%s mtu=%d",
		resp.SessionID, resp.ClientIP, resp.GatewayIP, resp.MTU)

	// 4. Create WinTun adapter
	adapter, err := OpenOrCreateWintunAdapter("Lagvex", "LagvexTunnel", "")
	if err != nil {
		conn.Close()
		e.resetState()
		return fmt.Errorf("initialize WinTun adapter: %w (ensure run as Administrator)", err)
	}

	// 5. Configure Adapter IP & MTU
	if err := e.routeManager.ConfigureAdapter(adapter.InterfaceIndex(), resp.ClientIP, 24, int(resp.MTU)); err != nil {
		log.Printf("[Engine] Warning: adapter configuration: %v", err)
	}

	// 6. Pin Relay route to prevent routing loops
	if err := e.routeManager.PinRelayRoute(relayIP); err != nil {
		log.Printf("[Engine] Warning: pin relay route: %v", err)
	}

	// 7. Resolve Game CIDRs
	var cidrs []string
	if gameID != "" && regionID != "" {
		resolved, err := e.profileMgr.GetRegionCIDRs(gameID, regionID)
		if err == nil {
			cidrs = resolved
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	e.mu.Lock()
	e.state = StateConnected
	e.sessionID = resp.SessionID
	e.clientIP = resp.ClientIP
	e.gatewayIP = resp.GatewayIP
	e.mtu = int(resp.MTU)
	e.relayAddr = relayAddrPort
	e.relayIP = relayIP
	e.psk = psk
	e.udpConn = conn
	e.wintun = adapter
	e.activeGame = gameID
	e.activeRegion = regionID
	e.gameCIDRs = cidrs
	e.cancelFunc = cancel
	e.mu.Unlock()

	// If manual mode, install routes right away
	if forceRoutesNow && len(cidrs) > 0 {
		e.routeManager.InstallGameRoutes(cidrs)
	}

	// 8. Start I/O pumps
	e.wg.Add(4)
	go e.pumpWinTunToUDP(ctx)
	go e.pumpUDPToWinTun(ctx)
	go e.loopKeepalive(ctx)
	go e.loopBitrateMeter(ctx)

	// 9. Start Process Watcher
	if gameID != "" {
		if game, ok := e.profileMgr.FindGameByID(gameID); ok && len(game.ProcessNames) > 0 {
			e.watcher = NewProcessWatcher(game.ProcessNames)
			e.watcher.OnGameStarted = func(proc string) {
				log.Printf("[Engine] Game process %q started! Activating game routes...", proc)
				e.mu.Lock()
				e.isGameActive = true
				cList := e.gameCIDRs
				e.mu.Unlock()
				e.routeManager.InstallGameRoutes(cList)
			}
			e.watcher.OnGameStopped = func(proc string) {
				log.Printf("[Engine] Game process %q exited. Removing game routes...", proc)
				e.mu.Lock()
				e.isGameActive = false
				e.mu.Unlock()
				e.routeManager.RemoveGameRoutes()
			}
			go e.watcher.Start(ctx, 2*time.Second)
		}
	}

	log.Printf("[Engine] Tunnel active and accelerating!")
	return nil
}

// Disconnect stops the tunnel and cleans up all routes.
func (e *Engine) Disconnect() error {
	e.mu.Lock()
	if e.state != StateConnected && e.state != StateConnecting {
		e.mu.Unlock()
		return nil
	}
	e.state = StateDisconnecting
	cancel := e.cancelFunc
	sessID := e.sessionID
	conn := e.udpConn
	rAddr := e.relayAddr
	adapter := e.wintun
	e.mu.Unlock()

	log.Printf("[Engine] Disconnecting tunnel...")

	// Send Disconnect message to relay
	if conn != nil {
		discBuf := protocol.EncodeDisconnect(sessID)
		udpTarget := net.UDPAddrFromAddrPort(rAddr)
		_, _ = conn.WriteToUDP(discBuf, udpTarget)
	}

	if cancel != nil {
		cancel()
	}

	if adapter != nil {
		_ = adapter.Close()
	}
	if conn != nil {
		_ = conn.Close()
	}

	e.wg.Wait()

	// Clean up routes
	e.routeManager.RemoveAll()

	e.mu.Lock()
	e.state = StateDisconnected
	e.isGameActive = false
	e.activeGame = ""
	e.activeRegion = ""
	e.pingMs.Store(0)
	e.mu.Unlock()

	log.Printf("[Engine] Tunnel disconnected and routes cleaned.")
	return nil
}

func (e *Engine) pumpWinTunToUDP(ctx context.Context) {
	defer e.wg.Done()
	inBuf := make([]byte, protocol.MaxPacketSize)
	outBuf := make([]byte, protocol.MaxPacketSize)
	udpTarget := net.UDPAddrFromAddrPort(e.relayAddr)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		n, err := e.wintun.ReadPacket(inBuf)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			continue
		}

		if n < 20 || (inBuf[0]>>4) != 4 {
			continue // Skip non-IPv4
		}

		pkt := protocol.EncodeDataPacket(outBuf, e.sessionID, inBuf[:n])
		_, err = e.udpConn.WriteToUDP(pkt, udpTarget)
		if err == nil {
			e.bytesUp.Add(uint64(n))
		}
	}
}

func (e *Engine) pumpUDPToWinTun(ctx context.Context) {
	defer e.wg.Done()
	inBuf := make([]byte, protocol.MaxPacketSize)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		n, _, err := e.udpConn.ReadFromUDP(inBuf)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			continue
		}

		if n < 1 {
			continue
		}

		ver, mtype := protocol.ParseHeader(inBuf[0])
		if ver != protocol.Version {
			continue
		}

		switch mtype {
		case protocol.TypeData:
			_, payload, err := protocol.DecodeDataPacket(inBuf[:n])
			if err == nil && len(payload) > 0 {
				_ = e.wintun.WritePacket(payload)
				e.bytesDown.Add(uint64(len(payload)))
			}

		case protocol.TypePong:
			_, sentTs, err := protocol.DecodePong(inBuf[:n])
			if err == nil {
				rtt := time.Since(time.Unix(0, int64(sentTs)))
				e.pingMs.Store(rtt.Milliseconds())
			}
		}
	}
}

func (e *Engine) loopKeepalive(ctx context.Context) {
	defer e.wg.Done()
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	udpTarget := net.UDPAddrFromAddrPort(e.relayAddr)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pingBuf := protocol.EncodePing(e.sessionID, uint64(time.Now().UnixNano()))
			_, _ = e.udpConn.WriteToUDP(pingBuf, udpTarget)
		}
	}
}

func (e *Engine) loopBitrateMeter(ctx context.Context) {
	defer e.wg.Done()
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	var lastUp, lastDown uint64

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			currUp := e.bytesUp.Load()
			currDown := e.bytesDown.Load()

			upDiff := int64(currUp - lastUp)
			downDiff := int64(currDown - lastDown)

			e.upRate.Store(upDiff)
			e.downRate.Store(downDiff)

			lastUp = currUp
			lastDown = currDown
		}
	}
}

func (e *Engine) resetState() {
	e.mu.Lock()
	e.state = StateDisconnected
	e.mu.Unlock()
}
