package audio

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/wizarki972/myone/internal/modules/display"
	"github.com/wizarki972/myone/internal/utils/cmds"
)

// swayosd icon for respective volume level
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

// swayosd for volume change
func NotifyVolume(device string) {

	if device == "source" {
		device = "@DEFAULT_SOURCE@"
	} else {
		device = "@DEFAULT_SINK@"
	}
	out, _ := cmds.ExecCommand(fmt.Sprintf("wpctl get-volume %s", device), false, true)
	current, _ := strconv.ParseFloat(strings.TrimPrefix(strings.TrimSpace(out), "Volume: "), 64)
	monitor, _ := display.ActiveMonitor()
	command := fmt.Sprintf("swayosd-client --monitor %s --custom-icon %s --custom-progress %.2f", monitor, getIcon(current*100.0), current)
	cmds.ExecCommand(command, false, false)
}
