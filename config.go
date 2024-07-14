package main

import (
	"os"
	"path/filepath"

	toml "github.com/pelletier/go-toml/v2"
)

type ProjectConfig struct {
	Alias         string
	Name          string
	Id            int
	DefaultTaskId int
}

type Config struct {
	User struct {
		AccessToken  string
		RefreshToken string
		Id           int
	}
	Projects map[string]ProjectConfig
}

const (
	ConfigFolderPath = ".config/coffeecup"
	ConfigFileName   = "coffeecup.toml"
)

func StoreTokens(accessToken string, refreshToken string) error {
	cfg, _ := ReadConfig()
	cfg.User.AccessToken = accessToken
	cfg.User.RefreshToken = refreshToken

	return WriteConfig(cfg)
}

func GetAccessTokenFromConfig() string {
	cfg, _ := ReadConfig()
	return cfg.User.AccessToken
}

func GetRefreshTokenFromConfig() string {
	cfg, _ := ReadConfig()
	return cfg.User.RefreshToken
}

func StoreUserId(userId int) {
	cfg, _ := ReadConfig()
	cfg.User.Id = userId

	WriteConfig(cfg)
}

func GetUserIdFromConfig() int {
	cfg, _ := ReadConfig()
	return cfg.User.Id
}

func ReadConfig() (Config, error) {
	err := os.MkdirAll(filepath.Join(os.Getenv("HOME"), ConfigFolderPath), os.ModePerm)
	if err != nil {
		return Config{}, err
	}

	configFilepath := filepath.Join(os.Getenv("HOME"), ConfigFolderPath, ConfigFileName)
	configFile, err := os.ReadFile(configFilepath)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	err = toml.Unmarshal(configFile, &cfg)
	if err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func WriteConfig(cfg Config) error {
	updatedConfig, err := toml.Marshal(cfg)
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Join(os.Getenv("HOME"), ConfigFolderPath), os.ModePerm)
	if err != nil {
		return err
	}

	configFilepath := filepath.Join(os.Getenv("HOME"), ConfigFolderPath, ConfigFileName)
	err = os.WriteFile(configFilepath, updatedConfig, 0644)
	if err != nil {
		return err
	}

	return nil
}

func GetConfigPath() string {
	return filepath.Join(os.Getenv("HOME"), ConfigFolderPath, ConfigFileName)
}
