package relay

import (
	"errors"
	"net/netip"
	"sync"
	"time"
)

var (
	ErrPoolExhausted = errors.New("address pool exhausted")
)

// IPPool manages dynamic client IPv4 allocations within a subnet.
type IPPool struct {
	mu           sync.Mutex
	subnet       netip.Prefix
	gatewayIP    netip.Addr
	startIP      netip.Addr
	endIP        netip.Addr
	usedIPs      map[netip.Addr]uint64   // IP -> SessionID
	reservations map[uint64]reservedAddr // ClientID -> reserved IP info
	reservedIPs  map[netip.Addr]uint64   // IP -> ClientID (for O(1) reservation check)
}

type reservedAddr struct {
	ip        netip.Addr
	expiresAt time.Time
}

// NewIPPool creates an IPPool for a given prefix (e.g. 10.88.0.0/24).
// First usable IP (.1) is assigned as the Gateway. The remaining IPs (.2..254) are allocated to clients.
func NewIPPool(subnet netip.Prefix) (*IPPool, error) {
	if !subnet.IsValid() || !subnet.Addr().Is4() {
		return nil, errors.New("invalid IPv4 subnet prefix")
	}

	gw := subnet.Addr().Next() // e.g. 10.88.0.1
	firstClient := gw.Next()   // e.g. 10.88.0.2

	// Compute broadcast address
	addrSlice := subnet.Addr().As4()
	mask := ^uint32(0) << (32 - subnet.Bits())
	ipInt := (uint32(addrSlice[0]) << 24) | (uint32(addrSlice[1]) << 16) | (uint32(addrSlice[2]) << 8) | uint32(addrSlice[3])
	bcastInt := ipInt | ^mask
	lastClientInt := bcastInt - 1 // 1 before broadcast

	lastClient := netip.AddrFrom4([4]byte{
		byte(lastClientInt >> 24),
		byte(lastClientInt >> 16),
		byte(lastClientInt >> 8),
		byte(lastClientInt),
	})

	return &IPPool{
		subnet:       subnet,
		gatewayIP:    gw,
		startIP:      firstClient,
		endIP:        lastClient,
		usedIPs:      make(map[netip.Addr]uint64),
		reservations: make(map[uint64]reservedAddr),
		reservedIPs:  make(map[netip.Addr]uint64),
	}, nil
}

// Gateway returns the gateway IP.
func (p *IPPool) Gateway() netip.Addr {
	return p.gatewayIP
}

// Allocate assigns an IP address for a session.
// If the clientID has a valid existing reservation, it re-allocates that address.
func (p *IPPool) Allocate(clientID, sessionID uint64) (netip.Addr, bool, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()

	// 1. Check reservation for this client
	if res, ok := p.reservations[clientID]; ok && now.Before(res.expiresAt) {
		// If not currently used by someone else
		if existingSess, inUse := p.usedIPs[res.ip]; !inUse || existingSess == sessionID {
			p.usedIPs[res.ip] = sessionID
			return res.ip, true, nil
		}
	}

	bindIP := func(ip netip.Addr) (netip.Addr, bool, error) {
		p.usedIPs[ip] = sessionID
		if oldRes, ok := p.reservations[clientID]; ok {
			delete(p.reservedIPs, oldRes.ip)
		}
		if prevClient, ok := p.reservedIPs[ip]; ok && prevClient != clientID {
			delete(p.reservations, prevClient)
		}
		p.reservations[clientID] = reservedAddr{
			ip:        ip,
			expiresAt: now.Add(24 * time.Hour),
		}
		p.reservedIPs[ip] = clientID
		return ip, false, nil
	}

	// 2. Pass 1: Scan for an available IP that is unused AND unreserved (or reservation expired)
	curr := p.startIP
	for {
		if _, inUse := p.usedIPs[curr]; !inUse {
			resClient, isReserved := p.reservedIPs[curr]
			if !isReserved || resClient == clientID {
				return bindIP(curr)
			}
			if r, ok := p.reservations[resClient]; !ok || !now.Before(r.expiresAt) {
				return bindIP(curr)
			}
		}
		if curr == p.endIP {
			break
		}
		curr = curr.Next()
	}

	// 3. Pass 2: Pool exhausted of unreserved IPs. Reclaim an unused IP even if reserved by another client.
	curr = p.startIP
	for {
		if _, inUse := p.usedIPs[curr]; !inUse {
			return bindIP(curr)
		}
		if curr == p.endIP {
			break
		}
		curr = curr.Next()
	}

	return netip.Addr{}, false, ErrPoolExhausted
}

// Release frees an IP allocated to a session and optionally keeps a reservation for clientID.
func (p *IPPool) Release(ip netip.Addr, clientID uint64, keepReservation bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.usedIPs, ip)
	if !keepReservation {
		delete(p.reservations, clientID)
		if curClient, ok := p.reservedIPs[ip]; ok && curClient == clientID {
			delete(p.reservedIPs, ip)
		}
	}
}
