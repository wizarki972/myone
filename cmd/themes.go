package cmd

import (
	"github.com/spf13/cobra"
	"github.com/wizarki972/myone/internal/modules/themer"
)

var update_themes, force_update_themes bool
var apply_theme string

var themesCMD = &cobra.Command{
	Use:   "themes",
	Short: "manage themes here...",
	RunE: func(cmd *cobra.Command, args []string) error {
		var t *themer.Themer

		if len(apply_theme) > 0 {
			t = themer.NewThemer(apply_theme)
			t.Apply_Theme()
		} else {
			t = themer.NewThemer("")
		}

		if update_themes && !force_update_themes {
			t.Update()
		}

		if force_update_themes {
			t.Download()
			t.Install()
		}

		return nil
	},
}

func initializeThemesFlags() {
	themesCMD.Flags().BoolVarP(&update_themes, "update", "u", false, "updates all the local themes.")

	themesCMD.Flags().BoolVarP(&force_update_themes, "force-update", "f", false, "Re-downloads all themes and installs it.")

	themesCMD.Flags().StringVarP(&apply_theme, "apply", "a", "", "Applies the specified theme.")

	rootCMD.AddCommand(themesCMD)
}
