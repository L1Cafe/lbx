package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

const (
	Info  = 0
	Warn  = 1
	Error = 2
	Fatal = 3
)

type Server struct {
	Url     string
	Healthy bool
}

// To ensure thread safety
var serverMutex sync.Mutex

// currentServerIndex is the index of the server we're currently using
var currentServerIndex int = 0

var servers = []Server{
	// TODO add configuration instead of hardcoding this
	{Url: "http://localhost:8081", Healthy: true},
	{Url: "http://localhost:8082", Healthy: true},
}

func logWrapper(level int, msg string) {
	var levelStr string
	switch level {
	case Info:
		levelStr = "INFO "
	case Warn:
		levelStr = "WARN "
	case Error:
		levelStr = "ERROR"
	case Fatal:
		levelStr = "FATAL"
	default:
		levelStr = "?"
	}
	log.Printf("[%s] %s\n", levelStr, msg)
}

func healthCheck() {
	for {
		for i, server := range servers {
			resp, err := http.Head(server.Url)
			if err != nil || resp.StatusCode != 200 {
				// Mark as unhealthy
				serverMutex.Lock()
				servers[i].Healthy = false
				serverMutex.Unlock()
			} else {
				serverMutex.Lock()
				servers[i].Healthy = true
				serverMutex.Unlock()
			}
		}
		time.Sleep(10 * time.Second) // TODO configure this timer
	}
}

func getServer() (string, error) {
	serverMutex.Lock()
	defer serverMutex.Unlock()
	for i := 0; i < len(servers); i++ {
		currentServer := servers[currentServerIndex]
		currentServerIndex = (currentServerIndex + 1) % len(servers)

		if currentServer.Healthy {
			return currentServer.Url, nil
		}
	}
	return "", fmt.Errorf("no healthy servers available")
}

func handler(w http.ResponseWriter, r *http.Request) {
	logWrapper(Info, fmt.Sprintf("Received request from %s: %s", r.RemoteAddr, r.RequestURI))
	serverAddr, serverErr := getServer()
	if serverErr != nil {
		logWrapper(Error, fmt.Sprintf("%s, request not fulfilled", serverErr.Error()))
		http.Error(w, serverErr.Error(), http.StatusServiceUnavailable)
		return
	}

	// Update the request URL with the backend server's URL
	r.URL.Scheme = "http"
	r.URL.Host = serverAddr

	// Forward the request to the chosen backend server
	resp, err := http.DefaultTransport.RoundTrip(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()
	logWrapper(Info, fmt.Sprintf("Forwarded request to %s\n", serverAddr))

	// Copy the response back to the client
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	logWrapper(Info, fmt.Sprintf("Responded to %s\n", r.RemoteAddr))
	io.Copy(w, resp.Body)
}

func main() {
	go healthCheck()
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
