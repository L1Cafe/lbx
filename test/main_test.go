package main

import (
	"fmt"
	"github.com/L1Cafe/lbx/site"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/L1Cafe/lbx/config"
	"github.com/google/uuid"
)

func TestGoogleEndToEnd(t *testing.T) {
	c, err := config.LoadConfig("e2e.yaml")
	if err != nil {
		t.Fatalf("%s", err.Error())
	}
	site.Init(c)
	tenSeconds, _ := time.ParseDuration("10s")
	time.Sleep(tenSeconds)
	sitePort := c.Sites["default"].Port
	err = nil
	_, err = http.Get("http://127.0.0.1:" + strconv.Itoa(int(sitePort)))
	if err != nil {
		t.Fatalf("%s", err.Error())
	}
}

func TestEndToEnd(t *testing.T) {
	var RandomUUID string
	RandomUUID = uuid.NewString()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprintf(w, "Hello, %s", RandomUUID)
		if err != nil {
			t.Fatal(err)
		}
	}))
	defer ts.Close()

	// Creating a configuration
	var RawConfig config.RawConfig
	var siteRawConfig config.SiteRawConfig
	var GlobalRawConfig config.GlobalRawConfig

	siteRawConfig = config.SiteRawConfig{
		//Servers:     []string{ts.URL},
		CheckPeriod: 10 * time.Second,
	}
	sites := make(map[string]config.SiteRawConfig)
	sites["default"] = siteRawConfig
	GlobalRawConfig = config.GlobalRawConfig{
		ListeningPort: 8080,
		LogLevel:      0,
	}
	RawConfig = config.RawConfig{Global: GlobalRawConfig, Sites: sites}

	// TODO need to start a process with the appropriate settings

	fmt.Printf("%v", RawConfig) // FIXME remove this when this function is finished

}
