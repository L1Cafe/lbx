package main

import (
	"github.com/L1Cafe/lbx/config"
	"testing"
)

func TestReadConfig(t *testing.T) {
	_, err := config.LoadConfig("config_test.yaml")
	if err != nil {
		t.Errorf("Loading config not successful: %s", err.Error())
	}
}
