package themer

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/wizarki972/myone/internal/common"
	"github.com/wizarki972/myone/internal/modules/display"
	"github.com/wizarki972/myone/internal/utils/fldir"
	"github.com/wizarki972/myone/internal/utils/logger"
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

// places files that changes based on themes
func (t *Themer) placeThemeDependentFiles() {
	td_path := filepath.Join(t.themesDir, "theme_deps")

	// checks
	info, err := os.Stat(td_path)
	if err != nil {
		if os.IsNotExist(err) {
			t.logg_book.EnterLogAndPrint("No theme dependent configs are found.", logger.LogTypes.Error, err)
			return
		}
		panic(err)
	}
	if !info.IsDir() {
		t.logg_book.EnterLogAndPrint("Instead of theme dependent config files directory, found a file. Skipping it.", logger.LogTypes.Warning, nil)
		return
	}

	// place files logic
	if err := t.placeFilesLogic(td_path, "", true); err != nil {
		t.logg_book.EnterLogAndPrint("An error occurred while placing theme dependent files. error => "+err.Error(), logger.LogTypes.Warning, err)
	}

}

// place files
func (t *Themer) placeCommonFiles() {
	common_dir := filepath.Join(t.themesDir, "common")
	if !fldir.IsPathExist(common_dir) {
		t.logg_book.EnterLogAndPrint("Theme not found, trying to download/update themes...", logger.LogTypes.Info, nil)
		t.Download()
	}

	if err := t.placeFilesLogic(common_dir, "", false); err != nil {
		t.logg_book.EnterLogAndPrint("An error occurred while placing common files. error => "+err.Error(), logger.LogTypes.Warning, err)
	}

	t.set_common_state(true)
}

// common logic for placing files
func (t *Themer) placeFilesLogic(path, suffix string, force_fill bool) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		t.logg_book.EnterLogAndPrint("Cannot read entries from directory - "+path, logger.LogTypes.Error, err)
	}

	if len(entries) == 0 {
		return fmt.Errorf("no files found in this directory (%s)", path)
	}

	for _, entry := range entries {
		entry_path := filepath.Join(path, entry.Name())

		if entry.IsDir() {
			t.placeFilesLogic(entry_path, filepath.Join(suffix, entry.Name()), force_fill)
		} else {
			if err := fldir.CreateDirectory(filepath.Join(t.homeDir, suffix)); err != nil {
				t.logg_book.EnterLogAndPrint(err.Error(), logger.LogTypes.Error, err)
			}
			if force_fill || strings.HasPrefix(entry.Name(), "$") {
				t.fill(entry_path, filepath.Join(t.homeDir, suffix, strings.TrimPrefix(entry.Name(), "$")))
			} else {
				if err := fldir.CopyFile(entry_path, filepath.Join(t.homeDir, suffix, entry.Name())); err != nil {
					t.logg_book.EnterLogAndPrint(err.Error(), logger.LogTypes.Error, err)
				}
			}
		}
	}

	return nil
}

// fills the data required by files.
func (t *Themer) fill(current_path, save_path string) {
	file, err := fldir.ReadFileAsString(current_path)
	if err != nil {
		t.logg_book.EnterLogAndPrint("Cannot read file - "+current_path, logger.LogTypes.Error, err)
	}

	for old, new := range t.themePlaceholderValues {
		file = strings.ReplaceAll(file, old, new)
	}
	if err := fldir.WriteStringToFile(file, save_path); err != nil {
		t.logg_book.EnterLogAndPrint(err.Error(), logger.LogTypes.Error, err)
	}
}
