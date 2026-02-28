package config

import (
	"fmt"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/wizarki972/myone/internal/utils/fldir"
	"github.com/wizarki972/myone/internal/utils/user"
)

const MYONE_ASCII = `
 _____ ______       ___    ___ ________  ________   _______      
|\   _ \  _   \    |\  \  /  /|\   __  \|\   ___  \|\  ___ \     
\ \  \\\__\ \  \   \ \  \/  / | \  \|\  \ \  \\ \  \ \   __/|    
 \ \  \\|__| \  \   \ \    / / \ \  \\\  \ \  \\ \  \ \  \_|/__  
  \ \  \    \ \  \   \/  /  /   \ \  \\\  \ \  \\ \  \ \  \_|\ \ 
   \ \__\    \ \__\__/  / /      \ \_______\ \__\\ \__\ \_______\
    \|__|     \|__|\___/ /        \|_______|\|__| \|__|\|_______|
                  \|___|/                                        
                                                                 
`
const VERSION = "0.7.51-alpha"
const VERSION_INT = 7.51
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
