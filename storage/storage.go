package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type AccountData struct {
	Address   string `json:"address"`
	Password  string `json:"password"`
	Token     string `json:"token"`
	AccountID string `json:"accountId"`
	CreatedAt string `json:"createdAt"`
}

var (
	configPath     string
	configPathOnce sync.Once
	configPathErr  error
)

func getConfigPath() (string, error) {
	configPathOnce.Do(func() {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			configPathErr = err
			return
		}
		configPath = filepath.Join(homeDir, ".burnmail.json")
	})
	return configPath, configPathErr
}

func Save(data *AccountData) error {
	path, err := getConfigPath()
	if err != nil {
		return err
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, jsonData, 0600)
}

func Load() (*AccountData, error) {
	path, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var account AccountData
	if err := json.Unmarshal(data, &account); err != nil {
		return nil, err
	}

	return &account, nil
}

func Delete() error {
	path, err := getConfigPath()
	if err != nil {
		return err
	}

	err = os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

func Exists() bool {
	path, err := getConfigPath()
	if err != nil {
		return false
	}

	_, err = os.Stat(path)
	return err == nil
}
