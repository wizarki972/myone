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
	Logs: Logs{
		Level:           3,
		Panic:           false,
		DirectoryPath:   getDefaultLogsDirPath(),
		SaveLogsOnError: false,
		LogSaveInterval: 10,
	},
	Experimental: Experimental{
		UseSerialIDForASD: false,
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
