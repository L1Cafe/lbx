package main

import (
	"fmt"
	"github.com/L1Cafe/lbx/site"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/L1Cafe/lbx/config"
	"github.com/google/uuid"
)

func TestWikipediaEndToEnd(t *testing.T) {
	c, confErr := config.LoadConfig("e2e_wikipedia.yaml")
	if confErr != nil {
		t.Fatalf("%s", confErr.Error())
	}
	site.Init(c)
	twoSeconds, _ := time.ParseDuration("2s")
	time.Sleep(twoSeconds)
	sitePort := c.Sites["default"].Port
	response, respErr := http.Get("http://127.0.0.1:" + strconv.Itoa(int(sitePort)))
	if respErr != nil {
		t.Fatalf("%s", respErr.Error())
	}
	if response.StatusCode != 200 {
		t.Fatalf("Received status code %d, expected a status code of 200", response.StatusCode)
	}
	body, readErr := io.ReadAll(response.Body)
	if readErr != nil {
		t.Fatalf("Error while reading body: %v", readErr.Error())
	}
	bodyStr := string(body)
	if !strings.Contains(bodyStr, "wikimedia") {
		t.Errorf("Expected to find the substring \"wikimedia\" on Wikipedia, but that wasn't the case: %s", bodyStr)
	}
	site.Stop()
}

func TestEndToEnd(t *testing.T) {
	var RandomUUID string
	RandomUUID = uuid.NewString()
	l, lErr := net.Listen("tcp", "127.0.0.1:5678")
	if lErr != nil {
		t.Fatalf("%v", lErr)
	}
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprintf(w, "Hello, %s", RandomUUID)
		if err != nil {
			t.Fatal(err)
		}
	}))
	ts.Listener.Close()
	ts.Listener = l
	ts.Start()
	defer ts.Close()
	twoSeconds, _ := time.ParseDuration("2s")
	defer ts.Close()

	c, confErr := config.LoadConfig("e2e.yaml")
	if confErr != nil {
		t.Fatalf("%s", confErr.Error())
	}
	site.Init(c)
	time.Sleep(twoSeconds)
	sitePort := c.Sites["default"].Port
	response, respErr := http.Get("http://127.0.0.1:" + strconv.Itoa(int(sitePort)))
	if respErr != nil {
		t.Fatalf("%s", respErr.Error())
	}
	if response.StatusCode != 200 {
		t.Fatalf("Received status code %d, expected a status code of 200", response.StatusCode)
	}
	body, readErr := io.ReadAll(response.Body)
	if readErr != nil {
		t.Fatalf("Error while reading body: %v", readErr.Error())
	}
	bodyStr := string(body)
	if !strings.Contains(bodyStr, RandomUUID) {
		t.Errorf("Expected to find the substring %q on the local server, but that wasn't the case: %s", RandomUUID, bodyStr)
	}
	site.Stop()
}
