package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	mcli "github.com/jxskiss/mcli"
	toml "github.com/pelletier/go-toml/v2"
)

func main() {
	mcli.Add("login", loginCommand, "Login to CoffeeCup")

	mcli.AddGroup("projects", "This is a command group called cmd2")
	mcli.Add("projects list", projectsListCommand, "Do something with cmd2 sub1")
	mcli.Add("projects alias", projectsListCommand, "Do something with cmd2 sub1")

	// Enable shell auto-completion, see `program completion -h` for help.
	mcli.AddCompletion()

	mcli.Run()
}

func loginCommand() {
	var args struct {
		CompanyUrl string `cli:"#R, -c, --company, The prefix of the company's CoffeeCup instance (the \"amazon\" in \"amazon.coffeecup.app\")"`
		Username   string `cli:"#R, -u, --username, The username of the user"`
		Password   string `cli:"#R, -p, --password, The password of the user"`
	}
	mcli.Parse(&args)

	reqBody := url.Values{
		"grant_type": []string{"password"},
		"username":   []string{args.Username},
		"password":   []string{args.Password},
	}

	req, err := http.NewRequest("POST", "https://api.coffeecupapp.com/oauth2/token", strings.NewReader(reqBody.Encode()))
	if err != nil {
		panic(err)
	}

	req.Header.Set("Authorization", "Basic ZW1iZXI6cHVibGlj")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("companyurl", strings.Join([]string{"https://", args.CompanyUrl, ".coffeecup.app"}, ""))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var responseBody map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	if err != nil {
		panic(err)
	}

	storeToken(responseBody["access_token"].(string), responseBody["refresh_token"].(string))
	fmt.Printf("Successfully logged in as %s\n", args.Username)
}

func projectsListCommand() {
	req, err := http.NewRequest("GET", "https://api.coffeecupapp.com/v1/projects", nil)
	if err != nil {
		panic(err)
	}

	req.Header.Set("Authorization", "Bearer "+readToken())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var responseBody map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	if err != nil {
		panic(err)
	}

	for _, project := range responseBody["projects"].([]interface{}) {
		if p, ok := project.(map[string]interface{}); ok {
			fmt.Printf("%d: %s\n", int(p["id"].(float64)), p["name"])
		} else {
			fmt.Print("Error\n")
			// Handle the case where the element is not of type map[string]interface{}
		}
	}
}

func storeToken(accessToken string, refreshToken string) {
	cfg := readConfig()
	cfg.User.AccessToken = accessToken
	cfg.User.RefreshToken = refreshToken

	writeConfig(cfg)
}

type MyConfig struct {
	User struct {
		AccessToken  string
		RefreshToken string
	}
	// aliases map[string]string
}

func readToken() string {
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
