package config

import (
	"errors"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/wizarki972/myone/internal/common"
	"github.com/wizarki972/myone/internal/utils/fldir"
)

// saves default config
func SaveDefaultConfig() error {
	configPath := filepath.Join(fldir.GetHomeDir(), common.CONFIG_PATH)
	file, err := fldir.CreateFile(configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := toml.NewEncoder(file).Encode(DefaultConfig); err != nil {
		return errors.New("Failed to save the config file to this location - " + configPath + ". Full Error => " + err.Error())
	}
	return nil
}

// reads user if found, else returns default
func GetConfig() *Config {
	var readConfig = DefaultConfig
	if _, err := toml.DecodeFile(filepath.Join(fldir.GetHomeDir(), common.CONFIG_PATH), &readConfig); err != nil {
		return &readConfig
	}
	return &readConfig
}
