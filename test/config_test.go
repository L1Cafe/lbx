package main

import (
	"fmt"
	"github.com/L1Cafe/lbx/config"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestBadPort(t *testing.T) {
	_, err := config.LoadConfig("bad_port.yaml")
	if err == nil {
		t.Errorf("Out of range port was accepted")
	}
	if !strings.Contains(err.Error(), "port number 999999999999 is out of range for site bad_port") {
		t.Errorf("Unexpected error message: %s", err.Error())
	}
}

func testInvalidEndpoint(t *testing.T) {
	invalidEndpoint, err := config.LoadConfig("invalid_endpoint.yaml")
	fmt.Println(invalidEndpoint)
	fmt.Println(err.Error())
}

func TestOverrideDefault(t *testing.T) {
	overrideConfig, err := config.LoadConfig("override_default.yaml")
	if err != nil {
		t.Errorf("Error loading configuration override_default.yaml")
	}
	if overrideConfig.Sites["default"].Domain != "" ||
		overrideConfig.Sites["default"].Path != "/" ||
		overrideConfig.Sites["default"].Port != overrideConfig.ListeningPort {
		t.Errorf("Default configuration must never be overridden for domain, path, or port:\n%#v",
			overrideConfig.Sites["default"])
	}
}

func TestReadConfig(t *testing.T) {
	c, err := config.LoadConfig("config_test.yaml")
	if err != nil {
		t.Errorf("Loading config not successful: %s", err.Error())
	}

	d1, _ := url.Parse("http://localhost:8081")
	d2, _ := url.Parse("http://localhost:8081")
	dDuration, _ := time.ParseDuration("10s")
	defaultSite := config.SiteParsedConfig{
		Endpoints:     []url.URL{*d1, *d2},
		RefreshPeriod: dDuration,
		Domain:        "",
		Path:          "/",
		Port:          8080,
	}
	s1u, _ := url.Parse("http://localhost:8083")
	s1Duration, _ := time.ParseDuration("60s")
	siteTest := config.SiteParsedConfig{
		Endpoints:     []url.URL{*s1u},
		RefreshPeriod: s1Duration,
		Domain:        "localhost",
		Path:          "/folder",
		Port:          5000,
	}
	du, _ := url.Parse("http://localhost:8280")
	defaultTest := config.SiteParsedConfig{
		Endpoints:     []url.URL{*du},
		RefreshPeriod: dDuration,
		Domain:        "",
		Path:          "/",
		Port:          c.ListeningPort,
	}
	pu, _ := url.Parse("http://localhost:8380")
	portTest := config.SiteParsedConfig{
		Endpoints:     []url.URL{*pu},
		RefreshPeriod: dDuration,
		Domain:        "",
		Path:          "/",
		Port:          6789,
	}
	pau, _ := url.Parse("http://localhost:8380")
	pathTest := config.SiteParsedConfig{
		Endpoints:     []url.URL{*pau},
		RefreshPeriod: dDuration,
		Domain:        "",
		Path:          "/examplepath",
		Port:          c.ListeningPort,
	}
	domu, _ := url.Parse("http://localhost:8479")
	domainTest := config.SiteParsedConfig{
		Endpoints:     []url.URL{*domu},
		RefreshPeriod: dDuration,
		Domain:        "example.com",
		Path:          "/",
		Port:          c.ListeningPort,
	}
	expectedConfig := config.ParsedConfig{
		ListeningPort: uint16(8080),
		LogLevel:      1,
		Sites: map[string]config.SiteParsedConfig{
			"default":      defaultSite,
			"site_test":    siteTest,
			"default_test": defaultTest,
			"domain_test":  domainTest,
			"path_test":    pathTest,
			"port_test":    portTest,
		},
	}
	if !reflect.DeepEqual(expectedConfig, *c) {
		fmt.Printf("Expected configuration: %v\n", expectedConfig)
		fmt.Printf("Configuration received: %v\n", *c)
		t.Errorf("Unexpected configuration file")
	}
}
