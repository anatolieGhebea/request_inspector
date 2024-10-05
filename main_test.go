package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
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
		t.Errorf("crete session handler returned wrong status code: got %v want %v", status, http.StatusOK)
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
	sessionID := "test-session-id1"
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
		t.Errorf("session handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := "\"Request processed\"\n"
	if rr.Body.String() != expected {
		t.Errorf("session handler returned unexpected response: got %v want %v", rr.Body.String(), expected)
	}
}

// Add more test functions for other handlers...

func TestCleanupExpiredSessions(t *testing.T) {
	sessionID := "test-session-id2"
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

func TestGetSessionHandler(t *testing.T) {
	// Create a test session
	sessionID := "test-session-id3"
	testRequests := []string{"request1", "request2", "request3"}
	sessions[sessionID] = &Session{
		Requests:       testRequests,
		ExpirationTime: time.Now().Add(sessionDuration),
	}

	// Create a new request with the session ID
	req, err := http.NewRequest("GET", "/api/session/"+sessionID, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a new response recorder
	rr := httptest.NewRecorder()

	// Create a new router and register the getSessionHandler
	router := mux.NewRouter()
	router.HandleFunc("/api/session/{id}", getSessionHandler)

	// Serve the request to the router
	router.ServeHTTP(rr, req)

	// Check the response status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("get session handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body
	var responseRequests []string
	err = json.Unmarshal(rr.Body.Bytes(), &responseRequests)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(responseRequests, testRequests) {
		t.Errorf("get session  handler returned unexpected response: got %v want %v", responseRequests, testRequests)
	}

	// Clean up the test session
	delete(sessions, sessionID)
}

func TestExtendSessionHandler(t *testing.T) {
	// Create a test session
	sessionID := "test-session-id4"
	sessions[sessionID] = &Session{
		Requests:       []string{},
		ExpirationTime: time.Now().Add(extensionThreshold / 2), // Set expiration time to half of the threshold
	}

	// Create a new request with the session ID
	req, err := http.NewRequest("PUT", "/api/extend/"+sessionID, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a new response recorder
	rr := httptest.NewRecorder()

	// Create a new router and register the extendSessionHandler
	router := mux.NewRouter()
	router.HandleFunc("/api/extend/{id}", extendSessionHandler)

	// Serve the request to the router
	router.ServeHTTP(rr, req)

	// Check the response status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("extend session handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body
	expected := "\"Session extended by one hour\"\n"
	if rr.Body.String() != expected {
		t.Errorf("extend session handler returned unexpected response: got %v want %v", rr.Body.String(), expected)
	}

	// Check if the session expiration time was extended
	session := sessions[sessionID]
	if session.ExpirationTime.Before(time.Now().Add(sessionDuration - time.Minute)) {
		t.Errorf("session expiration time was not extended correctly")
	}

	// Clean up the test session
	delete(sessions, sessionID)

	// Test the case when the session does not need extension
	sessions[sessionID] = &Session{
		Requests:       []string{},
		ExpirationTime: time.Now().Add(extensionThreshold + time.Minute), // Set expiration time beyond the threshold
	}

	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("extend session handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	expected = "{\"error\":\"Session does not need extension\"}\n"
	if rr.Body.String() != expected {
		t.Errorf("extennd session handler returned unexpected response: got %v want %v", rr.Body.String(), expected)
	}

	// Clean up the test session
	delete(sessions, sessionID)
}

func TestClearSessionHandler(t *testing.T) {
	// Create a test session
	sessionID := "test-session-id5"
	sessions[sessionID] = &Session{
		Requests:        []string{"request1", "request2", "request3"},
		RequestCount:    3,
		LastRequestTime: time.Now().Add(-time.Hour),
		ExpirationTime:  time.Now().Add(sessionDuration),
	}

	// Create a new request with the session ID
	req, err := http.NewRequest("DELETE", "/api/clear/"+sessionID, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a new response recorder
	rr := httptest.NewRecorder()

	// Create a new router and register the clearSessionHandler
	router := mux.NewRouter()
	router.HandleFunc("/api/clear/{id}", clearSessionHandler)

	// Serve the request to the router
	router.ServeHTTP(rr, req)

	// Check the response status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("clear session handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body
	expected := "\"Session cleared\"\n"
	if rr.Body.String() != expected {
		t.Errorf("clear session handler returned unexpected response: got %v want %v", rr.Body.String(), expected)
	}

	// Check if the session data was cleared
	session := sessions[sessionID]
	if len(session.Requests) != 0 {
		t.Errorf("session requests were not cleared")
	}
	if session.RequestCount != 0 {
		t.Errorf("session request count was not reset")
	}
	if session.LastRequestTime.IsZero() {
		t.Errorf("session last request time was not updated")
	}

	// Clean up the test session
	delete(sessions, sessionID)

	// Test the case when the session does not exist
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("clear ses handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}

	expected = "{\"error\":\"Session not found\"}\n"
	if rr.Body.String() != expected {
		t.Errorf("clear ses handler returned unexpected response: got %v want %v", rr.Body.String(), expected)
	}
}

func TestDeleteSessionHandler(t *testing.T) {
	// Create a test session
	sessionID := "test-session-id6"
	sessions[sessionID] = &Session{
		Requests:       []string{"request1", "request2", "request3"},
		ExpirationTime: time.Now().Add(sessionDuration),
	}

	// Create a new request with the session ID
	req, err := http.NewRequest("DELETE", "/api/delete/"+sessionID, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a new response recorder
	rr := httptest.NewRecorder()

	// Create a new router and register the deleteSessionHandler
	router := mux.NewRouter()
	router.HandleFunc("/api/delete/{id}", deleteSessionHandler)

	// Serve the request to the router
	router.ServeHTTP(rr, req)

	// Check the response status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("delete session handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body
	expected := "\"Session deleted\"\n"
	if rr.Body.String() != expected {
		t.Errorf("delete session handler returned unexpected response: got %v want %v", rr.Body.String(), expected)
	}

	// Check if the session was deleted
	if _, exists := sessions[sessionID]; exists {
		t.Errorf("session was not deleted")
	}

	// Test the case when the session does not exist
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}

	expected = "{\"error\":\"Session not found\"}\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected response: got %v want %v", rr.Body.String(), expected)
	}
}

func TestCorsMiddleware(t *testing.T) {
	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Test response"))
	})

	// Wrap the test handler with the CORS middleware
	handler := corsMiddleware(testHandler)

	// Create a new request
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a new response recorder
	rr := httptest.NewRecorder()

	// Serve the request to the handler
	handler.ServeHTTP(rr, req)

	// Check the response headers
	if rr.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("Access-Control-Allow-Origin header is missing or incorrect")
	}
	if rr.Header().Get("Access-Control-Allow-Headers") != HEAR_CONTENT_TYPE {
		t.Errorf("Access-Control-Allow-Headers header is missing or incorrect")
	}
	if rr.Header().Get("Access-Control-Allow-Methods") != "GET, POST, PUT, DELETE, OPTIONS" {
		t.Errorf("Access-Control-Allow-Methods header is missing or incorrect")
	}

	// Check the response body
	expected := "Test response"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected response: got %v want %v", rr.Body.String(), expected)
	}
}

func TestMain(m *testing.M) {
	// Set up the test environment
	os.Setenv("PORT", "8080")

	// Start the server in a separate goroutine
	go main()

	// Wait for the server to start
	time.Sleep(1 * time.Second)

	// Run the tests
	code := m.Run()

	// Exit with the test result
	os.Exit(code)
}

func TestRootHandler(t *testing.T) {
	// Create a new request
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a new response recorder
	rr := httptest.NewRecorder()

	// Create a file server for serving static files
	fs := http.FileServer(http.Dir("./static"))

	// Create a new router
	r := mux.NewRouter()
	r.Use(corsMiddleware)

	// Serve index.html for the root URL ("/")
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/index.html")
	})

	// Serve static files for all other routes
	r.PathPrefix("/").Handler(http.StripPrefix("/", fs))

	// Serve the request to the router
	r.ServeHTTP(rr, req)

	// Check the response status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response content type
	expected := "text/html; charset=utf-8"
	if contentType := rr.Header().Get("Content-Type"); contentType != expected {
		t.Errorf("handler returned unexpected content type: got %v want %v", contentType, expected)
	}
}
