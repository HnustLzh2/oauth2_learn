package config

import (
	"flag"
	"os"

	"gopkg.in/yaml.v3"
)

var cfg App

func Setup() {
	path := flag.String("config", "./config.yaml", "path to config file")
	flag.Parse()

	if path == nil {
		panic("config file path is nil")
	}
	cfgBytes, err := os.ReadFile(*path)
	if err != nil {
		panic(err)
	}

	err = yaml.Unmarshal(cfgBytes, &cfg)
	if err != nil {
		panic(err)
	}
}
