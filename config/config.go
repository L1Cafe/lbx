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

type ConfigParsed struct {
	Servers       []Server      `yaml:"servers"`
	RefreshPeriod time.Duration `yaml:"refresh_period"`
}

type ConfigRaw struct {
	Servers       []string      `yaml:"servers"`
	RefreshPeriod time.Duration `yaml:"refresh_period"`
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
	for _, s := range configRaw.Servers {
		configParsed.Servers = append(configParsed.Servers, Server{Url: s, Healthy: true})
	}
	configParsed.RefreshPeriod = configRaw.RefreshPeriod
	return &configParsed, nil
}
