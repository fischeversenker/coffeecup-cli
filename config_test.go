package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadConfig(t *testing.T) {
	// Create a temporary test directory
	tempDir := t.TempDir()

	// Set the HOME environment variable to the temporary directory
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	// Create the config file
	configFilepath := filepath.Join(tempDir, ".config", "coffeecup", "coffeecup.toml")
	configContent := []byte(`
		[User]
		AccessToken  = "test-access-token"
		RefreshToken = "test-refresh-token"
		Id           = 123

		[Projects]
		[Projects.project1]
		Alias         = "project1"
		Id            = 456
		DefaultTaskId = 789
	`)
	err := os.MkdirAll(filepath.Dir(configFilepath), os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(configFilepath, configContent, 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Call the ReadConfig function
	cfg, _ := ReadConfig()

	// Verify the expected values
	expectedAccessToken := "test-access-token"
	if cfg.User.AccessToken != expectedAccessToken {
		t.Errorf("Expected access token %q, got %q", expectedAccessToken, cfg.User.AccessToken)
	}

	expectedRefreshToken := "test-refresh-token"
	if cfg.User.RefreshToken != expectedRefreshToken {
		t.Errorf("Expected refresh token %q, got %q", expectedRefreshToken, cfg.User.RefreshToken)
	}

	expectedUserId := 123
	if cfg.User.Id != expectedUserId {
		t.Errorf("Expected user ID %d, got %d", expectedUserId, cfg.User.Id)
	}

	expectedAlias := "project1"
	if cfg.Projects["project1"].Alias != expectedAlias {
		t.Errorf("Expected project alias %q, got %q", expectedAlias, cfg.Projects["project1"].Alias)
	}

	expectedProjectId := 456
	if cfg.Projects["project1"].Id != expectedProjectId {
		t.Errorf("Expected project ID %d, got %d", expectedProjectId, cfg.Projects["project1"].Id)
	}

	expectedTaskId := 789
	if cfg.Projects["project1"].DefaultTaskId != expectedTaskId {
		t.Errorf("Expected task ID %d, got %d", expectedTaskId, cfg.Projects["project1"].DefaultTaskId)
	}
}

func TestReadNotExistingConfigError(t *testing.T) {
	// Create a temporary test directory
	tempDir := t.TempDir()

	// Set the HOME environment variable to the temporary directory
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	// Don't create the config file

	// Call the ReadConfig function
	_, err := ReadConfig()

	if err == nil {
		t.Error("Expected an error, got nil")
	}
}
