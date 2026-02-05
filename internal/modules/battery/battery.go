package battery

import (
	"log/slog"
	"path/filepath"
	"strconv"
	"time"

	"github.com/wizarki972/myone/internal/utils/cmds"
	"github.com/wizarki972/myone/internal/utils/fldir"
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
}

func NewBatteryMonitor() *Battery {
	out := &Battery{}
	out.BatteryCheck()
	return out
}

func (b *Battery) BatteryCheck() {
	bats, err := filepath.Glob("/sys/class/power_supply/BAT*")
	if err != nil {
		panic(err)
	}

	for _, bat := range bats {
		if fldir.ReadFileAsStringNoError(bat+"/present") == "1" && fldir.ReadFileAsStringNoError(bat+"/type") == "Battery" {
			b.BatteryPath = bat
			b.IsBatteryPresent = true
		}
	}
}

func (b *Battery) RemainingBatteryPercent() int {
	level, err := strconv.Atoi(fldir.ReadFileAsStringNoError(b.BatteryPath + "/capacity"))
	if err != nil {
		panic(err)
	}
	return level
}

func (b *Battery) BatteryState() string {
	return fldir.ReadFileAsStringNoError(b.BatteryPath + "/status")
}

func (b *Battery) CheckAndNotify() error {
	level := b.RemainingBatteryPercent()
	if b.BatteryState() == BatteryStates.Discharging && level < BatteryThreshold {
		err := cmds.ExecComamndWithError("notify-send -u critical '󱐋 Time to recharge!' 'Battery is down below 20%' -i battery-caution -t 30000")
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
