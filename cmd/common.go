package cmd

import (
	"github.com/wizarki972/myone/internal/config"
	"github.com/wizarki972/myone/internal/utils/fldir"
	"github.com/wizarki972/myone/internal/utils/logger"
)

func handleLogg() *logger.LogBook {
	var userConfig = config.GetConfig()
	var loggerInstance *logger.LogBook
	switch {
	case len(logPath) > 0:
		if fldir.IsPathExist(logPath) {
			logger.Log("Provided log path is occupied.", logger.LogTypes.Error, nil)
		} else {
			loggerInstance = logger.NewLogBook(logPath, true, true, userConfig)
		}
	case saveLog:
		loggerInstance = logger.NewLogBook("", true, true, userConfig)
	default:
		loggerInstance = logger.NewLogBook("", false, userConfig.Logs.SaveLogsOnError, userConfig)
	}
	return loggerInstance
}
