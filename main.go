package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/L1Cafe/lbx/config"
)

const (
	Info  = 0
	Warn  = 1
	Error = 2
	Fatal = 3
)

// To ensure thread safety
var serverMutex sync.Mutex // TODO this needs to be per-site not global

var appConfig *config.ParsedConfig

// currentServerIndex is the index of the server we're currently using
var currentServerIndex int = 0 // TODO this either needs to go into the site struct or make a map

func logWrapper(level int, msg string) { // TODO make this a variadic function with: client, server, site, response code
	var levelStr string // TODO configurable logging output
	fatal := false
	switch level {
	case Info:
		levelStr = "INFO "
	case Warn:
		levelStr = "WARN "
	case Error:
		levelStr = "ERROR"
	case Fatal:
		levelStr = "FATAL"
		fatal = true
	default:
		levelStr = strconv.Itoa(level)
	}
	if fatal {
		log.Fatalf("[%s] %s\n", levelStr, msg)
	} else {
		log.Printf("[%s] %s\n", levelStr, msg)
	}
}

func healthCheck(t time.Duration, key string) {
	servers := appConfig.Sites[key].Servers
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
		time.Sleep(t) // TODO configure this timer
	}
}

func getServer(key string) (string, error) {
	servers := appConfig.Sites[key].Servers
	// TODO Needs to be reworked to account for multiple sites
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
	serverAddr, serverErr := getServer("default")
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
	logWrapper(Info, fmt.Sprintf("Forwarded request to %s", serverAddr))

	// Copy the response back to the client
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		logWrapper(Fatal, fmt.Sprintf("Error attempting to respond to %s", r.RemoteAddr))
	}
	logWrapper(Info, fmt.Sprintf("Responded to %s", r.RemoteAddr))
}

func main() {
	c, err := config.LoadConfig("config.yaml")
	if err != nil {
		logWrapper(Fatal, fmt.Sprintf("Error loading config: %s", err.Error()))
	}
	appConfig = c
	// TODO need to make a wrapper to spawn one goroutine for each site
	go healthCheck(c.Sites["default"].RefreshPeriod, "default")
	http.HandleFunc("/", handler)
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		logWrapper(Fatal, fmt.Sprintf("Error starting web server: %s", err.Error()))
	}
}
