package display

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

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

func ActiveMonitor() (string, error) {
	var monitors []hyprMonitor

	output, _ := cmds.Exec_cmd_bytes(hyprlandMonitorsComamnd, true)
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

func swayOSDNotify(backlight_name string) {
	// maximum brightness
	out, _ := cmds.Exec_cmd("brightnessctl -d "+backlight_name+" m", false, true, false)
	maxi, err := strconv.ParseFloat(strings.TrimSpace(out), 64)
	if err != nil {
		panic(err)
	}
	// current brightness
	out, _ = cmds.Exec_cmd("brightnessctl -d "+backlight_name+" g", false, true, false)
	curr, err := strconv.ParseFloat(strings.TrimSpace(out), 64)
	if err != nil {
		panic(err)
	}
	percent := curr / maxi

	// focused monitor
	name, err := ActiveMonitor()
	if err != nil {
		panic("error while getting focused monitor for osd")
	}

	// osd command
	command := fmt.Sprintf("swayosd-client --monitor %s --custom-icon display-brightness --custom-progress %.2f --custom-progress-text '%.0f%%'", name, max(0.01, percent), percent*100)
	cmds.Exec_cmd(command, false, false, false)
}

func GetScreenResolution() []int {
	var monitors []hyprMonitor
	output, _ := cmds.Exec_cmd_bytes(hyprlandMonitorsComamnd, true)
	if err := json.Unmarshal(output, &monitors); err != nil {
		panic(err)
	}

	for _, monitor := range monitors {
		if monitor.Focused {
			return []int{monitor.Width, monitor.Height, int(monitor.Scale)}
		}
	}

	panic(errors.New("focused monitor not found"))
}

func GetHyprBorder() int {
	command := "hyprctl -j getoption decoration:rounding"
	output, _ := cmds.Exec_cmd_bytes(command, true)

	var hyprOption HyprOption
	if err := json.Unmarshal([]byte(output), &hyprOption); err != nil {
		panic(err)
	}

	return hyprOption.Int
}
