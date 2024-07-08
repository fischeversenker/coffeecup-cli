package main

import (
	"os"
	"path/filepath"

	toml "github.com/pelletier/go-toml/v2"
)

type Config struct {
	User struct {
		AccessToken  string
		RefreshToken string
		Id           int
		TeamId       int
	}
	Projects struct {
		Aliases map[string]int
	}
}

func StoreTokens(accessToken string, refreshToken string) {
	cfg := ReadConfig()
	cfg.User.AccessToken = accessToken
	cfg.User.RefreshToken = refreshToken

	WriteConfig(cfg)
}

func GetAccessTokenFromConfig() string {
	cfg := ReadConfig()
	return cfg.User.AccessToken
}

func GetRefreshTokenFromConfig() string {
	cfg := ReadConfig()
	return cfg.User.RefreshToken
}

func StoreUserId(userId int) {
	cfg := ReadConfig()
	cfg.User.Id = userId

	WriteConfig(cfg)
}

func GetUserIdFromConfig() int {
	cfg := ReadConfig()
	return cfg.User.Id
}

func storeTeamId(teamId int) {
	cfg := ReadConfig()
	cfg.User.TeamId = teamId

	WriteConfig(cfg)
}

func GetTeamIdFromConfig() int {
	cfg := ReadConfig()
	return cfg.User.TeamId
}

func ReadConfig() Config {
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

	var cfg Config
	err = toml.Unmarshal(configFile, &cfg)
	if err != nil {
		panic(err)
	}

	return cfg
}

func WriteConfig(cfg Config) {
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
