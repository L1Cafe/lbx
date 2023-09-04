package main

import (
	"fmt"
	"github.com/L1Cafe/lbx/config"
	"github.com/L1Cafe/lbx/log"
	"github.com/L1Cafe/lbx/site"
	"io"
	"net/http"
)

var appConfig *config.ParsedConfig

func handler(w http.ResponseWriter, r *http.Request) {
	log.Wrapper(log.Info, fmt.Sprintf("Received request from %s: %s", r.RemoteAddr, r.RequestURI))
	serverAddr, serverErr := site.GetRandomHealthyEndpoint("default")
	if serverErr != nil {
		log.Wrapper(log.Error, fmt.Sprintf("%s, request not fulfilled", serverErr.Error()))
		http.Error(w, serverErr.Error(), http.StatusServiceUnavailable)
		return
	}

	// TODO Update the request URL with the backend server's URL
	r.URL.Scheme = "http"
	r.URL.Host = "??????????"

	// Forward the request to the chosen backend server
	resp, err := http.DefaultTransport.RoundTrip(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()
	log.Wrapper(log.Info, fmt.Sprintf("Forwarded request to %s", serverAddr.String()))

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
	http.HandleFunc("/", handler)
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Wrapper(log.Fatal, fmt.Sprintf("Error starting web server: %s", err.Error()))
	}
}
