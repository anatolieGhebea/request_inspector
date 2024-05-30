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

type ErrorJsonResponse struct {
	Error string `json:"error"`
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
	w.Header().Set("Content-Type", "application/json")
	id := uuid.New().String()
	sessions[id] = &Session{
		Requests:       make([]string, 0),
		ExpirationTime: time.Now().Add(sessionDuration),
	}

	response := map[string]string{"session_id": id}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorJsonResponse{Error: "Failed to create session"})
	}
}

func sessionRequestHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	id := vars["id"]
	session, exists := sessions[id]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorJsonResponse{Error: "Session not found"})
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
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(ErrorJsonResponse{Error: "Request limit exceeded"})
		return
	}

	// Process the request
	reqBytes, err := httputil.DumpRequest(r, true)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorJsonResponse{Error: "Failed to dump request"})
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
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	id := vars["id"]
	session, exists := sessions[id]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorJsonResponse{Error: "Session not found"})
		return
	}
	session.Mutex.Lock()
	defer session.Mutex.Unlock()

	json.NewEncoder(w).Encode(session.Requests)
}

func extendSessionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	id := vars["id"]

	session, exists := sessions[id]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorJsonResponse{Error: "Session not found"})
		return
	}

	session.Mutex.Lock()
	defer session.Mutex.Unlock()

	remainingTime := time.Until(session.ExpirationTime)
	if remainingTime < extensionThreshold {
		session.ExpirationTime = time.Now().Add(sessionDuration)
		json.NewEncoder(w).Encode("Session extended by one hour")
	} else {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorJsonResponse{Error: "Session does not need extension"})
	}
}

func clearSessionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	id := vars["id"]

	session, exists := sessions[id]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorJsonResponse{Error: "Session not found"})
		return
	}

	session.Mutex.Lock()
	defer session.Mutex.Unlock()

	session.Requests = []string{}
	session.RequestCount = 0
	session.LastRequestTime = time.Now()

	json.NewEncoder(w).Encode("Session cleared")
}

func deleteSessionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	id := vars["id"]

	_, exists := sessions[id]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorJsonResponse{Error: "Session not found"})
		return
	}

	sessionsMutex.Lock()
	delete(sessions, id)
	sessionsMutex.Unlock()

	json.NewEncoder(w).Encode("Session deleted")
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

	// Create a file server for serving static files
	fs := http.FileServer(http.Dir("./static"))

	r := mux.NewRouter()
	r.Use(corsMiddleware) // Apply the middleware

	// Serve index.html for the root URL ("/")
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/index.html")
	})

	// API routes
	apiRouter := r.PathPrefix("/api").Subrouter()
	apiRouter.HandleFunc("/create", createSessionHandler)
	apiRouter.HandleFunc("/request/{id}", sessionRequestHandler)
	apiRouter.HandleFunc("/session/{id}", getSessionHandler)
	apiRouter.HandleFunc("/extend/{id}", extendSessionHandler)
	apiRouter.HandleFunc("/clear/{id}", clearSessionHandler)
	apiRouter.HandleFunc("/clear/{id}", clearSessionHandler)
	apiRouter.HandleFunc("/delete/{id}", deleteSessionHandler)

	// Serve static files for all other routes
	r.PathPrefix("/").Handler(http.StripPrefix("/", fs))

	fmt.Println("Server started on port", port)
	http.ListenAndServe(port, r)
}
