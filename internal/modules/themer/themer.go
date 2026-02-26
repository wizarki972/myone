package themer

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/wizarki972/myone/internal/utils/config"
	themes_config "github.com/wizarki972/myone/internal/utils/config/themes"
	"github.com/wizarki972/myone/internal/utils/fldir"
	"github.com/wizarki972/myone/internal/utils/user"
)

const THEMES_ZIP_URL = "https://raw.githubusercontent.com/wizarki972/mythemes/main/zips/themes.zip"
const VERSION_URL = "https://raw.githubusercontent.com/wizarki972/mythemes/main/VERSION"

func NewThemer(theme_name string) *Themer {
	if theme_name == "default" {
		return &Themer{
			ThemeName: config.DEFAULT_THEME,
			homeDir:   user.GetHomeDir(),
		}
	}

	return &Themer{
		ThemeName: theme_name,
		homeDir:   user.GetHomeDir(),
	}
}

type Themer struct {
	ThemeName string
	homeDir   string
}

func Download() {
	// MAIN AREA CHECK
	themes_path := config.GetDirPathFor("themes")
	fldir.CreateDirectory(themes_path)

	// CACHE PATH CHECK
	cache_dir := filepath.Join(config.GetDirPathFor("cache"), "themes")
	cache_path := filepath.Join(config.GetDirPathFor("cache"), "themes/themes.zip")
	fldir.CreateDirectory(cache_dir)

	// DOWNLOADING ZIP
	fldir.DownloadURL(THEMES_ZIP_URL, cache_path)

	// REMOVING CURRENTLY DOWNLOADED VERSION
	if err := os.RemoveAll(themes_path); err != nil {
		slog.Error("Failed to remove themes ==> " + err.Error())
	}
	fldir.CreateDirectory(themes_path)

	// Moving downloaded files to main area
	fldir.Unzip(cache_path, themes_path)

	slog.Info("Cleaning Up...")
	if err := os.RemoveAll(cache_dir); err != nil {
		slog.Error("Failed to remove cache")
	}
}

func (t *Themer) Install() {
	themepath := filepath.Join(config.GetDirPathFor("themes"), t.ThemeName)
	if !fldir.IsPathExist(themepath) {
		slog.Info("Theme not found, trying to update themes...")
		Download()
	}

	t.copy_files(themepath, "")

}

func (t *Themer) copy_files(path, suffix string) {
	entries, err := os.ReadDir(path)
	if err != nil {
		panic(err)
	}

	for _, entry := range entries {
		entry_path := filepath.Join(path, entry.Name())
		info, err := os.Stat(entry_path)
		if err != nil {
			panic(err)
		}

		if info.IsDir() {
			t.copy_files(entry_path, filepath.Join(suffix, entry.Name()))
		} else {
			fldir.CreateDirectory(filepath.Join(t.homeDir, suffix))
			if strings.HasPrefix(entry.Name(), "$") {
				fill(entry_path, filepath.Join(t.homeDir, suffix, strings.TrimPrefix(entry.Name(), "$")))
			} else {
				fldir.CopyFile(entry_path, filepath.Join(t.homeDir, suffix, entry.Name()))
			}
		}
	}

}

func fill(current_path, save_path string) {
	file, err := fldir.ReadFileAsString(current_path)
	if err != nil {
		panic(err)
	}

	for old, new := range themes_config.ThemePlaceholderValues {
		file = strings.ReplaceAll(file, old, new)
	}
	fldir.WriteFile(file, save_path)
}
