package main

import (
	"os"
	"path/filepath"
	"time"

	toml "github.com/pelletier/go-toml/v2"
)

type ProjectConfig struct {
	Alias         string
	Name          string
	Id            int
	DefaultTaskId int
}

type JiraConfig struct {
	Enabled      bool
	TicketPrefix string
}

type Config struct {
	User struct {
		AccessToken  string
		ExpiresAt    int64
		RefreshToken string
		Id           int
		Company      string
	}
	Projects map[string]ProjectConfig
	Jira     JiraConfig
}

const (
	ConfigFolderPath = ".config/aerion"
	ConfigFileName   = "config.toml"
)

func StoreTokens(accessToken string, refreshToken string, expiresIn int) error {
	cfg, _ := ReadConfig()
	cfg.User.AccessToken = accessToken
	cfg.User.RefreshToken = refreshToken

	expiresAt := time.Now().Unix() + int64(expiresIn)*1000
	cfg.User.ExpiresAt = expiresAt

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

func StoreCompany(company string) {
	cfg, _ := ReadConfig()
	cfg.User.Company = company

	WriteConfig(cfg)
}

func GetCompanyFromConfig() string {
	cfg, _ := ReadConfig()
	return cfg.User.Company
}

func ReadConfig() (Config, error) {
	err := os.MkdirAll(filepath.Join(os.Getenv("HOME"), ConfigFolderPath), os.ModePerm)
	if err != nil {
		return Config{}, err
	}

	configFilepath := GetConfigPath()
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

	configFilepath := GetConfigPath()
	err = os.WriteFile(configFilepath, updatedConfig, 0644)
	if err != nil {
		return err
	}

	return nil
}

func GetConfigPath() string {
	return filepath.Join(os.Getenv("HOME"), ConfigFolderPath, ConfigFileName)
}
