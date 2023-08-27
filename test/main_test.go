package main

import (
	"fmt"
	"github.com/L1Cafe/lbx/config"
	"github.com/google/uuid"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestEndToEnd(t *testing.T) {
	var RandomUUID string
	RandomUUID = uuid.NewString()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %s", RandomUUID)
	}))
	defer ts.Close()

	// Creating a configuration
	var configRaw config.ConfigRaw
	var siteConfigRaw config.SiteConfigRaw
	var globalConfig config.GlobalConfig

	siteConfigRaw = config.SiteConfigRaw{
		Servers:       []string{ts.URL},
		RefreshPeriod: 10 * time.Second,
	}
	sites := make(map[string]config.SiteConfigRaw)
	sites["default"] = siteConfigRaw
	globalConfig = config.GlobalConfig{
		ListeningPort: 8080,
		LogLevel:      0,
	}
	configRaw = config.ConfigRaw{Global: globalConfig, Sites: sites}

	// TODO need to start a process with the appropriate settings

	fmt.Printf("%v", configRaw)

}
