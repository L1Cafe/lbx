package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/L1Cafe/lbx/config"
	"github.com/google/uuid"
)

func TestEndToEnd(t *testing.T) {
	var RandomUUID string
	RandomUUID = uuid.NewString()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %s", RandomUUID)
	}))
	defer ts.Close()

	// Creating a configuration
	var RawConfig config.RawConfig
	var siteRawConfig config.SiteRawConfig
	var GlobalRawConfig config.GlobalRawConfig

	siteRawConfig = config.SiteRawConfig{
		Servers:       []string{ts.URL},
		RefreshPeriod: 10 * time.Second,
	}
	sites := make(map[string]config.SiteRawConfig)
	sites["default"] = siteRawConfig
	GlobalRawConfig = config.GlobalRawConfig{
		ListeningPort: 8080,
		LogLevel:      0,
	}
	RawConfig = config.RawConfig{Global: GlobalRawConfig, Sites: sites}

	// TODO need to start a process with the appropriate settings

	fmt.Printf("%v", RawConfig)

}
