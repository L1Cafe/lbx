package main

import (
	"fmt"
	"github.com/L1Cafe/lbx/config"
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
	d1 := config.Server{Url: "http://localhost:8081", Healthy: true}
	d2 := config.Server{Url: "http://localhost:8082", Healthy: true}
	dDuration, _ := time.ParseDuration("10s")
	defaultSite := config.SiteParsedConfig{
		Mutex:         c.Sites["default"].Mutex,
		Servers:       []config.Server{d1, d2},
		RefreshPeriod: dDuration,
		Domain:        "",
		Path:          "/",
		Port:          8080,
	}
	s1 := config.Server{Url: "http://localhost:8083", Healthy: true}
	s1Duration, _ := time.ParseDuration("60s")
	siteTest := config.SiteParsedConfig{
		Mutex:         c.Sites["site_test"].Mutex,
		Servers:       []config.Server{s1},
		RefreshPeriod: s1Duration,
		Domain:        "localhost",
		Path:          "/folder",
		Port:          5000,
	}
	ds := config.Server{Url: "http://localhost:8280", Healthy: true}
	defaultTest := config.SiteParsedConfig{
		Mutex:         c.Sites["default_test"].Mutex,
		Servers:       []config.Server{ds},
		RefreshPeriod: dDuration,
		Domain:        "",
		Path:          "/",
		Port:          c.ListeningPort,
	}
	ps := config.Server{Url: "http://localhost:8380", Healthy: true}
	portTest := config.SiteParsedConfig{
		Mutex:         c.Sites["port_test"].Mutex,
		Servers:       []config.Server{ps},
		RefreshPeriod: dDuration,
		Domain:        "",
		Path:          "/",
		Port:          6789,
	}
	paths := config.Server{Url: "http://localhost:5305", Healthy: true}
	pathTest := config.SiteParsedConfig{
		Mutex:         c.Sites["path_test"].Mutex,
		Servers:       []config.Server{paths},
		RefreshPeriod: dDuration,
		Domain:        "",
		Path:          "/examplepath",
		Port:          c.ListeningPort,
	}
	doms := config.Server{Url: "http://localhost:8479", Healthy: true}
	domainTest := config.SiteParsedConfig{
		Mutex:         c.Sites["domain_test"].Mutex,
		Servers:       []config.Server{doms},
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
