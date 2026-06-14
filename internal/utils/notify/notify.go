package notify

import (
	"fmt"

	"github.com/wizarki972/myone/internal/utils/cmds"
)

type URGENCY string
type CATEGORY string

// app name to be used for messages.
const APP_NAME = "myone"

// notification catgory constant values
const CATG_OSD CATEGORY = "osd"
const CATG_URG_LOW CATEGORY = "urgency_low"
const CATG_URG_NORMAL CATEGORY = "urgency_normal"
const CATG_URG_CRITICAL CATEGORY = "urgency_critical"

// to send notification without progress bar.
func Notify(category CATEGORY, summary, body string) {
	NotifyWithProgressBar(-1, category, summary, body)
}

// sends a notification with progress bar, progressbarValue value range is [0 - 100]
func NotifyWithProgressBar(progressBarValue int, category CATEGORY, summary, body string) {
	var command, appID string
	if category == CATG_OSD {
		appID = "-r 007"
	} else if len(category) == 0 {
		category = CATG_URG_NORMAL
	}
	if progressBarValue >= 0 {
		command = fmt.Sprintf("notify-send -a %s --hint=int:value:%d %s -c %s \"%s\" \"%s\"", APP_NAME, min(progressBarValue, 100), appID, category, summary, body)
	} else {
		command = fmt.Sprintf("notify-send -a %s -c %s \"%s\" \"%s\"", APP_NAME, category, summary, body)
	}

	cmds.ExecCommand(command, false, false)
}
