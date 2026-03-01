package config

import (
	"fmt"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/wizarki972/myone/internal/utils/fldir"
	"github.com/wizarki972/myone/internal/utils/user"
)

const DEFAULT_THEME = "tokyonight"

var Default_Config = Config{
	general: General{
		config_path: filepath.Join(user.GetHomeDir(), ".config/myone/config.toml"),
	},
}

func SaveConfig() {
	file := fldir.CreateFile(Default_Config.general.config_path)
	defer file.Close()

	if err := toml.NewEncoder(file).Encode(Default_Config); err != nil {
		panic(err)
	}

	fmt.Println("config saved successfully...")
}

func ReadConfig() *Config {
	var readConfig Config
	if _, err := toml.DecodeFile(Default_Config.general.config_path, &readConfig); err != nil {
		panic(err)
	}

	return &readConfig
}
