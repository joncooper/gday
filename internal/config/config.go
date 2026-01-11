package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const (
	configDir       = ".gday"
	credentialsFile = "credentials.json"
	tokenFile       = "token.json"
)

// Config holds the application configuration
type Config struct {
	ConfigDir string
}

// GetConfigDir returns the path to the config directory
func GetConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, configDir)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	return dir, nil
}

// GetCredentialsPath returns the path to the OAuth credentials file
func GetCredentialsPath() (string, error) {
	dir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, credentialsFile), nil
}

// GetTokenPath returns the path to the OAuth token file
func GetTokenPath() (string, error) {
	dir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, tokenFile), nil
}

// CredentialsExist checks if OAuth credentials have been configured
func CredentialsExist() bool {
	path, err := GetCredentialsPath()
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}

// TokenExists checks if a valid token file exists
func TokenExists() bool {
	path, err := GetTokenPath()
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}

// ReadCredentials reads the OAuth credentials from file
func ReadCredentials() ([]byte, error) {
	path, err := GetCredentialsPath()
	if err != nil {
		return nil, err
	}
	return os.ReadFile(path)
}

// SaveCredentials saves OAuth credentials to file
func SaveCredentials(data []byte) error {
	path, err := GetCredentialsPath()
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// ReadToken reads the OAuth token from file
func ReadToken() ([]byte, error) {
	path, err := GetTokenPath()
	if err != nil {
		return nil, err
	}
	return os.ReadFile(path)
}

// SaveToken saves OAuth token to file
func SaveToken(token interface{}) error {
	path, err := GetTokenPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// DeleteToken removes the cached token
func DeleteToken() error {
	path, err := GetTokenPath()
	if err != nil {
		return err
	}
	return os.Remove(path)
}
