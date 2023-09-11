package main

import (
	"fmt"
	"github.com/L1Cafe/lbx/config"
	"github.com/L1Cafe/lbx/log"
	"github.com/L1Cafe/lbx/site"
)

var appConfig *config.ParsedConfig

func main() {
	c, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Wrapper(log.Fatal, fmt.Sprintf("Error loading config: %s", err.Error()))
	}
	appConfig = c
	log.Init(c.LogLevel)
	site.Init(c)
}
