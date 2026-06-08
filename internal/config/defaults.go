package config

import (
	"os"
	"path/filepath"

	"github.com/wizarki972/myone/internal/common"
	"github.com/wizarki972/myone/internal/utils/fldir"
)

var DefaultConfig = Config{
	Battery: Battery{
		Threshold: 20,
	},
	Monitor: Monitor{
		Check_Backlight_For_ALL_Displays: false,
		Ignore_Apple_Studio_Displays:     true,
	},
	Logs: Logs{
		Level:              3,
		Panic:              false,
		Directory_Path:     getDefaultLogsDirPath(),
		Save_Logs_On_Error: false,
		Logs_Save_Interval: 1,
	},
	Experimental: Experimental{
		Use_Serial_ID_For_Apple_Studio_Displays: false,
	},
}

func getDefaultLogsDirPath() string {
	state := os.Getenv("XDG_STATE_HOME")
	if len(state) == 0 {
		return filepath.Join(fldir.GetHomeDir(), common.LOGS_DIR)
	} else {
		return filepath.Join(state, "myone/logs/")
	}
}
