package main

import (
	"io"
	"net/http"
)

var servers = []string{
	"http://localhost:8081",
	"http://localhost:8082",
	// Add more backend servers as needed
}

var currentServer = 0

func getNextServer() string {
	server := servers[currentServer]
	currentServer = (currentServer + 1) % len(servers)
	return server
}

func handler(w http.ResponseWriter, r *http.Request) {
	server := getNextServer()

	// Update the request URL with the backend server's URL
	r.URL.Scheme = "http"
	r.URL.Host = server

	// Forward the request to the chosen backend server
	resp, err := http.DefaultTransport.RoundTrip(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Copy the response back to the client
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	io.Copy(w, resp.Body)
	resp.Body.Close()
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
