package main

import (
	"io"
	"log"
	"net/http"
	"time"
)

type ServerBackend struct {
	Url     string
	Healthy bool
}

var servers = []ServerBackend{
	// TODO add configuration instead of hardcoding this
	{Url: "http://localhost:8081", Healthy: true},
	{Url: "http://localhost:8082", Healthy: true},
}

var currentServer = 0

func healthCheck() {
	for {
		for i, server := range servers {
			resp, err := http.Head(server.Url)
			if err != nil || resp.StatusCode != 200 {
				// Mark as unhealthy
				servers[i].Healthy = false
			} else {
				servers[i].Healthy = true
			}
		}
		time.Sleep(10 * time.Second)
	}
}

func getNextServer() string {
	for {
		// TODO fix this loop
		server := servers[currentServer]
		currentServer = (currentServer + 1) % len(servers)
		if server.Healthy {
			return server.Url
		}
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received request from %s\n", r.RemoteAddr)
	server := getNextServer()
	// TODO fix this handling of bad servers
	if server == "" {
		http.Error(w, "No available servers", http.StatusServiceUnavailable)
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
