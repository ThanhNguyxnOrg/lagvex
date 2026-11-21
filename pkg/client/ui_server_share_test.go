package client

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ThanhNguyxnOrg/lagvex/pkg/profiles"
)

func newTestShareServer(t *testing.T) *UIServer {
	t.Helper()
	pm, err := profiles.NewManager("")
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	u := &UIServer{
		profileMgr:  pm,
		prober:      NewProber(),
		configStore: NewConfigStore(filepath.Join(t.TempDir(), "user_config.json")),
		cpuSampler:  NewProcessCPUSampler(),
	}
	return u
}

func doGuardedPost(t *testing.T, handler http.Handler, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	data, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(dashboardRequestHeader, "1")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func TestRelayImportExportDeleteURI(t *testing.T) {
	u := newTestShareServer(t)
	h := withLocalRequestGuard(u.setupRoutes())

	uri := "lagvex://203.0.113.55:51820?psk=secret123&name=SG+Squad&loc=Singapore&continent=Asia-Pacific&id=sg-squad"
	rec := doGuardedPost(t, h, "/api/relay/import-uri", map[string]string{"uri": uri})
	if rec.Code != http.StatusOK {
		t.Fatalf("import-uri status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "secret123") {
		t.Fatalf("import-uri response must not leak PSK: %s", rec.Body.String())
	}

	// Export URI
	expRec := doGuardedPost(t, h, "/api/relay/export-uri", map[string]string{"id": "sg-squad"})
	if expRec.Code != http.StatusOK {
		t.Fatalf("export-uri status = %d, body = %s", expRec.Code, expRec.Body.String())
	}
	var expBody struct {
		URI string `json:"uri"`
	}
	if err := json.Unmarshal(expRec.Body.Bytes(), &expBody); err != nil {
		t.Fatalf("decode export-uri: %v", err)
	}
	if !strings.Contains(expBody.URI, "psk=secret123") {
		t.Fatalf("exported URI missing PSK: %s", expBody.URI)
	}

	// Delete relay
	delRec := doGuardedPost(t, h, "/api/relay/delete", map[string]string{"id": "sg-squad"})
	if delRec.Code != http.StatusOK {
		t.Fatalf("delete status = %d, body = %s", delRec.Code, delRec.Body.String())
	}
	if len(u.profileMgr.Catalog().Relays) != 0 {
		t.Fatalf("expected 0 relays after delete, got %d", len(u.profileMgr.Catalog().Relays))
	}
}

func TestRelayImportWireGuardRedactsSecrets(t *testing.T) {
	u := newTestShareServer(t)
	h := withLocalRequestGuard(u.setupRoutes())

	priv := base64.StdEncoding.EncodeToString([]byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"))
	pub := base64.StdEncoding.EncodeToString([]byte("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"))
	conf := "[Interface]\nPrivateKey = " + priv + "\nAddress = 172.16.0.2/32\nMTU = 1280\n[Peer]\nPublicKey = " + pub + "\nEndpoint = 162.159.192.1:2408\n"

	rec := doGuardedPost(t, h, "/api/relay/import-wireguard", map[string]string{
		"config": conf,
		"name":   "WARP Custom",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("import-wireguard status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), priv) {
		t.Fatalf("import-wireguard leaked PrivateKey")
	}

	// GET /api/relays must also never expose PrivateKey
	getReq := httptest.NewRequest(http.MethodGet, "/api/relays", nil)
	getRec := httptest.NewRecorder()
	h.ServeHTTP(getRec, getReq)
	if strings.Contains(getRec.Body.String(), priv) {
		t.Fatalf("/api/relays leaked WireGuard PrivateKey")
	}
}

func TestRelayImportSubscription(t *testing.T) {
	subSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("lagvex://198.51.100.10:51820?psk=k1&name=Node+A&id=node-a\nlagvex://198.51.100.11:51820?psk=k2&name=Node+B&id=node-b\n"))
	}))
	defer subSrv.Close()

	allowLocalSubscriptionForTests = true
	defer func() { allowLocalSubscriptionForTests = false }()

	u := newTestShareServer(t)
	h := withLocalRequestGuard(u.setupRoutes())

	rec := doGuardedPost(t, h, "/api/relay/import-subscription", map[string]string{"url": subSrv.URL})
	if rec.Code != http.StatusOK {
		t.Fatalf("import-subscription status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if len(u.profileMgr.Catalog().Relays) != 2 {
		t.Fatalf("expected 2 imported relays, got %d", len(u.profileMgr.Catalog().Relays))
	}
}
