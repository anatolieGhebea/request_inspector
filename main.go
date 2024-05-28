package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type Session struct {
	Requests        []string
	Mutex           sync.Mutex
	RequestCount    int
	LastRequestTime time.Time
	ExpirationTime  time.Time
}

var sessions = make(map[string]*Session)
var sessionsMutex sync.RWMutex // Use RWMutex for read-write locking

const (
	defaultPort        = 8189
	requestLimit       = 10
	requestWindow      = time.Minute
	sessionDuration    = time.Hour
	extensionThreshold = 10 * time.Minute
)

func createSessionHandler(w http.ResponseWriter, r *http.Request) {
	id := uuid.New().String()
	sessions[id] = &Session{
		Requests:       make([]string, 0),
		ExpirationTime: time.Now().Add(sessionDuration),
	}

	response := map[string]string{"session_id": id}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func sessionRequestHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	session, exists := sessions[id]
	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}
	session.Mutex.Lock()
	defer session.Mutex.Unlock()

	// Throttling logic
	now := time.Now()
	if now.Sub(session.LastRequestTime) > requestWindow {
		// Reset count and time window if the time window has passed
		session.RequestCount = 0
		session.LastRequestTime = now
	}

	if session.RequestCount >= requestLimit {
		http.Error(w, "Request limit exceeded", http.StatusTooManyRequests)
		return
	}

	// Process the request
	reqBytes, err := httputil.DumpRequest(r, true)
	if err != nil {
		http.Error(w, "Failed to dump request", http.StatusInternalServerError)
		return
	}
	reqString := string(reqBytes)
	if len(session.Requests) >= 5 {
		session.Requests = session.Requests[1:]
	}
	session.Requests = append(session.Requests, reqString)
	session.RequestCount++

	json.NewEncoder(w).Encode("Request processed")
}

func getSessionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	session, exists := sessions[id]
	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}
	session.Mutex.Lock()
	defer session.Mutex.Unlock()
	json.NewEncoder(w).Encode(session.Requests)
}

func extendSessionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	session, exists := sessions[id]
	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	session.Mutex.Lock()
	defer session.Mutex.Unlock()

	remainingTime := time.Until(session.ExpirationTime)
	if remainingTime < extensionThreshold {
		session.ExpirationTime = time.Now().Add(sessionDuration)
		json.NewEncoder(w).Encode("Session extended by one hour")
	} else {
		http.Error(w, "Session does not need extension", http.StatusBadRequest)
	}
}

func cleanupExpiredSessions() {
	for {
		time.Sleep(time.Minute) // Run cleanup every minute
		now := time.Now()

		sessionsMutex.Lock()
		for id, session := range sessions {
			if now.After(session.ExpirationTime) {
				delete(sessions, id)
			}
		}
		sessionsMutex.Unlock()
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	go cleanupExpiredSessions() // Start the cleanup goroutine

	strPort, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		strPort = defaultPort
	}

	port := fmt.Sprintf(":%d", strPort)

	r := mux.NewRouter()
	r.Use(corsMiddleware) // Apply the middleware
	r.HandleFunc("/create", createSessionHandler)
	r.HandleFunc("/request/{id}", sessionRequestHandler)
	r.HandleFunc("/session/{id}", getSessionHandler)
	r.HandleFunc("/extend/{id}", extendSessionHandler)
	fmt.Println("Server started on port", port)
	http.ListenAndServe(port, r)
}
