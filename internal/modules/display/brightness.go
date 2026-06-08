package display

import (
	"encoding/json"
	"net"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/wizarki972/myone/internal/common"
	"github.com/wizarki972/myone/internal/services"
	"github.com/wizarki972/myone/internal/utils/cmds"
	"github.com/wizarki972/myone/internal/utils/fldir"
	"github.com/wizarki972/myone/internal/utils/logger"
	"github.com/wizarki972/myone/internal/utils/process"
)

// sends a signal to monitor daemon to change the brightness
func ChangeBrightness(value string, loggBook *logger.LogBook) {
	defaultMethod := func(msg string) {
		loggBook.EnterLogAndPrint(msg, logger.LogTypes.Warning, nil)
		loggBook.EnterLogAndPrint("Performing default brightnessctl operation, since socket not found.", logger.LogTypes.Info, nil)
		command := "brightnessctl s " + strings.TrimSpace(value)
		cmds.ExecCommandDetached(command)
	}
	runtimeDir, err := fldir.GetRuntimeDir()
	if err != nil {
		loggBook.EnterLogAndPrint(err.Error(), logger.LogTypes.Error, err)
		return
	}

	isRunning, _, err := process.IsOldProcessRunning(common.MONITOR_MON_PID_FILE_PATH)
	if err != nil {
		loggBook.EnterLogAndPrint(err.Error(), logger.LogTypes.Error, err)
		return
	}

	if isRunning {
		socketPath := filepath.Join(runtimeDir, common.DISPLAY_DEVICES_MONITOR_SOCKET)
		conn, err := net.Dial("unix", socketPath)
		if err != nil {
			defaultMethod(err.Error())
			return
		}
		defer conn.Close()

		activeMonitor, err := ActiveMonitor()
		if err != nil {
			loggBook.EnterLogAndPrint("Failed to get active/focused monitor.", logger.LogTypes.Error, err)
			return
		}
		command := services.MMCommand{Command: "brightness", TargetMonitor: activeMonitor}

		if (value[0] == '+') || (value[0] == '-') {
			command.Prefix = rune(value[0])
			value = value[1:]
		}

		value = strings.TrimSuffix(strings.TrimSpace(value), "%")
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			loggBook.EnterLogAndPrint("Failed to parse float value.", logger.LogTypes.Error, err)
			return
		}
		command.Value = floatVal

		if err = json.NewEncoder(conn).Encode(command); err != nil {
			defaultMethod(err.Error())
		}
	} else {
		defaultMethod("socket not found")
	}
}
