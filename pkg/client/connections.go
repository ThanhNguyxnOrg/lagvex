package client

import (
	"bufio"
	"encoding/hex"
	"io"
	"net"
	"net/netip"
	"sort"
	"strconv"
	"strings"
)

// ProcessConnection is one network socket observed for a running process.
// All fields come from the local OS socket tables; nothing is inferred from
// packet payloads.
type ProcessConnection struct {
	Protocol string `json:"protocol"` // "udp" or "tcp"
	Family   string `json:"family"`   // "ipv4" or "ipv6"
	Local    string `json:"local"`
	Remote   string `json:"remote"`
	State    string `json:"state,omitempty"`
}

// mergeConnections dedupes sockets by remote endpoint and sorts the result so
// a process holding several sockets to the same server lists it once.
func mergeConnections(existing, extra []ProcessConnection) []ProcessConnection {
	seen := make(map[string]struct{}, len(existing)+len(extra))
	out := make([]ProcessConnection, 0, len(existing)+len(extra))
	for _, list := range [][]ProcessConnection{existing, extra} {
		for _, conn := range list {
			// Dedupe by remote peer, falling back to the local endpoint for
			// sockets whose OS table has no peer (for example UDP on Windows).
			endpoint := conn.Remote
			if endpoint == "" {
				endpoint = conn.Local
			}
			key := conn.Protocol + "|" + endpoint
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, conn)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Remote != out[j].Remote {
			return out[i].Remote < out[j].Remote
		}
		return out[i].Protocol < out[j].Protocol
	})
	return out
}

// filterPublicPeerIPs keeps remote addresses that are safe to route: public,
// global-unicast IPv4 endpoints held by the observed process. Private,
// loopback, link-local, multicast, CGNAT and IPv6 addresses are skipped.
func filterPublicPeerIPs(conns []ProcessConnection) []string {
	seen := make(map[string]struct{}, len(conns))
	out := make([]string, 0, len(conns))
	for _, conn := range conns {
		host, _, err := net.SplitHostPort(conn.Remote)
		if err != nil {
			continue
		}
		ip, err := netip.ParseAddr(host)
		if err != nil || !ip.Is4() {
			continue
		}
		if !ip.IsGlobalUnicast() || ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() {
			continue
		}
		// Exclude carrier-grade NAT space (100.64.0.0/10): never a game server.
		octets := ip.As4()
		if octets[0] == 100 && octets[1] >= 64 && octets[1] <= 127 {
			continue
		}
		text := ip.String()
		if _, ok := seen[text]; ok {
			continue
		}
		seen[text] = struct{}{}
		out = append(out, text)
	}
	sort.Strings(out)
	return out
}

// decodeProcNetAddress decodes the hex address:port form used by /proc/net/*.
// IPv4 addresses are four little-endian bytes; IPv6 addresses are four
// little-endian 32-bit words (as printed by the Linux kernel).
func decodeProcNetAddress(raw string, ipv6 bool) (string, bool) {
	addrHex, portHex, found := strings.Cut(raw, ":")
	if !found {
		return "", false
	}
	port, err := strconv.ParseUint(portHex, 16, 32)
	if err != nil {
		return "", false
	}
	addrBytes, err := hex.DecodeString(addrHex)
	if err != nil {
		return "", false
	}

	if ipv6 {
		if len(addrBytes) != 16 {
			return "", false
		}
		ip := make(net.IP, net.IPv6len)
		for word := 0; word < 4; word++ {
			for b := 0; b < 4; b++ {
				ip[word*4+b] = addrBytes[word*4+(3-b)]
			}
		}
		return net.JoinHostPort(ip.String(), strconv.Itoa(int(port))), true
	}

	if len(addrBytes) != 4 {
		return "", false
	}
	ip := net.IPv4(addrBytes[3], addrBytes[2], addrBytes[1], addrBytes[0])
	return net.JoinHostPort(ip.String(), strconv.Itoa(int(port))), true
}

// procNetTCPStates maps the hex connection state used by /proc/net/tcp.
var procNetTCPStates = map[string]string{
	"01": "established",
	"02": "syn-sent",
	"03": "syn-recv",
	"04": "fin-wait-1",
	"05": "fin-wait-2",
	"06": "time-wait",
	"07": "close",
	"08": "close-wait",
	"09": "last-ack",
	"0A": "listen",
	"0B": "closing",
}

// parseProcNetTable parses one /proc/net/{tcp,udp}{,6} table. The returned map
// is keyed by socket inode. Rows without a remote peer are skipped because
// they cannot identify a game server.
func parseProcNetTable(r io.Reader, protocol string, ipv6 bool) map[string]ProcessConnection {
	result := make(map[string]ProcessConnection)
	scanner := bufio.NewScanner(r)
	first := true
	for scanner.Scan() {
		if first {
			first = false // header line
			continue
		}
		fields := strings.Fields(strings.TrimSpace(scanner.Text()))
		if len(fields) < 10 {
			continue
		}
		local, ok := decodeProcNetAddress(fields[1], ipv6)
		if !ok {
			continue
		}
		remote, ok := decodeProcNetAddress(fields[2], ipv6)
		if !ok || isUnspecifiedRemote(remote) {
			continue
		}
		inode := fields[9]
		if inode == "" || inode == "0" {
			continue
		}

		family := "ipv4"
		if ipv6 {
			family = "ipv6"
		}
		conn := ProcessConnection{
			Protocol: protocol,
			Family:   family,
			Local:    local,
			Remote:   remote,
		}
		if protocol == "tcp" {
			conn.State = procNetTCPStates[strings.ToUpper(fields[3])]
		}
		result[inode] = conn
	}
	return result
}

// isUnspecifiedRemote reports whether a decoded remote endpoint has no peer.
func isUnspecifiedRemote(remote string) bool {
	host, _, err := net.SplitHostPort(remote)
	if err != nil {
		return true
	}
	ip := net.ParseIP(host)
	return ip == nil || ip.IsUnspecified()
}

// parseLsofConnections extracts connected sockets from `lsof -nP -i` output.
// lsof prints NAME as "local->remote" for connected sockets; listening or
// unconnected sockets have no arrow and are skipped.
func parseLsofConnections(output string) []ProcessConnection {
	var conns []ProcessConnection
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 9 {
			continue
		}
		name := fields[8]
		local, remote, found := strings.Cut(name, "->")
		if !found {
			continue
		}

		protocol := ""
		for _, field := range fields {
			switch strings.ToUpper(field) {
			case "TCP":
				protocol = "tcp"
			case "UDP":
				protocol = "udp"
			}
		}
		if protocol == "" {
			continue
		}

		family := "ipv4"
		if strings.EqualFold(fields[3], "IPv6") {
			family = "ipv6"
		}

		conn := ProcessConnection{
			Protocol: protocol,
			Family:   family,
			Local:    local,
			Remote:   remote,
		}
		if len(fields) >= 10 {
			conn.State = strings.ToLower(strings.Trim(fields[9], "()"))
		}
		conns = append(conns, conn)
	}
	return conns
}
