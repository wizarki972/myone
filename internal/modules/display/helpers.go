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
	Focused bool   `json:"focused"`
	Name    string `json:"name"`
	Id      int    `json:"id"`
}

func ActiveMonitor() (string, error) {
	var monitors []hyprMonitor

	output := cmds.ExecCommandWithOutput(hyprlandMonitorsComamnd)
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
	maxi, err := strconv.ParseFloat(strings.TrimSpace(string(cmds.ExecCommandWithOutput("brightnessctl -d "+backlight_name+" m"))), 64)
	if err != nil {
		panic(err)
	}
	// current brightness
	curr, err := strconv.ParseFloat(strings.TrimSpace(string(cmds.ExecCommandWithOutput("brightnessctl -d "+backlight_name+" g"))), 64)
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
	cmds.ExecCommandNoFeedback(command)
}
