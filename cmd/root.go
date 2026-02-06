package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/wizarki972/myone/internal/modules/display"
)

var brightness string

var rootCMD = &cobra.Command{
	Use:   "myone",
	Short: "my one utility for my needs",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(brightness) > 0 {
			display.ChangeBrightness(brightness)
		}

		return nil
	},
}

func initializeFlags() {
	rootCMD.Flags().StringVar(&brightness, "b", "", "+5% - increases the brightness by 5%, \n-5% decreases the brightness by 5%")
}

func Execute() {
	initializeFlags()
	if err := rootCMD.Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
