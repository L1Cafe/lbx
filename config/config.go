package config

import (
	"errors"
	"io/ioutil"
	"strconv"
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
	Port string `yaml:"port"`
}

type SiteConfigParsed struct {
	// TODO this needs a mutex for thread safety when serving several sites
	Servers       []Server
	RefreshPeriod time.Duration
	Domain        string
	Path          string
	Port          uint16
}

// ParsedConfig is the actual configuration that the application uses
type ParsedConfig struct {
	ListeningPort int
	LogLevel      int
	Sites         map[string]SiteConfigParsed
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
	pConfig.Sites = make(map[string]SiteConfigParsed)
	for siteName, siteValue := range rConfig.Sites {
		var parsedSite SiteConfigParsed
		for _, server := range siteValue.Servers {
			parsedSite.Servers = append(parsedSite.Servers, Server{Url: server, Healthy: true})
		}
		parsedSite.RefreshPeriod = siteValue.RefreshPeriod
		if siteName == "default" {
			parsedSite.Domain = "default"
			parsedSite.Path = "/"
			parsedSite.Port = uint16(rConfig.Global.ListeningPort)
		} else {
			parsedSite.Domain = siteValue.Domain
			parsedSite.Path = siteValue.Path
			parsedPort, err := strconv.Atoi(siteValue.Port)
			if err == nil {
				return nil, err
			}
			if parsedPort < 1 || parsedPort > 65535 {
				return nil, errors.New("port number is out of range")
			}
			parsedSite.Port = uint16(parsedPort)
		}
		pConfig.Sites[siteName] = parsedSite
	}
	pConfig.LogLevel = rConfig.Global.LogLevel
	pConfig.ListeningPort = rConfig.Global.ListeningPort
	return &pConfig, nil
}

func StringConfig(c ParsedConfig) (string, error) {
	data, err := yaml.Marshal(c)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
