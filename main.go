package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

type ServerBackend struct {
	Url     string
	Healthy bool
}

var currentServerIndex int = 0

var serverMutex sync.Mutex

var servers = []ServerBackend{
	// TODO add configuration instead of hardcoding this
	{Url: "http://localhost:8081", Healthy: true},
	{Url: "http://localhost:8082", Healthy: true},
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
		time.Sleep(10 * time.Second)
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
	log.Printf("Received request from %s\n", r.RemoteAddr)
	server, serverErr := getServer()
	// TODO fix this handling of bad servers
	if serverErr != nil {
		http.Error(w, serverErr.Error(), http.StatusServiceUnavailable)
		return
	}

	// Update the request URL with the backend server's URL
	r.URL.Scheme = "http"
	r.URL.Host = server

	// Forward the request to the chosen backend server
	resp, err := http.DefaultTransport.RoundTrip(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	log.Printf("Forwarded request to %s\n", server)

	// Copy the response back to the client
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	log.Printf("Responded to %s\n", r.RemoteAddr)
	io.Copy(w, resp.Body)
	resp.Body.Close()
}

func main() {
	go healthCheck()
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
