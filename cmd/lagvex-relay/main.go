package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/netip"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/ThanhNguyxnOrg/lagvex/pkg/relay"
)

var (
	version = "1.0.0"
)

func main() {
	listenFlag := flag.String("listen", ":51820", "UDP listen address for client connections")
	httpFlag := flag.String("http", "", "HTTP address for Squad Signaling & status (defaults to UDP port + 1)")
	tunFlag := flag.String("tun", "lagvex0", "Linux TUN device name")
	subnetFlag := flag.String("subnet", "10.88.0.0/24", "IPv4 subnet for connected clients")
	pskFlag := flag.String("psk", "", "Pre-shared key for authentication")
	pskFileFlag := flag.String("psk-file", "", "Path to file containing the pre-shared key")
	mtuFlag := flag.Int("mtu", 1400, "MTU for the TUN interface")
	idleTimeoutFlag := flag.Duration("idle-timeout", 90*time.Second, "Session idle timeout")
	maxClientsFlag := flag.Int("max-clients", 0, "Maximum concurrent clients (0 = unconstrained)")
	versionFlag := flag.Bool("version", false, "Print version and exit")

	flag.Parse()

	if *versionFlag {
		fmt.Printf("Lagvex Relay v%s\n", version)
		return
	}

	// Resolve PSK from flag, file, or environment
	psk := *pskFlag
	if psk == "" && *pskFileFlag != "" {
		data, err := os.ReadFile(*pskFileFlag)
		if err != nil {
			log.Fatalf("Failed to read PSK file: %v", err)
		}
		psk = strings.TrimSpace(string(data))
	}
	if psk == "" {
		psk = os.Getenv("LAGVEX_PSK")
	}
	if psk == "" {
		log.Fatalf("Error: Pre-shared key (PSK) is required via -psk, -psk-file, or LAGVEX_PSK env var")
	}

	subnet, err := netip.ParsePrefix(*subnetFlag)
	if err != nil {
		log.Fatalf("Invalid subnet: %v", err)
	}

	cfg := relay.Config{
		ListenAddr:  *listenFlag,
		HTTPAddr:    *httpFlag,
		TunName:     *tunFlag,
		Subnet:      subnet,
		PSK:         []byte(psk),
		MTU:         *mtuFlag,
		IdleTimeout: *idleTimeoutFlag,
		MaxClients:  *maxClientsFlag,
	}

	srv, err := relay.NewServer(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize relay server: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Printf("[Relay] Received signal %v, terminating...", sig)
		cancel()
	}()

	log.Printf("==================================================")
	log.Printf("  Lagvex Relay v%s starting up", version)
	log.Printf("  Subnet: %s | Listen: %s | MTU: %d", subnet, *listenFlag, *mtuFlag)
	log.Printf("==================================================")

	if err := srv.Start(ctx); err != nil {
		log.Fatalf("Relay server error: %v", err)
	}
}
