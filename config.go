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

type Binding struct {
	Key         string `toml:"key"`
	Command     string `toml:"command"`
}

type Plugin struct {
	Section string `toml:"section"`
	Command string `toml:"command"`
	Help    string `toml:"help"`
}

type Config struct {
	DefaultPlugins     bool              `toml:"default_plugins"`
	DefaultBindings    bool              `toml:"default_bindings"`
	Plugins            []Plugin          `toml:"plugins"`
	Bindings           []Binding         `toml:"bindings"`
	WdReplacements     []WdReplacement   `toml:"wd_replacements"`
}

func LoadConfig() {
	log.Print("Loading Config")

	configPath := filepath.Join(home,".config/bfm/bfmrc")
	if _, err := os.Stat(configPath); err != nil {
		log.Print("Could not find bfmrc")
		SetDefaultBindings()
		return
	}

	_, err := toml.DecodeFile(configPath, &config)
	if err != nil {
		log.Print("Error parsing bfmrc")
		SetDefaultBindings()
		return
	}

	if config.DefaultPlugins {
		log.Print("Setting Default Plugins")
		SetDefaultPlugins()
	}

	if config.DefaultBindings {
		log.Print("Setting Default Bindings")
		SetDefaultBindings()
	}
}
