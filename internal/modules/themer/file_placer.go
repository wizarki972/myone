package themer

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/wizarki972/myone/internal/common"
	"github.com/wizarki972/myone/internal/modules/display"
	"github.com/wizarki972/myone/internal/utils/fldir"
)

// generates dynamic/placeholder values for dynamic config files
func (t *Themer) generatePlaceholderValues() {
	t.themePlaceholderValues = map[string]string{
		"${SCRIPTS_DIRECTORY_PATH}":   filepath.Join(t.homeDir, common.SCRIPTS_DIR),
		"${CURRENT_WALLPAPER_PATH}":   filepath.Join(t.homeDir, common.CURRENT_WALLPAPER_ENTRY_PATH),
		"${ALL_WALLS_DIRECTORY_PATH}": filepath.Join(t.homeDir, common.ALL_WALLS_DIR),
		"${ROFI_IMAGE}":               t.get_rofi_image(),
		"${SCREEN_WIDTH}":             strconv.Itoa(display.GetScreenResolution()[0]),
		"${SCREEN_HEIGHT}":            strconv.Itoa(display.GetScreenResolution()[1]),
	}
}

func (t *Themer) placeThemeDependentFiles() {
	td_path := filepath.Join(t.themesDir, "theme_deps")

	// checks
	info, err := os.Stat(td_path)
	if err != nil {
		if os.IsNotExist(err) {
			slog.Info("No theme dependent configs are found.")
			return
		}
		panic(err)
	}
	if !info.IsDir() {
		slog.Info("Instead of theme dependent config files directory, found a file. So Skipping...")
		return
	}

	// place files logic
	if err := t.placeFilesLogic(td_path, "", true); err != nil {
		slog.Warn(err.Error())
	}

}

func (t *Themer) placeCommonFiles() {
	common_dir := filepath.Join(t.themesDir, "common")
	if !fldir.IsPathExist(common_dir) {
		slog.Info("Theme not found, trying to update themes...")
		t.Download()
	}

	if err := t.placeFilesLogic(common_dir, "", false); err != nil {
		panic(err)
	}

	t.set_common_state(true)
}

func (t *Themer) placeFilesLogic(path, suffix string, force_fill bool) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		panic(err)
	}

	if len(entries) == 0 {
		return fmt.Errorf("no files found in this directory (%s)", path)
	}

	for _, entry := range entries {
		entry_path := filepath.Join(path, entry.Name())

		if entry.IsDir() {
			t.placeFilesLogic(entry_path, filepath.Join(suffix, entry.Name()), force_fill)
		} else {
			fldir.CreateDirectory(filepath.Join(t.homeDir, suffix))
			if force_fill || strings.HasPrefix(entry.Name(), "$") {
				t.fill(entry_path, filepath.Join(t.homeDir, suffix, strings.TrimPrefix(entry.Name(), "$")))
			} else {
				fldir.CopyFile(entry_path, filepath.Join(t.homeDir, suffix, entry.Name()))
			}
		}
	}

	return nil
}

func (t *Themer) fill(current_path, save_path string) {
	file, err := fldir.ReadFileAsString(current_path)
	if err != nil {
		panic(err)
	}

	for old, new := range t.themePlaceholderValues {
		file = strings.ReplaceAll(file, old, new)
	}
	fldir.WriteStringToFile(file, save_path)
}
