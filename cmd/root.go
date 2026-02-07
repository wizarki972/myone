package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/wizarki972/myone/internal/modules/audio"
	"github.com/wizarki972/myone/internal/modules/display"
	"github.com/wizarki972/myone/internal/modules/screenshot"
)

var brightness string
var vol_notify, screen_shot bool

var rootCMD = &cobra.Command{
	Use:   "myone",
	Short: "my one utility for my needs",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(brightness) > 0 {
			display.ChangeBrightness(brightness)
		}

		if vol_notify {
			audio.NotifyVolume()
		}

		if screen_shot {
			screenshot.OpenGUI()
		}

		return nil
	},
}

func initializeFlags() {
	rootCMD.Flags().StringVar(&brightness, "b", "", "+5% - increases the brightness by 5%, \n-5% decreases the brightness by 5%")

	rootCMD.Flags().BoolVar(&vol_notify, "volume-osd", false, "just tells swayosd to show current volume level of the current sink.")

	rootCMD.Flags().BoolVar(&screen_shot, "screenshot", false, "opens flameshot gui with the XDG_USER_DIR/Screenshot as the save path")
}

func Execute() {
	initializeFlags()
	if err := rootCMD.Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
