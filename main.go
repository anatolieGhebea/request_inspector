package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
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
}

var sessions = make(map[string]*Session)

const requestLimit = 10
const requestWindow = time.Minute

func createSessionHandler(w http.ResponseWriter, r *http.Request) {
	id := uuid.New().String()
	sessions[id] = &Session{
		Requests: make([]string, 0),
	}
	fmt.Fprint(w, id)
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

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/create", createSessionHandler)
	r.HandleFunc("/request/{id}", sessionRequestHandler)
	r.HandleFunc("/session/{id}", getSessionHandler)
	http.ListenAndServe(":8189", r)
}
