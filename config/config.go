package config

import (
	"errors"
	"fmt"
	"github.com/L1Cafe/lbx/log"
	"io/ioutil"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v2"
)

type Server struct {
	Url     string
	Healthy bool
}

type GlobalRawConfig struct {
	ListeningPort int `yaml:"listening_port"`
	LogLevel      int `yaml:"log_level"`
}

type SiteRawConfig struct {
	Servers       []string      `yaml:"servers"`
	RefreshPeriod time.Duration `yaml:"refresh_period"`
	// Domain is the FQDN. disabled by default
	Domain string `yaml:"domain"`
	// Path is what comes after the FQDN, "/" by default
	Path string `yaml:"path"`
	// Port is the same as the global listening port by default
	Port int `yaml:"port"`
}

type SiteParsedConfig struct {
	// TODO this needs a mutex for thread safety when serving several sites
	Mutex         *sync.Mutex
	Servers       []Server
	RefreshPeriod time.Duration
	Domain        string
	Path          string
	Port          uint16
}

// ParsedConfig is the actual configuration that the application uses
type ParsedConfig struct {
	ListeningPort uint16
	LogLevel      uint8
	Sites         map[string]SiteParsedConfig
}

// RawConfig is the struct that matches the configuration file
type RawConfig struct {
	Global GlobalRawConfig          `yaml:"global"`
	Sites  map[string]SiteRawConfig `yaml:"sites"`
}

func LoadConfig(file string) (*ParsedConfig, error) {
	var rConfig RawConfig
	var pConfig ParsedConfig
	data, err := ioutil.ReadFile(file) // FIXME ReadFile is deprecated
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(data, &rConfig)
	if err != nil {
		return nil, err
	}

	// Config parsing
	parsedLogLevel := rConfig.Global.LogLevel
	if parsedLogLevel < 0 {
		return nil, errors.New(fmt.Sprintf("logging level cannot be negative: %d", parsedLogLevel))
	}
	if parsedLogLevel > 3 {
		// Logging cannot be higher than 3 (Fatal)
		parsedLogLevel = 3
	}
	pConfig.LogLevel = uint8(rConfig.Global.LogLevel)
	globalPort := rConfig.Global.ListeningPort
	if globalPort < 1 || globalPort > 65535 {
		return nil, errors.New(fmt.Sprintf("invalid global port number %d", globalPort))
	}
	pConfig.ListeningPort = uint16(rConfig.Global.ListeningPort)
	pConfig.Sites = make(map[string]SiteParsedConfig)
	for siteName, siteValue := range rConfig.Sites {
		var parsedSite SiteParsedConfig
		parsedSite.Mutex = new(sync.Mutex)
		for _, server := range siteValue.Servers {
			parsedSite.Servers = append(parsedSite.Servers, Server{Url: server, Healthy: true})
		}

		parsedSite.RefreshPeriod = siteValue.RefreshPeriod
		if siteName == "default" {
			parsedSite.Domain = ""
			parsedSite.Path = "/"
			parsedSite.Port = uint16(rConfig.Global.ListeningPort)
		} else {
			if siteValue.Path == "" {
				siteValue.Path = "/"
			}
			if siteValue.Domain == "" {
				siteValue.Domain = ""
			}
			if siteValue.Port == 0 {
				siteValue.Port = int(pConfig.ListeningPort)
			}
			if siteValue.RefreshPeriod == 0 {
				siteValue.RefreshPeriod = 10 * time.Second
			}
			minRefreshPeriod, _ := time.ParseDuration("1s")
			if siteValue.RefreshPeriod < minRefreshPeriod {
				// TODO turn this into a Warn message
				return nil, errors.New(fmt.Sprintf("refresh period %v for site %s cannot be less than %v", siteValue.RefreshPeriod, siteName, minRefreshPeriod))
			}
			if !strings.HasPrefix(siteValue.Path, "/") {
				// TODO separate logging features
				// TODO add Warn message here for incorrect path prefix
				siteValue.Path = "/" + siteValue.Path
			}
			parsedSite.Path = siteValue.Path
			parsedSite.Domain = siteValue.Domain
			sitePort := siteValue.Port
			if sitePort == 0 {
				// TODO add Info message for port override with global
				sitePort = int(pConfig.ListeningPort)
			}
			if sitePort < 1 || sitePort > 65535 {
				return nil, errors.New(fmt.Sprintf("port number %d is out of range for site %s", sitePort, siteName))
			}
			parsedSite.Port = uint16(sitePort)
			parsedSite.RefreshPeriod = siteValue.RefreshPeriod
		}
		pConfig.Sites[siteName] = parsedSite
	}

	return &pConfig, nil
}

func StringConfig(c ParsedConfig) (string, error) {
	data, err := yaml.Marshal(c)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
