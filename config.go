package main

import (
	"os"
	"path/filepath"

	toml "github.com/pelletier/go-toml/v2"
)

type MyConfig struct {
	User struct {
		AccessToken  string
		RefreshToken string
	}
	Projects struct {
		Aliases map[string]string
	}
}

func storeTokens(accessToken string, refreshToken string) {
	cfg := readConfig()
	cfg.User.AccessToken = accessToken
	cfg.User.RefreshToken = refreshToken

	writeConfig(cfg)
}

func getAccessToken() string {
	cfg := readConfig()
	return cfg.User.AccessToken
}

func readConfig() MyConfig {
	configFolderpath := filepath.Join(os.Getenv("HOME"), ".config", "coffeecup")
	err := os.MkdirAll(configFolderpath, os.ModePerm)
	if err != nil {
		panic(err)
	}

	configFilepath := filepath.Join(configFolderpath, "config.toml")
	configFile, err := os.ReadFile(configFilepath)
	if err != nil {
		panic(err)
	}

	var cfg MyConfig
	err = toml.Unmarshal(configFile, &cfg)
	if err != nil {
		panic(err)
	}

	return cfg
}

func writeConfig(cfg MyConfig) {
	updatedConfig, err := toml.Marshal(cfg)
	if err != nil {
		panic(err)
	}

	configFolderpath := filepath.Join(os.Getenv("HOME"), ".config", "coffeecup")
	err = os.MkdirAll(configFolderpath, os.ModePerm)
	if err != nil {
		panic(err)
	}

	configFilepath := filepath.Join(configFolderpath, "config.toml")
	err = os.WriteFile(configFilepath, updatedConfig, 0644)
	if err != nil {
		panic(err)
	}
}
