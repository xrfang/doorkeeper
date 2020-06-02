package main

import (
	"os"

	"gopkg.in/yaml.v2"
)

type config struct {
}

var cf config

func loadConfig(fn string) {
	f, err := os.Open(fn)
	assert(err)
	defer f.Close()
	assert(yaml.NewDecoder(f).Decode(&cf))
	//TODO: set default values.
}
