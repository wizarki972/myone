package cmd

import (
	"fmt"
	"strings"

	"github.com/wizarki972/myone/internal/config"
	"github.com/wizarki972/myone/internal/utils/fldir"
	"github.com/wizarki972/myone/internal/utils/logger"
)

func handleLogg() *logger.LogBook {
	var userConfig = config.GetConfig()
	var loggerInstance *logger.LogBook
	switch {
	case len(logPath) > 0:
		if !strings.HasPrefix(logPath, fldir.GetHomeDir()) {
			fmt.Println("-> [WARN] The logpath is outside of user's home directory, make sure to have necessary permission for saving the log.")
		}
		loggerInstance = logger.NewLogBook(logPath, true, true, userConfig)
	case saveLog:
		loggerInstance = logger.NewLogBook("", true, true, userConfig)
	default:
		loggerInstance = logger.NewLogBook("", false, userConfig.Logs.SaveLogsOnError, userConfig)
	}
	return loggerInstance
}
