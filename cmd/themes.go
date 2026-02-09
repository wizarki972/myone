package cmd

import (
	"github.com/spf13/cobra"
	"github.com/wizarki972/myone/internal/modules/themer"
)

var update_themes bool
var apply_theme string

var themesCMD = &cobra.Command{
	Use:   "theme",
	Short: "manage themes here...",
	RunE: func(cmd *cobra.Command, args []string) error {
		if update_themes {
			themer.Download()
		}

		if len(apply_theme) > 0 {
			themer.NewThemer(apply_theme).Install()
		}

		return nil
	},
}

func initializeThemesFlags() {
	themesCMD.Flags().BoolVarP(&update_themes, "update", "u", false, "updates all the local themes")

	themesCMD.Flags().StringVarP(&apply_theme, "apply", "a", "", "Applies the specified theme")

	rootCMD.AddCommand(themesCMD)
}
