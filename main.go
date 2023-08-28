package main

import (
	"fmt"
	"github.com/L1Cafe/lbx/log"
	"io"
	"net/http"
	"time"

	"github.com/L1Cafe/lbx/config"
)

var appConfig *config.ParsedConfig

// currentServerIndex is the index of the server we're currently using
var currentServerIndex int = 0 // TODO this either needs to go into the site struct or make a map

func healthCheck(t time.Duration, key string) {
	servers := appConfig.Sites[key].Servers
	serverMutex := appConfig.Sites[key].Mutex
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
		time.Sleep(appConfig.Sites[key].RefreshPeriod)
	}
}

func getServer(key string) (string, error) {
	servers := appConfig.Sites[key].Servers
	// TODO Needs to be reworked to account for multiple sites
	serverMutex := appConfig.Sites[key].Mutex
	serverMutex.Lock()
	defer serverMutex.Unlock()
	for i := 0; i < len(servers); i++ {
		currentServer := servers[currentServerIndex]
		currentServerIndex = (currentServerIndex + 1) % len(servers)

		if currentServer.Healthy {
			return currentServer.Url, nil
		}
	}
	return "", fmt.Errorf("no healthy servers available for site %s", key)
}

func handler(w http.ResponseWriter, r *http.Request) {
	log.Wrapper(Info, fmt.Sprintf("Received request from %s: %s", r.RemoteAddr, r.RequestURI))
	serverAddr, serverErr := getServer("default")
	if serverErr != nil {
		log.Wrapper(log.Error, fmt.Sprintf("%s, request not fulfilled", serverErr.Error()))
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
	log.Wrapper(log.Info, fmt.Sprintf("Forwarded request to %s", serverAddr))

	// Copy the response back to the client
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Wrapper(log.Fatal, fmt.Sprintf("Error attempting to respond to %s", r.RemoteAddr))
	}
	log.Wrapper(log.Info, fmt.Sprintf("Responded to %s", r.RemoteAddr))
}

func main() {
	c, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Wrapper(log.Fatal, fmt.Sprintf("Error loading config: %s", err.Error()))
	}
	appConfig = c
	log.Init(c.LogLevel)
	// TODO need to make a wrapper to spawn one goroutine for each site
	go healthCheck(c.Sites["default"].RefreshPeriod, "default")
	http.HandleFunc("/", handler)
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Wrapper(log.Fatal, fmt.Sprintf("Error starting web server: %s", err.Error()))
	}
}
