package config

import (
	"github.com/BurntSushi/toml"
	"github.com/wizarki972/myone/internal/utils/fldir"
)

// saves default config
func SaveConfig() {
	file := fldir.CreateFile(DefaultConfig.general.configPath)
	defer file.Close()

	if err := toml.NewEncoder(file).Encode(DefaultConfig); err != nil {
		Log("Failed to save the config file to this location - "+DefaultConfig.general.configPath, "ERROR", err)
	}

	Log("Default config is saved here - "+DefaultConfig.general.configPath, "INFO", nil)
}

// reads user if found, else returns default
func GetConfig() *Config {
	var readConfig = DefaultConfig
	if _, err := toml.DecodeFile(readConfig.general.configPath, &readConfig); err != nil {
		return &readConfig
	}
	return &readConfig
}
