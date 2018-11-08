package config

import (
	"log"

	"github.com/BurntSushi/toml"
)

type Config struct {
	To   To              `toml:"to"`
	From map[string]From `toml:"from"`
}

// for toml
type To struct {
	Token string `toml:"token"`
}

// for toml
type From struct {
	Token string `toml:"token"`
}

func LoadConfig(configPath string) (string, map[string]string, error) {
	var tomlConfig Config

	var toToken string
	var err error
	froms := map[string]string{}

	// load comfig file
	_, err = toml.DecodeFile(configPath, &tomlConfig)
	if err != nil {
		log.Println("[ERROR] loadConfig is fail", err)
		return "", nil, err
	}

	toToken = tomlConfig.To.Token

	for name, data := range tomlConfig.From {
		froms[name] = data.Token
	}

	return toToken, froms, err
}
