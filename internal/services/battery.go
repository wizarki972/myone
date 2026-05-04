package services

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/wizarki972/myone/internal/common"
	"github.com/wizarki972/myone/internal/config"
	"github.com/wizarki972/myone/internal/utils/cmds"
	"github.com/wizarki972/myone/internal/utils/fldir"
	"github.com/wizarki972/myone/internal/utils/logger"
)

// All battery states like Discharging, Full, Charging, NotCharging
var stati = struct {
	Discharging string
	Full        string
	Charging    string
	NotCharging string
}{
	Discharging: "Discharging",
	Full:        "Full",
	Charging:    "Charging",
	NotCharging: "Not charging",
}

type BattMon struct {
	config           *config.Config
	isBatteryPresent bool
	chargePath       string
	statusPath       string
	loggBook         *logger.LogBook
}

// If a battery is present, the first found is used, if battery not found the service never starts
func NewBattMon(loggBook *logger.LogBook, userConfig *config.Config) *BattMon {
	var err error
	entries, err := filepath.Glob("/sys/class/power_supply/BAT*")
	if err != nil {
		loggBook.EnterLogAndPrint("Error in getting entries from /sys/class/power_supply/ that has BAT as prefix.", logger.LogTypes.Error, err)
	}

	for _, entry := range entries {
		var present, batteryType string
		present, err = fldir.ReadFileAsString(filepath.Join(entry, "present"))
		if err != nil {
			loggBook.EnterLogAndPrint("Failed to read battery present status from path "+entry+"/present", logger.LogTypes.Error, err)
		}
		batteryType, err = fldir.ReadFileAsString(filepath.Join(entry, "type"))
		if err != nil {
			loggBook.EnterLogAndPrint("Failed to read battery type from path "+entry+"/type", logger.LogTypes.Error, err)
		}

		if present == "1" && batteryType == "Battery" {
			return &BattMon{
				config:           userConfig,
				isBatteryPresent: true,
				chargePath:       filepath.Join(entry, "capacity"),
				statusPath:       filepath.Join(entry, "status"),
				loggBook:         loggBook,
			}
		}
	}

	loggBook.EnterLogAndPrint("No Battery Found: service exited...", logger.LogTypes.Error, errors.New("no Battery Found: service exited"))
	return nil
}

// gets the remaining charge of the battery
func (bm *BattMon) GetChargeLeft() int {
	if !bm.isBatteryPresent {
		bm.loggBook.EnterLogAndPrint("No Battery Found: service exited...", logger.LogTypes.Error, errors.New("no Battery Found: service exited"))
	}

	data, err := fldir.ReadFileAsString(bm.chargePath)
	if err != nil {
		bm.loggBook.EnterLogAndPrint(err.Error(), logger.LogTypes.Error, err)
	}

	remBatt, err := strconv.Atoi(data)
	if err != nil {
		bm.loggBook.EnterLogAndPrint(err.Error(), logger.LogTypes.Error, err)
	}

	return remBatt
}

// gets the current state of the battery, i.e. Charging, Discharging, Full, Notcharging
func (bm *BattMon) GetStatus() string {
	state, err := fldir.ReadFileAsString(bm.statusPath)
	if err != nil {
		bm.loggBook.EnterLogAndPrint(err.Error(), logger.LogTypes.Error, err)
	}
	return state
}

// Monitors the battery level monitor, alerts when below the threshold value.
func (bm *BattMon) StartService() {
	isRunning, pid, err := isOldProcessRunning(common.BATT_MON_PID_FILE_NAME)
	if err != nil {
		bm.loggBook.EnterLogAndPrint("Cannot determine whether an old precess is running or not.", logger.LogTypes.Warning, nil)
	}

	if isRunning {
		if err = killProcess(pid); err != nil {
			bm.loggBook.EnterLogAndPrint("Failed to kill already running battery monitor service.", logger.LogTypes.Error, err)
		}
		return
	}

	if err := savePID(common.BATT_MON_PID_FILE_NAME, os.Getpid()); err != nil {
		bm.loggBook.EnterLogAndPrint("Failed to save PID  so the current service is stopped.", logger.LogTypes.Error, err)
	}

	bm.loggBook.EnterLogAndPrint("Started battery monitor...", logger.LogTypes.Info, nil)
	for {
		remBatt := bm.GetChargeLeft()
		state := bm.GetStatus()
		if state == stati.Discharging && remBatt <= bm.config.Battery.Threshold {
			command := fmt.Sprintf("notify-send -u critical '󱐋 Time to charge!' 'Current battery level is %d%%' -i battery-caution -t 30000", config.DefaultConfig.Battery.Threshold)
			_, err := cmds.ExecCommand(command, false, false)
			if err != nil {
				bm.loggBook.EnterLogAndPrint(err.Error(), logger.LogTypes.Error, err)
			}
		}
		time.Sleep(45 * time.Second)
	}
}
