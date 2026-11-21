package client

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/netip"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ThanhNguyxnOrg/lagvex/pkg/profiles"
	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/device"
	wgtun "golang.zx2c4.com/wireguard/tun"
)

// wgTunBridge adapts Lagvex's cross-platform WintunAdapter (WinTun on Windows,
// utun on macOS, /dev/net/tun on Linux) to wireguard-go's tun.Device interface.
// When adapter is nil (userspace fallback without root/admin privileges), reads
// block until close and writes are discarded so the WireGuard handshake and
// diagnostics still run cleanly without panicking.
type wgTunBridge struct {
	adapter   *WintunAdapter
	name      string
	mtu       int
	events    chan wgtun.Event
	closeCh   chan struct{}
	closeOnce sync.Once
	bytesUp   *atomic.Uint64
	bytesDown *atomic.Uint64
}

func newWGTunBridge(adapter *WintunAdapter, mtu int, bytesUp, bytesDown *atomic.Uint64) *wgTunBridge {
	if mtu < 1200 || mtu > 1500 {
		mtu = 1280
	}
	ch := make(chan wgtun.Event, 4)
	ch <- wgtun.EventUp
	return &wgTunBridge{
		adapter:   adapter,
		name:      "LagvexWG",
		mtu:       mtu,
		events:    ch,
		closeCh:   make(chan struct{}),
		bytesUp:   bytesUp,
		bytesDown: bytesDown,
	}
}

func (b *wgTunBridge) File() *os.File { return nil }

func (b *wgTunBridge) BatchSize() int { return 1 }

func (b *wgTunBridge) MTU() (int, error) { return b.mtu, nil }

func (b *wgTunBridge) Name() (string, error) { return b.name, nil }

func (b *wgTunBridge) Events() <-chan wgtun.Event { return b.events }

func (b *wgTunBridge) Read(bufs [][]byte, sizes []int, offset int) (int, error) {
	if len(bufs) == 0 || len(sizes) == 0 {
		return 0, nil
	}
	if b.adapter == nil {
		<-b.closeCh
		return 0, os.ErrClosed
	}

	target := bufs[0]
	if offset >= len(target) {
		return 0, io.ErrShortBuffer
	}

	n, err := b.adapter.ReadPacket(target[offset:])
	if err != nil {
		select {
		case <-b.closeCh:
			return 0, os.ErrClosed
		default:
		}
		if strings.Contains(err.Error(), "closed") {
			return 0, os.ErrClosed
		}
		return 0, err
	}
	if n <= 0 {
		return 0, nil
	}
	sizes[0] = n
	if b.bytesUp != nil {
		b.bytesUp.Add(uint64(n))
	}
	return 1, nil
}

func (b *wgTunBridge) Write(bufs [][]byte, offset int) (int, error) {
	if b.adapter == nil {
		return len(bufs), nil
	}
	written := 0
	for _, buf := range bufs {
		if offset >= len(buf) {
			continue
		}
		pkt := buf[offset:]
		if len(pkt) < 20 {
			continue
		}
		if err := b.adapter.WritePacket(pkt); err != nil {
			select {
			case <-b.closeCh:
				return written, os.ErrClosed
			default:
			}
			return written, err
		}
		if b.bytesDown != nil {
			b.bytesDown.Add(uint64(len(pkt)))
		}
		written++
	}
	return written, nil
}

func (b *wgTunBridge) Close() error {
	b.closeOnce.Do(func() {
		close(b.closeCh)
		close(b.events)
	})
	return nil
}

// wgTunnelStats holds real counters read from wireguard-go's UAPI.
type wgTunnelStats struct {
	LastHandshake time.Time
	TxBytes       uint64
	RxBytes       uint64
}

// wireguardTunnel wraps a live userspace wireguard-go Device.
type wireguardTunnel struct {
	dev      *device.Device
	bridge   *wgTunBridge
	endpoint netip.AddrPort
}

// buildWireGuardUAPI converts a validated WireGuardConfig and resolved peer
// endpoint into the line-oriented UAPI configuration accepted by device.IpcSet.
func buildWireGuardUAPI(cfg *profiles.WireGuardConfig, endpoint netip.AddrPort) (string, error) {
	privHex, err := profiles.DecodeWireGuardKeyHex(cfg.PrivateKey)
	if err != nil {
		return "", fmt.Errorf("decode private key: %w", err)
	}
	pubHex, err := profiles.DecodeWireGuardKeyHex(cfg.PublicKey)
	if err != nil {
		return "", fmt.Errorf("decode public key: %w", err)
	}

	var sb strings.Builder
	sb.WriteString("private_key=")
	sb.WriteString(privHex)
	sb.WriteByte('\n')
	sb.WriteString("replace_peers=true\n")
	sb.WriteString("public_key=")
	sb.WriteString(pubHex)
	sb.WriteByte('\n')

	if strings.TrimSpace(cfg.PresharedKey) != "" {
		pskHex, err := profiles.DecodeWireGuardKeyHex(cfg.PresharedKey)
		if err != nil {
			return "", fmt.Errorf("decode preshared key: %w", err)
		}
		sb.WriteString("preshared_key=")
		sb.WriteString(pskHex)
		sb.WriteByte('\n')
	}

	sb.WriteString("endpoint=")
	sb.WriteString(endpoint.String())
	sb.WriteByte('\n')
	// Split-tunneling is enforced by the OS RouteManager (only game CIDRs and
	// optional peer /32s enter the TUN interface), so WireGuard cryptokey
	// routing accepts any IPv4 game destination routed into the adapter.
	sb.WriteString("allowed_ip=0.0.0.0/0\n")

	keepalive := cfg.PersistentKeepalive
	if keepalive <= 0 {
		keepalive = 25
	}
	sb.WriteString("persistent_keepalive_interval=")
	sb.WriteString(strconv.Itoa(keepalive))
	sb.WriteByte('\n')

	return sb.String(), nil
}

func startWireGuardTunnel(adapter *WintunAdapter, cfg *profiles.WireGuardConfig, endpoint netip.AddrPort, bytesUp, bytesDown *atomic.Uint64) (*wireguardTunnel, error) {
	uapi, err := buildWireGuardUAPI(cfg, endpoint)
	if err != nil {
		return nil, err
	}

	mtu := cfg.MTU
	if mtu < 1200 || mtu > 1500 {
		mtu = 1280
	}

	bridge := newWGTunBridge(adapter, mtu, bytesUp, bytesDown)
	logger := device.NewLogger(device.LogLevelError, "[WireGuard] ")
	dev := device.NewDevice(bridge, conn.NewDefaultBind(), logger)

	if err := dev.IpcSet(uapi); err != nil {
		dev.Close()
		_ = bridge.Close()
		return nil, fmt.Errorf("configure wireguard device: %w", err)
	}
	if err := dev.Up(); err != nil {
		dev.Close()
		_ = bridge.Close()
		return nil, fmt.Errorf("bring wireguard device up: %w", err)
	}

	return &wireguardTunnel{
		dev:      dev,
		bridge:   bridge,
		endpoint: endpoint,
	}, nil
}

func (w *wireguardTunnel) Stats() wgTunnelStats {
	if w == nil || w.dev == nil {
		return wgTunnelStats{}
	}
	raw, err := w.dev.IpcGet()
	if err != nil {
		return wgTunnelStats{}
	}
	var st wgTunnelStats
	scanner := bufio.NewScanner(strings.NewReader(raw))
	for scanner.Scan() {
		k, v, ok := strings.Cut(scanner.Text(), "=")
		if !ok {
			continue
		}
		switch k {
		case "last_handshake_time_sec":
			if sec, err := strconv.ParseInt(v, 10, 64); err == nil && sec > 0 {
				st.LastHandshake = time.Unix(sec, 0)
			}
		case "tx_bytes":
			if n, err := strconv.ParseUint(v, 10, 64); err == nil {
				st.TxBytes = n
			}
		case "rx_bytes":
			if n, err := strconv.ParseUint(v, 10, 64); err == nil {
				st.RxBytes = n
			}
		}
	}
	return st
}

func (w *wireguardTunnel) Close() {
	if w == nil {
		return
	}
	if w.bridge != nil {
		_ = w.bridge.Close()
	}
	if w.dev != nil {
		w.dev.Close()
	}
}

// measureWireGuardEndpointRTT measures honest reachability RTT to a WireGuard
// endpoint host using ICMP echo (when permitted) or TCP handshake to port 443.
// Returns (0, false) when neither can be measured — never invents a number.
func measureWireGuardEndpointRTT(ctx context.Context, ip net.IP) (int64, bool) {
	if ip == nil {
		return 0, false
	}
	if res, err := probeICMP(ctx, ip, 1, 700*time.Millisecond); err == nil && res.Samples > 0 && res.RTTMs > 0 {
		return int64(res.RTTMs + 0.5), true
	}
	// Fallback: TCP connect timing to port 443 on the endpoint IP (e.g.
	// Cloudflare WARP edge IPs answer TLS/RST on 443).
	dialCtx, cancel := context.WithTimeout(ctx, 700*time.Millisecond)
	defer cancel()
	rtts, err := probeTCPGateway(dialCtx, net.JoinHostPort(ip.String(), "443"), 1)
	if err == nil && len(rtts) > 0 {
		return int64(rtts[0] + 0.5), true
	}
	if errors.Is(err, context.Canceled) {
		return 0, false
	}
	return 0, false
}

// loopWireGuardMonitor watches a live WireGuard tunnel, updating real byte
// telemetry and measuring honest endpoint RTT while connected.
func (e *Engine) loopWireGuardMonitor(ctx context.Context, wgTun *wireguardTunnel, endpointIP net.IP) {
	defer e.wg.Done()
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	probeOnce := func() {
		probeCtx, cancel := context.WithTimeout(ctx, 1500*time.Millisecond)
		rttMs, ok := measureWireGuardEndpointRTT(probeCtx, endpointIP)
		cancel()
		if ok {
			e.pingMs.Store(rttMs)
			e.pingSampled.Store(true)
			e.lastPong.Store(time.Now().UnixNano())
			e.pongTotal.Add(1)
		}
		st := wgTun.Stats()
		if !st.LastHandshake.IsZero() {
			// A completed Noise_IKpsk2 handshake confirms the peer is alive.
			if time.Since(st.LastHandshake) < 3*time.Minute && !ok {
				e.lastPong.Store(time.Now().UnixNano())
			}
		}
	}

	probeOnce()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			probeOnce()
			log.Printf("[WireGuard] Monitor tick: ping=%dms sampled=%v up=%dB down=%dB",
				e.pingMs.Load(), e.pingSampled.Load(), e.bytesUp.Load(), e.bytesDown.Load())
		}
	}
}
