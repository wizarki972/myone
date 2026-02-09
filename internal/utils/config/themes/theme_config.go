package themes_config

import (
	"path/filepath"
	"strconv"

	"github.com/wizarki972/myone/internal/modules/display"
	"github.com/wizarki972/myone/internal/utils/config"
)

const CURRENT_WALL_NAME = ".wallpaper"

var ThemePlaceholderValues = map[string]string{
	"${WALLPAPER_PATH}":     filepath.Join(config.BASE_DIR, CURRENT_WALL_NAME),
	"${ALL_WALLS_DIR_PATH}": ".local/share/mywalls",
	"${SCRIPTS_DIR_PATH}":   config.GetDirPathFor("scripts"),
	"${SCREEN_WIDTH}":       strconv.Itoa(display.GetScreenresolution()[0]),
	"${SCREEN_HEIGHT}":      strconv.Itoa(display.GetScreenresolution()[1]),
}
