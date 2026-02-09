package audio

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/wizarki972/myone/internal/modules/display"
	"github.com/wizarki972/myone/internal/utils/cmds"
)

func getIcon(value float64) string {
	switch {
	case value == -999:
		return "audio-volume-muted"
	case value <= 30:
		return "audio-volume-low"
	case value <= 60:
		return "audio-volume-medium"
	default:
		return "audio-volume-high"
	}
}

func NotifyVolume(device string) {

	if device == "source" {
		device = "@DEFAULT_SOURCE@"
	} else {
		device = "@DEFAULT_SINK@"
	}

	current, _ := strconv.ParseFloat(strings.TrimPrefix(strings.TrimSpace(string(cmds.ExecCommandWithOutput(fmt.Sprintf("wpctl get-volume %s", device)))), "Volume: "), 64)
	monitor, _ := display.ActiveMonitor()
	command := fmt.Sprintf("swayosd-client --monitor %s --custom-icon %s --custom-progress %.2f", monitor, getIcon(current*100.0), current)
	cmds.ExecCommandNoFeedback(command)
}
