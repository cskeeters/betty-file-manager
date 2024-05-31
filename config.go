// This file code for reading the configuration file

package main

import (
	"os"
	"log"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Global variable with loaded config
var config Config


type WdReplacement struct {
	Real string
	Repl string
}

type Config struct {
	WdReplacement    []WdReplacement
}


func LoadConfig() {
	configPath := filepath.Join(home,".config/bfm/bfmrc")
	if _, err := os.Stat(configPath); err != nil {
		log.Print("Could not find bfmrc")
		return
	}

	_, err := toml.DecodeFile(configPath, &config)
	if err != nil {
		log.Print("Error parsing bfmrc")
	}
}
