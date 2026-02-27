package cmd

import (
	"github.com/spf13/cobra"
	"github.com/wizarki972/myone/internal/modules/walls"
)

var list_walls, wall_menu bool
var install_wall_pack, remove_wall_pack string

var wallpaperCMD = &cobra.Command{
	Use:   "wallpapers",
	Short: "manages wallpapers packs",
	RunE: func(cmd *cobra.Command, args []string) error {

		wall := walls.NewWall()

		if list_walls {
			wall.List()
		}

		if wall_menu {
			wall.ShowWallpaperChangeMenu()
		}

		if len(install_wall_pack) > 0 {
			wall.Install(install_wall_pack)
		}

		if len(remove_wall_pack) > 0 {
			wall.Remove(remove_wall_pack)
		}

		return nil
	},
}

func initializeWallsFlags() {
	wallpaperCMD.Flags().BoolVarP(&list_walls, "list", "l", false, "lists every repo, and installed/update status.")

	wallpaperCMD.Flags().StringVarP(&install_wall_pack, "install", "i", "", "Installs mentioned packs from the repo.")

	wallpaperCMD.Flags().StringVarP(&remove_wall_pack, "remove", "r", "", "Removes the mentioned wall pack from the system.")

	wallpaperCMD.Flags().BoolVarP(&wall_menu, "menu", "m", false, "Shows a rofi menu for the wallpaper choosing.")

	// SET WALLPAPER

	// UPDATE PACKS

	// INSTALL ALL

	// REMOVE ALL

	rootCMD.AddCommand(wallpaperCMD)

}
