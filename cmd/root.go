package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wizarki972/myone/internal/common"
	"github.com/wizarki972/myone/internal/modules/audio"
	"github.com/wizarki972/myone/internal/modules/battery"
	"github.com/wizarki972/myone/internal/modules/bootstrap"
	"github.com/wizarki972/myone/internal/modules/display"
	"github.com/wizarki972/myone/internal/modules/logout"
	"github.com/wizarki972/myone/internal/modules/screenshot"
	"github.com/wizarki972/myone/internal/utils/logger"
	"github.com/wizarki972/myone/internal/utils/pkg"
)

var brightness, volumeNotify, logPath string
var logOut int
var screenShot, monitorDaemon, batteryMonitor, version, update, depCheck, saveLog bool
var loggerInstance *logger.LogBook

var rootCMD = &cobra.Command{
	Use:   "myone",
	Short: "my one utility for my needs",
	RunE: func(cmd *cobra.Command, args []string) error {
		loggerInstance = handleLogg()
		loggerInstance.AddSubCommand("myone")

		if len(brightness) > 0 {
			display.ChangeBrightness(brightness)
		}

		if logOut > 0 {
			loggerInstance.AddFlag("logout")
			logout.Logout(min(logOut, 2), loggerInstance)
		}

		if len(volumeNotify) > 0 {
			audio.NotifyVolume(volumeNotify)
		}

		if screenShot {
			screenshot.OpenGUI()
		}

		if batteryMonitor {
			loggerInstance.AddFlag("battery-monitor")
			battery.NewBatteryMonitor(loggerInstance).StartMonitor()
		}

		if monitorDaemon {
			loggerInstance.AddFlag("monitor-daemon")
			display.NewMonitorDaemon(loggerInstance).StartDaemon()
		}

		if version {
			fmt.Println(common.VERSION)
		}

		if update {
			loggerInstance.AddFlag("update")
			bootstrap.SelfUpdate(loggerInstance)
		}

		if depCheck {
			pkg.Dependency_check()
		}

		if saveLog || len(logPath) > 0 {
			loggerInstance.SaveBook()
		}

		return nil
	},
}

func initializeFlags() {
	rootCMD.Flags().StringVar(&brightness, "bright", "", "+5% - increases the brightness by 5%, \n-5% decreases the brightness by 5%.")

	rootCMD.Flags().IntVar(&logOut, "logout", 0, "accepted values 1, 2. Displays power menu.")

	rootCMD.Flags().StringVar(&volumeNotify, "volume-osd", "", "just tells swayosd to show current volume level of the current sink(speaker or output device)/source(microphone or input device).\nAccepted values: sink, source.")

	rootCMD.Flags().BoolVar(&screenShot, "screenshot", false, "opens flameshot gui with the XDG_USER_DIR/Screenshot as the save path.")

	rootCMD.Flags().BoolVar(&batteryMonitor, "battery-monitor", false, "continously checks battery level and notifies the user when its lower.")

	rootCMD.Flags().BoolVar(&monitorDaemon, "monitor-daemon", false, "continuosly checks for new/removed monitors and changes the brightness based on the focused monitor.\n NOTE: does not support OLED or LED displays. Only supports LCD displays (displays with backlight).")

	rootCMD.Flags().BoolVarP(&version, "version", "v", false, "prints the package version.")

	rootCMD.Flags().BoolVarP(&update, "update", "u", false, "for updating the package.")

	rootCMD.Flags().BoolVar(&depCheck, "dependency-check", false, "checks whether all dependencies are installed.")

	rootCMD.Flags().BoolVar(&saveLog, "save-log", false, "saves the based on the default path or path specified in config.\nNo need to use this flag, if you are using --log-path flag.")

	rootCMD.Flags().StringVar(&logPath, "log-path", "", "Enter the path to save the log.")

	initializeThemesFlags()
	initializeWallsFlags()
}

func Execute() {
	initializeFlags()
	if err := rootCMD.Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
