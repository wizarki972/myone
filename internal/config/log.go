package config

import (
	"fmt"
	"os"
)

// Log method to avoid import cycle error in go
func Log(log string, logType string, err error) {
	userConfig := GetConfig()
	if len(log) == 0 {
		fmt.Printf("-> [%s] IF YOU ARE SEEING THIS ERROR THAN THAT MEANS AN EMPTY LOG WAS PROVIDED.\n", logType)
		return
	}
	fmt.Printf("-> %s\n", log)

	if logType == "ERROR" {
		if userConfig.Logs.Panic && err != nil {
			panic(err)
		}
		os.Exit(1)
	}
}
