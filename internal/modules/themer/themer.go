package themer

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/wizarki972/myone/internal/utils/config"
	themes_config "github.com/wizarki972/myone/internal/utils/config/themes"
	"github.com/wizarki972/myone/internal/utils/fldir"
	"github.com/wizarki972/myone/internal/utils/user"
)

// Maybe make choosing and installing themes - in future

const THEMES_ZIP_URL = "https://raw.githubusercontent.com/wizarki972/mythemes/main/zips/themes.zip"
const VERSION_URL = "https://raw.githubusercontent.com/wizarki972/mythemes/main/VERSION"

func NewThemer(theme_name string) *Themer {
	base := config.GetDirPathFor("base")

	var t = &Themer{
		homeDir:  user.GetHomeDir(),
		themeDir: filepath.Join(base, "themes"),
	}

	// if a theme name is given
	if len(theme_name) > 0 {
		if theme_name == "default" {
			t.ThemeName = config.DEFAULT_THEME
		} else {
			t.ThemeName = theme_name
		}
		return t
	}

	current_theme_path := filepath.Join(base, themes_config.CURRENT_THEME_NAME_ENTRY)
	if fldir.IsPathExist(current_theme_path) {

		current, err := fldir.ReadFileAsString(current_theme_path)
		if err != nil {
			// Change the error to something like `cannot get current theme`
			panic(err)
		}
		t.ThemeName = strings.TrimSpace(current)

		// second check for errors
		if len(t.ThemeName) == 0 {
			panic(errors.New("cannot get currently applied theme"))
		}

		return t

	} else {
		t.ThemeName = config.DEFAULT_THEME
		return t
	}

}

type Themer struct {
	ThemeName string
	homeDir   string
	themeDir  string
}

func (t *Themer) Update() {
	var local_v, repo_v float64
	var local_sv, repo_sv int

	// local version
	var version_path = filepath.Join(t.themeDir, "VERSION")
	verStr, err := fldir.ReadFileAsString(version_path)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No themes found. Downloading themes...")
			t.Download()
			return
		}
		panic(err)
	}
	local_v, local_sv = version_parser(verStr)

	// repo version
	verStr = fldir.ReadTextFileFromURL(VERSION_URL, false, "")
	repo_v, repo_sv = version_parser(verStr)

	if (local_v < repo_v) || (local_v == repo_v && local_sv < repo_sv) {
		t.Download()
		return
	} else {
		fmt.Println("Themes are already up-to-date.")
	}

}

func (t *Themer) Download() {
	// CACHE PATH CHECK
	cache_dir := filepath.Join(config.CACHE_BASE_DIR, "themes")
	cache_path := filepath.Join(cache_dir, "themes.zip")
	fldir.CreateDirectory(cache_dir)

	// DOWNLOADING ZIP
	fldir.DownloadURL(THEMES_ZIP_URL, cache_path, true)

	// REMOVING CURRENTLY DOWNLOADED VERSION
	if err := os.RemoveAll(t.themeDir); err != nil {
		slog.Error("Failed to remove themes ==> " + err.Error())
	}
	fldir.CreateDirectory(t.themeDir)

	// Moving downloaded files to all themes directory
	fldir.Unzip(cache_path, t.themeDir)

	slog.Info("Cleaning Up...")
	if err := os.RemoveAll(cache_dir); err != nil {
		slog.Error("Failed to remove cache")
	}
}

func (t *Themer) Install() {
	themepath := filepath.Join(t.themeDir, t.ThemeName)
	if !fldir.IsPathExist(themepath) {
		slog.Info("Theme not found, trying to update themes...")
		t.Download()
	}

	t.copy_files(themepath, "")
	fldir.WriteStringToFile(t.ThemeName, filepath.Join(config.GetDirPathFor("base"), themes_config.CURRENT_THEME_NAME_ENTRY))
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
	fldir.WriteStringToFile(file, save_path)
}

func version_parser(version string) (float64, int) {
	version_parts := strings.Split(version, "-")
	version_fl, err := strconv.ParseFloat(strings.SplitN(version_parts[0], ".", 2)[1], 64)
	if err != nil {
		panic(err)
	}
	sub_version, err := strconv.Atoi(strings.TrimSpace(version_parts[1]))
	if err != nil {
		panic(err)
	}
	return version_fl, sub_version
}
