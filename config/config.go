package config

import (
	"io/ioutil"
	"time"

	"gopkg.in/yaml.v2"
)

type Server struct {
	Url     string
	Healthy bool
}

type GlobalConfig struct {
	ListeningPort int `yaml:"listening_port"`
	LogLevel      int `yaml:"log_level"`
}

type SiteConfigRaw struct {
	Servers       []string      `yaml:"servers"`
	RefreshPeriod time.Duration `yaml:"refresh_period"`
}

type SiteConfigParsed struct {
	Servers       []Server
	RefreshPeriod time.Duration
}

type ConfigParsed struct {
	ListeningPort int
	LogLevel      int
	Sites         map[string]SiteConfigParsed
}

type ConfigRaw struct {
	Global GlobalConfig             `yaml:"global"`
	Sites  map[string]SiteConfigRaw `yaml:"sites"`
}

func LoadConfig(file string) (*ConfigParsed, error) {
	var configRaw ConfigRaw
	var configParsed ConfigParsed
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(data, &configRaw)
	if err != nil {
		return nil, err
	}

	// Config parsing
	configParsed.Sites = make(map[string]SiteConfigParsed)
	for siteKey, siteValue := range configRaw.Sites {
		var parsedSite SiteConfigParsed
		parsedSite.RefreshPeriod = siteValue.RefreshPeriod
		for _, server := range siteValue.Servers {
			parsedSite.Servers = append(parsedSite.Servers, Server{Url: server, Healthy: true})
		}
		configParsed.Sites[siteKey] = parsedSite
	}
	configParsed.LogLevel = configRaw.Global.LogLevel
	configParsed.ListeningPort = configRaw.Global.ListeningPort
	return &configParsed, nil
}

func StringConfig(c ConfigParsed) (string, error) {
	data, err := yaml.Marshal(c)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
