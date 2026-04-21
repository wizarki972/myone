package battery

import (
	"log/slog"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/wizarki972/myone/internal/utils/cmds"
	"github.com/wizarki972/myone/internal/utils/fldir"
	"github.com/wizarki972/myone/internal/utils/logger"
)

var BatteryStates = struct {
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

var BatteryThreshold = 20

type Battery struct {
	BatteryPath      string
	IsBatteryPresent bool
	logg_book        *logger.LogBook
}

func NewBatteryMonitor(logg_book *logger.LogBook) *Battery {
	out := &Battery{
		logg_book: logg_book,
	}
	out.BatteryCheck()
	return out
}

func (b *Battery) BatteryCheck() {
	bats, err := filepath.Glob("/sys/class/power_supply/BAT*")
	if err != nil {
		b.logg_book.EnterLogAndPrint("Error in getting entries from /sys/class/power_supply/ that has BAT as prefix.", logger.LogTypes.Error, err)
	}

	for _, bat := range bats {
		if strings.TrimSpace(fldir.ReadFileAsStringNoError(bat+"/present")) == "1" && strings.TrimSpace(fldir.ReadFileAsStringNoError(bat+"/type")) == "Battery" {
			b.BatteryPath = bat
			b.IsBatteryPresent = true
		}
	}
}

func (b *Battery) RemainingBatteryPercent() int {
	level, err := strconv.Atoi(strings.TrimSpace(fldir.ReadFileAsStringNoError(filepath.Join(b.BatteryPath, "capacity"))))
	if err != nil {
		b.logg_book.EnterLogAndPrint("Error while converting battery value obtaned into integer format.", logger.LogTypes.Error, err)
	}
	return level
}

func (b *Battery) BatteryState() string {
	return fldir.ReadFileAsStringNoError(b.BatteryPath + "/status")
}

func (b *Battery) CheckAndNotify() error {
	level := b.RemainingBatteryPercent()
	if b.BatteryState() == BatteryStates.Discharging && level < BatteryThreshold {
		_, err := cmds.ExecCommand("notify-send -u critical '󱐋 Time to recharge!' 'Battery is down below 20%' -i battery-caution -t 30000", false, false)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *Battery) StartMonitor() {
	var err_count int
	for {
		if err := b.CheckAndNotify(); err != nil {
			slog.Error(err.Error())
			err_count += 1
		}
		time.Sleep(1 * time.Minute)
	}
}
