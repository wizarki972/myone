package display

import (
	"encoding/json"
	"errors"

	"github.com/wizarki972/myone/internal/common"
	"github.com/wizarki972/myone/internal/utils/cmds"
)

type hyprMonitor struct {
	Id      int     `json:"id"`
	Name    string  `json:"name"`
	Width   int     `json:"width"`
	Height  int     `json:"height"`
	Scale   float64 `json:"scale"`
	Focused bool    `json:"focused"`
}

type HyprOption struct {
	Option string `json:"option"`
	Int    int    `json:"int"`
	Set    bool   `json:"set"`
}

// gets active monitor using hyprctl
func ActiveMonitor() (string, error) {
	var monitors []hyprMonitor

	output, _ := cmds.ExecCommandBytes(common.HYPRCTL_MONITORS_CMD, true)
	if err := json.Unmarshal(output, &monitors); err != nil {
		panic(err)
	}

	for _, monitor := range monitors {
		if monitor.Focused {
			return monitor.Name, nil
		}
	}
	return "", errors.New("no active monitor found...")
}

// gets screen resolution using hyprctl
func GetScreenResolution() (int, int, int, error) {
	var monitors []hyprMonitor
	output, _ := cmds.ExecCommandBytes(common.HYPRCTL_MONITORS_CMD, true)
	if err := json.Unmarshal(output, &monitors); err != nil {
		return 0, 0, 0, err
	}

	for _, monitor := range monitors {
		if monitor.Focused {
			return monitor.Width, monitor.Height, int(monitor.Scale), nil
		}
	}

	return 0, 0, 0, errors.New("focused monitor not found")
}

// gets border thickness value from hyprland config
func GetHyprBorder() (int, error) {
	command := "hyprctl -j getoption decoration:rounding"
	output, _ := cmds.ExecCommandBytes(command, true)

	var hyprOption HyprOption
	if err := json.Unmarshal([]byte(output), &hyprOption); err != nil {
		return 0, err
	}

	return hyprOption.Int, nil
}
