package cmd

import (
	"github.com/spf13/cobra"
	"github.com/wizarki972/myone/internal/modules/themer"
)

var updateThemes, forceUpdateThemes bool
var applyTheme string

var themesCMD = &cobra.Command{
	Use:   "themes",
	Short: "manage themes here...",
	RunE: func(cmd *cobra.Command, args []string) error {
		loggerInstance = handleLogg()
		loggerInstance.AddSubCommand("myone themes")

		var t *themer.Themer
		if len(applyTheme) > 0 {
			loggerInstance.AddFlag("apply")
			t = themer.NewThemer(applyTheme, loggerInstance)
			t.Apply_Theme()
		} else {
			t = themer.NewThemer("", loggerInstance)
		}

		if updateThemes && !forceUpdateThemes {
			loggerInstance.AddFlag("update")
			t.Update()
		}

		if forceUpdateThemes {
			loggerInstance.AddFlag("force-update")
			t.Download()
			t.Install()
		}

		if saveLog || len(logPath) > 0 {
			loggerInstance.SaveBook()
		}

		return nil
	},
}

func initializeThemesFlags() {
	themesCMD.Flags().BoolVarP(&updateThemes, "update", "u", false, "updates all the local themes.")

	themesCMD.Flags().BoolVarP(&forceUpdateThemes, "force-update", "f", false, "Re-downloads all themes and installs it.")

	themesCMD.Flags().StringVarP(&applyTheme, "apply", "a", "", "Applies the specified theme.")

	themesCMD.Flags().BoolVar(&saveLog, "save-log", false, "saves the based on the default path or path specified in config.\nNo need to use this flag, if you are using --log-path flag.")

	themesCMD.Flags().StringVar(&logPath, "log-path", "", "Enter the path to save the log.")

	rootCMD.AddCommand(themesCMD)
}
