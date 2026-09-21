package client

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSquadSignalingEndpoints(t *testing.T) {
	server := NewUIServer(nil, nil, "")
	handler := server.setupRoutes()

	// 1. Create Room
	createBody := map[string]interface{}{
		"code": "LGVX-TEST",
		"name": "AcePlayer#1001",
		"game": "Valorant",
		"ping": 18,
		"isp":  "192.168.1.1",
	}
	b, _ := json.Marshal(createBody)
	req := httptest.NewRequest("POST", "/api/squad/create", bytes.NewReader(b))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on squad create, got %d", rec.Code)
	}

	// 2. Join Room
	joinBody := map[string]interface{}{
		"code": "LGVX-TEST",
		"name": "Teammate#2002",
		"game": "Valorant",
		"ping": 24,
		"isp":  "10.0.0.1",
	}
	b2, _ := json.Marshal(joinBody)
	req2 := httptest.NewRequest("POST", "/api/squad/join", bytes.NewReader(b2))
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200 on squad join, got %d", rec2.Code)
	}

	var joinRes struct {
		Members []SquadMemberInfo `json:"members"`
	}
	if err := json.NewDecoder(rec2.Body).Decode(&joinRes); err != nil {
		t.Fatalf("failed to decode squad join response: %v", err)
	}
	if len(joinRes.Members) != 2 {
		t.Fatalf("expected 2 members in squad room, got %d", len(joinRes.Members))
	}

	// 3. Heartbeat
	hbBody := map[string]interface{}{
		"code": "LGVX-TEST",
		"name": "Teammate#2002",
		"ping": 21,
	}
	b3, _ := json.Marshal(hbBody)
	req3 := httptest.NewRequest("POST", "/api/squad/heartbeat", bytes.NewReader(b3))
	rec3 := httptest.NewRecorder()
	handler.ServeHTTP(rec3, req3)
	if rec3.Code != http.StatusOK {
		t.Fatalf("expected 200 on heartbeat, got %d", rec3.Code)
	}

	// 4. Query Room
	req4 := httptest.NewRequest("GET", "/api/squad/room?code=LGVX-TEST", nil)
	rec4 := httptest.NewRecorder()
	handler.ServeHTTP(rec4, req4)
	if rec4.Code != http.StatusOK {
		t.Fatalf("expected 200 on query room, got %d", rec4.Code)
	}

	var roomRes struct {
		Code    string            `json:"code"`
		Members []SquadMemberInfo `json:"members"`
	}
	if err := json.NewDecoder(rec4.Body).Decode(&roomRes); err != nil {
		t.Fatalf("failed to decode query room: %v", err)
	}
	if len(roomRes.Members) != 2 || roomRes.Members[1].Ping != 21 {
		t.Fatalf("unexpected room state after heartbeat: %+v", roomRes.Members)
	}

	// 5. Leave Room
	leaveBody := map[string]interface{}{
		"code": "LGVX-TEST",
		"name": "Teammate#2002",
	}
	b5, _ := json.Marshal(leaveBody)
	req5 := httptest.NewRequest("POST", "/api/squad/leave", bytes.NewReader(b5))
	rec5 := httptest.NewRecorder()
	handler.ServeHTTP(rec5, req5)
	if rec5.Code != http.StatusOK {
		t.Fatalf("expected 200 on squad leave, got %d", rec5.Code)
	}

	// Verify 1 member left
	req6 := httptest.NewRequest("GET", "/api/squad/room?code=LGVX-TEST", nil)
	rec6 := httptest.NewRecorder()
	handler.ServeHTTP(rec6, req6)
	var roomResAfterLeave struct {
		Members []SquadMemberInfo `json:"members"`
	}
	_ = json.NewDecoder(rec6.Body).Decode(&roomResAfterLeave)
	if len(roomResAfterLeave.Members) != 1 {
		t.Fatalf("expected 1 member remaining, got %d", len(roomResAfterLeave.Members))
	}
}

func TestSystemInfoElevation(t *testing.T) {
	server := NewUIServer(nil, nil, "")
	handler := server.setupRoutes()

	req := httptest.NewRequest("GET", "/api/system-info", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on system-info, got %d", rec.Code)
	}

	var res map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("failed to decode system-info response: %v", err)
	}

	if _, exists := res["isAdmin"]; !exists {
		t.Fatalf("expected 'isAdmin' field in system-info response")
	}
}
