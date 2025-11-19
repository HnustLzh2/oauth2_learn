package config_test

import (
	"oauth2/config"
	"testing"
)

func TestConfigTypeConv(t *testing.T) {
	err := config.ConfigTypeConv("../config.json", "../config.yaml")
	if err != nil {
		t.Error("config type conv failed:", err)
	}
	cfg := config.GetCfg()
	t.Log("config.yaml:", cfg)
}
