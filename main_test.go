package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
)

func TestCreateSessionHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/create", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		createSessionHandler(w, r, sessionDuration)
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatal(err)
	}

	if _, ok := response["session_id"]; !ok {
		t.Errorf("handler returned unexpected response: got %v want session_id", response)
	}
}

func TestSessionRequestHandler(t *testing.T) {
	sessionID := "test-session-id"
	sessions[sessionID] = &Session{
		Requests:       make([]string, 0),
		ExpirationTime: time.Now().Add(sessionDuration),
	}

	req, err := http.NewRequest("POST", "/api/request/"+sessionID, bytes.NewBufferString("test request"))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/api/request/{id}", sessionRequestHandler)
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := "\"Request processed\"\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected response: got %v want %v", rr.Body.String(), expected)
	}
}

// Add more test functions for other handlers...

func TestCleanupExpiredSessions(t *testing.T) {
	sessionID := "test-session-id"
	expiredSessionID := "expired-session-id"

	testDuration := 100 * time.Millisecond // 100ms

	sessions[sessionID] = &Session{
		Requests:       make([]string, 0),
		ExpirationTime: time.Now().Add(3 * testDuration),
	}
	sessions[expiredSessionID] = &Session{
		Requests:       make([]string, 0),
		ExpirationTime: time.Now().Add(testDuration),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cleanupInterval := testDuration / 10
	go cleanupExpiredSessions(ctx, cleanupInterval)

	// Wait for a few cleanup cycles to run
	time.Sleep(2 * testDuration)

	sessionsMutex.RLock()
	defer sessionsMutex.RUnlock()

	if _, ok := sessions[sessionID]; !ok {
		t.Errorf("valid session was unexpectedly deleted")
	}

	if _, ok := sessions[expiredSessionID]; ok {
		t.Errorf("expired session was not deleted")
	}
}
