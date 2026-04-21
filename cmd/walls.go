package cmd

import (
	"github.com/spf13/cobra"
	"github.com/wizarki972/myone/internal/modules/walls"
)

var list, wallMenu bool
var installWallPack, removeWallPack string

var wallpaperCMD = &cobra.Command{
	Use:   "wallpapers",
	Short: "manages wallpaper packs",
	RunE: func(cmd *cobra.Command, args []string) error {
		loggerInstance = handleLogg()
		loggerInstance.AddSubCommand("myone wallpapers")

		wall := walls.NewWall(loggerInstance)
		if list {
			loggerInstance.AddFlag("list")
			wall.List()
		}

		if wallMenu {
			loggerInstance.AddFlag("menu")
			wall.ShowWallpaperChangeMenu()
		}

		if len(installWallPack) > 0 {
			loggerInstance.AddFlag("install")
			wall.Install(installWallPack)
		}

		if len(removeWallPack) > 0 {
			loggerInstance.AddFlag("remove")
			wall.Remove(removeWallPack)
		}

		if saveLog || len(logPath) > 0 {
			loggerInstance.SaveBook()
		}

		return nil
	},
}

func initializeWallsFlags() {
	wallpaperCMD.Flags().BoolVarP(&list, "list", "l", false, "lists every repo, and installed/update status.")

	wallpaperCMD.Flags().StringVarP(&installWallPack, "install", "i", "", "Installs mentioned packs from the repo.")

	wallpaperCMD.Flags().StringVarP(&removeWallPack, "remove", "r", "", "Removes the mentioned wall pack from the system.")

	wallpaperCMD.Flags().BoolVarP(&wallMenu, "menu", "m", false, "Shows a rofi menu for the wallpaper choosing.")

	wallpaperCMD.Flags().BoolVar(&saveLog, "save-log", false, "saves the based on the default path or path specified in config.\nNo need to use this flag, if you are using --log-path flag.")

	wallpaperCMD.Flags().StringVar(&logPath, "log-path", "", "Enter the path to save the log.")

	// SET WALLPAPER

	// UPDATE PACKS

	// INSTALL ALL

	// REMOVE ALL

	rootCMD.AddCommand(wallpaperCMD)

}
