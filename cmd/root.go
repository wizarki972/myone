package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wizarki972/myone/internal/modules/audio"
	"github.com/wizarki972/myone/internal/modules/battery"
	"github.com/wizarki972/myone/internal/modules/display"
	"github.com/wizarki972/myone/internal/modules/logout"
	"github.com/wizarki972/myone/internal/modules/screenshot"
)

const VERSION = "0.5.5-alpha"

var brightness string
var log_out int
var vol_notify, screen_shot, monitor_daemon, batt_mon, version bool

var rootCMD = &cobra.Command{
	Use:   "myone",
	Short: "my one utility for my needs",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(brightness) > 0 {
			display.ChangeBrightness(brightness)
		}

		if log_out > 0 {
			logout.Logout(min(log_out, 2))
		}

		if vol_notify {
			audio.NotifyVolume()
		}

		if screen_shot {
			screenshot.OpenGUI()
		}

		if batt_mon {
			battery.NewBatteryMonitor().StartMonitor()
		}

		if monitor_daemon {
			display.NewMonitorDaemon().StartDaemon()
		}

		if version {
			fmt.Println(VERSION)
		}

		return nil
	},
}

func initializeFlags() {
	rootCMD.Flags().StringVar(&brightness, "bright", "", "+5% - increases the brightness by 5%, \n-5% decreases the brightness by 5%")

	rootCMD.Flags().IntVar(&log_out, "logout", 0, "accepted values 1, 2. Displays power menu.")

	rootCMD.Flags().BoolVar(&vol_notify, "volume-osd", false, "just tells swayosd to show current volume level of the current sink.")

	rootCMD.Flags().BoolVar(&screen_shot, "screenshot", false, "opens flameshot gui with the XDG_USER_DIR/Screenshot as the save path")

	rootCMD.Flags().BoolVar(&batt_mon, "battery-monitor", false, "continously checks battery level and notifies the user when its lower")

	rootCMD.Flags().BoolVar(&monitor_daemon, "monitor-daemon", false, "continuosly checks for new/removed monitors and changes the brightness based on the focused monitor.\n NOTE: does not support OLED or LED displays. Only supports LCD displays (displays with backlight)")

	rootCMD.Flags().BoolVarP(&version, "version", "v", false, "prints the package version")
}

func Execute() {
	initializeFlags()
	if err := rootCMD.Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
