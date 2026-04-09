package themer

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/wizarki972/myone/internal/common"
	"github.com/wizarki972/myone/internal/modules/bootstrap"
	"github.com/wizarki972/myone/internal/modules/display"
	"github.com/wizarki972/myone/internal/utils/fldir"
)

// Maybe make choosing and installing themes - in future
const DEFAULT_THEME = "tokyonight"
const THEMES_ZIP_URL = "https://raw.githubusercontent.com/wizarki972/mythemes/main/zips/themes.zip"
const VERSION_URL = "https://raw.githubusercontent.com/wizarki972/mythemes/main/VERSION"

// Themer struct generator
func NewThemer(theme_name string) *Themer {
	home := fldir.GetHomeDir()

	var t = &Themer{
		homeDir:   home,
		themesDir: filepath.Join(home, common.THEMES_DIR),
		cacheDir:  filepath.Join(home, common.CACHE_DIR, "themes"),

		commonStatePath:      filepath.Join(home, common.COMMON_PLACED_STATE_PATH),
		currentThemeNamePath: filepath.Join(home, common.CURRENT_THEME_NAME_ENTRY_PATH),
	}

	// if a theme name is given
	if len(theme_name) > 0 {
		if theme_name == "default" {
			t.ThemeName = DEFAULT_THEME
		} else {
			t.ThemeName = theme_name
		}
		return t
	}

	// current theme
	if fldir.IsPathExist(t.currentThemeNamePath) {

		var err error
		t.ThemeName, err = fldir.ReadFileAsString(t.currentThemeNamePath)
		if err != nil {
			// Change the error to something like `cannot get current theme`
			panic(err)
		}

		// second check for errors
		if len(t.ThemeName) == 0 {
			panic(errors.New("cannot get currently applied theme"))
		}

		return t

	} else {
		t.ThemeName = DEFAULT_THEME
		return t
	}
}

type Themer struct {
	ThemeName string

	homeDir   string
	themesDir string
	cacheDir  string

	commonStatePath      string
	currentThemeNamePath string

	themePlaceholderValues map[string]string
}

// Updates the config/theme files if a new version is available
func (t *Themer) Update() {
	var local_v, repo_v float64
	var local_sv, repo_sv int

	// local version
	verStr, err := fldir.ReadFileAsString(filepath.Join(t.themesDir, "VERSION"))
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

	// update starts
	if (local_v < repo_v) || (local_v == repo_v && local_sv < repo_sv) {
		t.Download()
		t.Install()
		return
	} else {
		fmt.Println("Themes are already up-to-date.")
	}
}

// Downloads the files from the repo
func (t *Themer) Download() {
	// CACHE PATH CHECK
	cache_path := filepath.Join(t.cacheDir, "themes.zip")
	fldir.CreateDirectory(t.cacheDir)

	// DOWNLOADING ZIP
	fldir.DownloadURL(THEMES_ZIP_URL, cache_path, true)

	// REMOVING CURRENTLY INSTALLED VERSION
	if err := os.RemoveAll(t.themesDir); err != nil {
		slog.Error("Failed to remove themes ==> " + err.Error())
	}
	fldir.CreateDirectory(t.themesDir)

	// Extracting downloaded file
	fldir.Unzip(cache_path, t.themesDir)

	slog.Info("Cleaning Up...")
	if err := os.RemoveAll(t.cacheDir); err != nil {
		slog.Error("Failed to remove cache")
	}
}

// Installs the downloaded config/theme files
func (t *Themer) Install() {
	t.generate_placeholder_values()
	// Dir checks
	if !fldir.IsPathExist(t.themesDir) {
		slog.Info("Theme not found, trying to update themes...")
		t.Download()
	}

	entries, err := os.ReadDir(t.themesDir)
	if err != nil {
		panic(err)
	}
	if len(entries) == 0 {
		slog.Info("Nothing found in the themes directory. Trying to populate it...")
		t.Download()
	}

	// placing files
	t.place_common_files()

	// theme dependent file
	t.place_theme_dependent_files()

	// applying colors
	t.apply_colors()

	// dependency check
	dep_lst_path := filepath.Join(t.themesDir, "deps.lst")
	if fldir.IsPathExist(dep_lst_path) {
		fmt.Println("Dependency check...")
		bootstrap.Install_pkgs_from_file(dep_lst_path)
	}

	// writing current theme
	fldir.WriteStringToFile(t.ThemeName, t.currentThemeNamePath)
}

// applies themes
func (t *Themer) Apply_Theme() {
	t.generate_placeholder_values()
	if !t.common_state() {
		t.place_common_files()
	}
	t.place_theme_dependent_files()
	t.apply_colors()
	fldir.WriteStringToFile(t.ThemeName, t.currentThemeNamePath)
}

// generates dynamic/placeholder values for dynamic config files
func (t *Themer) generate_placeholder_values() {
	t.themePlaceholderValues = map[string]string{
		"${SCRIPTS_DIRECTORY_PATH}":   filepath.Join(t.homeDir, common.SCRIPTS_DIR),
		"${CURRENT_WALLPAPER_PATH}":   filepath.Join(t.homeDir, common.CURRENT_WALLPAPER_ENTRY_PATH),
		"${ALL_WALLS_DIRECTORY_PATH}": filepath.Join(t.homeDir, common.ALL_WALLS_DIR),
		"${ROFI_IMAGE}":               t.get_rofi_image(),
		"${SCREEN_WIDTH}":             strconv.Itoa(display.GetScreenResolution()[0]),
		"${SCREEN_HEIGHT}":            strconv.Itoa(display.GetScreenResolution()[1]),
	}
}

// changes colors files based on the theme
func (t *Themer) apply_colors() {
	colors_dir := filepath.Join(t.themesDir, "colors", t.ThemeName)

	// Checks
	info, err := os.Stat(colors_dir)
	if err != nil {
		panic(err)
	}
	if !info.IsDir() {
		panic(errors.New("colors directory not found for this theme"))
	}

	// loading schema
	schema_path := filepath.Join(t.themesDir, "colors", "schema")
	content, err := fldir.ReadFileAsString(schema_path)
	if err != nil {
		panic(err)
	}
	var schema map[string]string = make(map[string]string)
	for line := range strings.SplitSeq(content, "\n") {
		parts := strings.Split(line, "=")
		schema[parts[0]] = filepath.Join(t.homeDir, parts[1])
	}

	// logic
	entries, err := os.ReadDir(colors_dir)
	if err != nil {
		panic(err)
	}
	for _, entry := range entries {
		// entry check
		if entry.IsDir() {
			slog.Warn(fmt.Sprintf("Folder %s is skipped", filepath.Join(t.themesDir, entry.Name())))
		}

		target_path, ok := schema[entry.Name()]
		if !ok {
			slog.Warn("unknown colors file found. Skipping")
			continue
		}

		// applying
		fldir.CopyFile(filepath.Join(colors_dir, entry.Name()), target_path)
	}
}

// stores whether config files are installed or not
func (t *Themer) common_state() bool {
	if fldir.IsPathExist(t.commonStatePath) {
		data, err := fldir.ReadFileAsString(t.commonStatePath)
		if err != nil {
			panic(err)
		}
		if data == "1" {
			return true
		}
	}
	return false
}

func (t *Themer) set_common_state(state bool) {
	var content string
	if state {
		content = "1"
	} else {
		content = "0"
	}
	fldir.WriteStringToFile(content, t.commonStatePath)
}

// getting rofi launcher image path based on the theme
func (t *Themer) get_rofi_image() string {
	rofi_img_dir := filepath.Join(t.themesDir, "assets", "images", "rofi")
	entries, err := os.ReadDir(rofi_img_dir)
	if err != nil {
		panic(err)
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), t.ThemeName) {
			return filepath.Join(rofi_img_dir, entry.Name())
		}
	}

	panic(errors.New("rofi theme image pair not found"))
}

// version string parser
func version_parser(version string) (float64, int) {
	version_parts := strings.Split(version, "-")
	version_fl, err := strconv.ParseFloat(strings.SplitN(version_parts[0], ".", 2)[1], 64)
	if err != nil {
		panic(err)
	}
	sub_version, err := strconv.Atoi(version_parts[1])
	if err != nil {
		panic(err)
	}
	return version_fl, sub_version
}
