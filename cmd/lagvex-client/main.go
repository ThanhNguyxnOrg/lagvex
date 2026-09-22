package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/ThanhNguyxnOrg/lagvex/pkg/client"
	"github.com/ThanhNguyxnOrg/lagvex/pkg/profiles"
)

var (
	version = "1.0.0"
)

func main() {
	attachParentConsole()

	profileFlag := flag.String("profiles", "", "Path to profiles.json")
	httpFlag := flag.String("http", "127.0.0.1:18888", "HTTP address for Web UI Dashboard")
	noBrowserFlag := flag.Bool("no-browser", false, "Do not auto-open browser on startup")

	// Direct CLI connection flags
	connectFlag := flag.Bool("connect", false, "Directly initiate tunnel from CLI")
	relayFlag := flag.String("relay", "", "Relay server endpoint (host:port)")
	pskFlag := flag.String("psk", "", "Pre-shared key")
	gameFlag := flag.String("game", "", "Target game ID (e.g. valorant, cs2, pubg)")
	regionFlag := flag.String("region", "asia-sg", "Target server region ID (e.g. asia-sg)")
	forceFlag := flag.Bool("force", false, "Force install routes immediately without waiting for game process")
	versionFlag := flag.Bool("version", false, "Print version")

	flag.Parse()

	if *versionFlag {
		fmt.Printf("Lagvex Client v%s\n", version)
		return
	}

	// Single instance check: if an instance is already running on httpFlag, open UI and exit
	if !*connectFlag {
		checkClient := &http.Client{Timeout: 400 * time.Millisecond}
		if resp, err := checkClient.Get(fmt.Sprintf("http://%s/api/status", *httpFlag)); err == nil {
			_ = resp.Body.Close()
			log.Printf("[Lagvex] Instance already running on http://%s. Opening dashboard...", *httpFlag)
			if !*noBrowserFlag {
				openDesktopWindow(fmt.Sprintf("http://%s", *httpFlag))
			}
			return
		}
	}

	// Locate profiles.json
	profilePath := *profileFlag
	if profilePath == "" {
		exe, _ := os.Executable()
		candidates := []string{
			filepath.Join(filepath.Dir(exe), "configs", "profiles.json"),
			"configs/profiles.json",
		}
		for _, c := range candidates {
			if _, err := os.Stat(c); err == nil {
				profilePath = c
				break
			}
		}
	}

	pm, err := profiles.NewManager(profilePath)
	if err != nil {
		log.Printf("[Warning] Failed loading profiles: %v", err)
	} else {
		cat := pm.Catalog()
		modeStr := "custom disk config"
		if profilePath == "" {
			modeStr = "embedded zero-config"
		}
		log.Printf("[Profiles] Loaded %d game profiles & %d community relays (%s)", len(cat.Games), len(cat.Relays), modeStr)
	}

	engine, err := client.NewEngine(pm)
	if err != nil {
		log.Fatalf("Failed to initialize client engine: %v", err)
	}

	_, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Printf("[Client] Received signal %v, disconnecting...", sig)
		_ = engine.Disconnect()
		cancel()
		os.Exit(0)
	}()

	// CLI Direct Mode
	if *connectFlag {
		if *relayFlag == "" || *pskFlag == "" {
			log.Fatalf("Error: -relay and -psk are required when using -connect")
		}
		log.Printf("Connecting via CLI to %s for game %s (%s)...", *relayFlag, *gameFlag, *regionFlag)
		if err := engine.Connect(*relayFlag, []byte(*pskFlag), *gameFlag, *regionFlag, *forceFlag); err != nil {
			log.Fatalf("Connection failed: %v", err)
		}

		// Keep alive in CLI mode
		for {
			time.Sleep(3 * time.Second)
			stats := engine.Stats()
			log.Printf("[Status] %s | Ping: %d ms | Up: %d KB/s | Down: %d KB/s | Routes: %d | GameActive: %v",
				stats.State, stats.PingMs, stats.UpRateBps/1024, stats.DownRateBps/1024, stats.RouteCount, stats.GameRunning)
		}
	}

	// Web Dashboard Mode
	uiServer := client.NewUIServer(engine, pm, "")

	if !*noBrowserFlag {
		go func() {
			time.Sleep(800 * time.Millisecond)
			openDesktopWindow(fmt.Sprintf("http://%s", *httpFlag))
		}()
	}

	log.Printf("==================================================")
	log.Printf("  Lagvex Gaming Latency Reducer v%s", version)
	log.Printf("  Dashboard: http://%s", *httpFlag)
	log.Printf("  [Notice] Educational & network research use only.")
	log.Printf("==================================================")

	if err := uiServer.Start(*httpFlag); err != nil {
		log.Fatalf("UI server failed: %v", err)
	}
}

func openDesktopWindow(url string) {
	if runtime.GOOS == "windows" {
		// 1. Try Microsoft Edge standalone App Mode (clean desktop window with Lagvex UI)
		edgeCandidates := []string{
			filepath.Join(os.Getenv("ProgramFiles(x86)"), "Microsoft", "Edge", "Application", "msedge.exe"),
			filepath.Join(os.Getenv("ProgramFiles"), "Microsoft", "Edge", "Application", "msedge.exe"),
			filepath.Join(os.Getenv("LocalAppData"), "Microsoft", "Edge", "Application", "msedge.exe"),
		}
		for _, p := range edgeCandidates {
			if _, err := os.Stat(p); err == nil {
				cmd := exec.Command(p, fmt.Sprintf("--app=%s", url), "--window-size=1220,840")
				if err := cmd.Start(); err == nil {
					return
				}
			}
		}

		// 2. Try Google Chrome standalone App Mode
		chromeCandidates := []string{
			filepath.Join(os.Getenv("ProgramFiles"), "Google", "Chrome", "Application", "chrome.exe"),
			filepath.Join(os.Getenv("ProgramFiles(x86)"), "Google", "Chrome", "Application", "chrome.exe"),
			filepath.Join(os.Getenv("LocalAppData"), "Google", "Chrome", "Application", "chrome.exe"),
		}
		for _, p := range chromeCandidates {
			if _, err := os.Stat(p); err == nil {
				cmd := exec.Command(p, fmt.Sprintf("--app=%s", url), "--window-size=1220,840")
				if err := cmd.Start(); err == nil {
					return
				}
			}
		}

		// 3. Fallback to default browser
		_ = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
		return
	}

	if runtime.GOOS == "darwin" {
		_ = exec.Command("open", url).Start()
		return
	}

	_ = exec.Command("xdg-open", url).Start()
}
