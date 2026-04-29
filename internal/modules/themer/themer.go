package themer

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/wizarki972/myone/internal/common"
	"github.com/wizarki972/myone/internal/utils/fldir"
	"github.com/wizarki972/myone/internal/utils/logger"
	"github.com/wizarki972/myone/internal/utils/pkg"
)

// Maybe make choosing and installing themes - in future
const DEFAULT_THEME = "tokyonight"
const THEMES_ZIP_URL = "https://raw.githubusercontent.com/wizarki972/mythemes/main/zips/themes.zip"
const VERSION_URL = "https://raw.githubusercontent.com/wizarki972/mythemes/main/VERSION"

// Themer struct generator
func NewThemer(theme_name string, logg_book *logger.LogBook) *Themer {
	home := fldir.GetHomeDir()

	var t = &Themer{
		logg_book: logg_book,
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
			t.logg_book.EnterLogAndPrint("Cannot get current theme name from the path - "+t.currentThemeNamePath, logger.LogTypes.Error, err)
		}

		// second check for errors
		if len(t.ThemeName) == 0 {
			t.logg_book.EnterLogAndPrint("Theme name from the path - "+t.currentThemeNamePath+" is empty.", logger.LogTypes.Error, err)
		}

		return t

	} else {
		t.ThemeName = DEFAULT_THEME
		return t
	}
}

type Themer struct {
	logg_book *logger.LogBook
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
			t.logg_book.EnterLogAndPrint("No themes found. Downloading themes...", logger.LogTypes.Info, nil)
			t.Download()
			return
		}
		t.logg_book.EnterLogAndPrint("Error while reading installed themes version from path - "+t.themesDir+"/VERSION", logger.LogTypes.Error, err)
	}
	local_v, local_sv = t.version_parser(verStr)

	// repo version
	verStr, err = fldir.ReadTextFileFromURL(VERSION_URL, false, "")
	if err != nil {
		t.logg_book.EnterLogAndPrint(err.Error(), logger.LogTypes.Error, err)
	}
	repo_v, repo_sv = t.version_parser(verStr)

	// update starts
	if (local_v < repo_v) || (local_v == repo_v && local_sv < repo_sv) {
		t.Download()
		t.Install()
		return
	} else {
		t.logg_book.EnterLogAndPrint("Themes are already up-to-date.", logger.LogTypes.Info, nil)
	}
}

// Downloads the files from the repo
func (t *Themer) Download() {
	// CACHE PATH CHECK
	cache_path := filepath.Join(t.cacheDir, "themes.zip")
	if err := fldir.CreateDirectory(t.cacheDir); err != nil {
		t.logg_book.EnterLogAndPrint(err.Error(), logger.LogTypes.Error, err)
	}

	// DOWNLOADING ZIP
	if err := fldir.DownloadURL(THEMES_ZIP_URL, cache_path, true); err != nil {
		t.logg_book.EnterLogAndPrint(err.Error(), logger.LogTypes.Error, err)
	}

	// REMOVING CURRENTLY INSTALLED VERSION
	if err := os.RemoveAll(t.themesDir); err != nil {
		t.logg_book.EnterLogAndPrint("Failed to remove old themes.", logger.LogTypes.Error, err)
	}
	if err := fldir.CreateDirectory(t.themesDir); err != nil {
		t.logg_book.EnterLogAndPrint(err.Error(), logger.LogTypes.Error, err)
	}

	// Extracting downloaded file
	fldir.Unzip(cache_path, t.themesDir)

	t.logg_book.EnterLogAndPrint("Clearing cache...", logger.LogTypes.Info, nil)
	if err := os.RemoveAll(t.cacheDir); err != nil {
		t.logg_book.EnterLogAndPrint("Failed to clear cache.", logger.LogTypes.Error, nil)
	}
}

// Installs the downloaded config/theme files
func (t *Themer) Install() {
	t.generatePlaceholderValues()
	// Dir checks
	if !fldir.IsPathExist(t.themesDir) {
		t.logg_book.EnterLogAndPrint("Theme not found, trying to update themes...", logger.LogTypes.Info, nil)
		t.Download()
	}

	entries, err := os.ReadDir(t.themesDir)
	if err != nil {
		t.logg_book.EnterLogAndPrint("Cannot get theme entries from - "+t.themesDir, logger.LogTypes.Error, err)
	}
	if len(entries) == 0 {
		t.logg_book.EnterLogAndPrint("Nothing found in the themes directory. Trying to populate it...", logger.LogTypes.Info, nil)
		t.Download()
	}

	// placing files
	t.placeCommonFiles()

	// theme dependent file
	t.placeThemeDependentFiles()

	// applying colors
	t.apply_colors()

	// dependency check
	dep_lst_path := filepath.Join(t.themesDir, "deps.lst")
	if fldir.IsPathExist(dep_lst_path) {
		t.logg_book.EnterLogAndPrint("Performing dependency check for the themes...", logger.LogTypes.Info, nil)
		if err := pkg.InstallPkgsFromFile(dep_lst_path); err != nil {
			t.logg_book.EnterLogAndPrint(err.Error(), logger.LogTypes.Error, err)
		}
	}

	// writing current theme
	if err := fldir.WriteStringToFile(t.ThemeName, t.currentThemeNamePath); err != nil {
		t.logg_book.EnterLogAndPrint(err.Error(), logger.LogTypes.Error, err)
	}
}

// applies themes
func (t *Themer) Apply_Theme() {
	t.generatePlaceholderValues()
	if !t.common_state() {
		t.placeCommonFiles()
	}
	t.placeThemeDependentFiles()
	t.apply_colors()
	if err := fldir.WriteStringToFile(t.ThemeName, t.currentThemeNamePath); err != nil {
		t.logg_book.EnterLogAndPrint(err.Error(), logger.LogTypes.Error, err)
	}
	t.refreshDesktop()
}

// changes colors files based on the theme
func (t *Themer) apply_colors() {
	colors_dir := filepath.Join(t.themesDir, "colors", t.ThemeName)

	// Checks
	info, err := os.Stat(colors_dir)
	if err != nil {
		t.logg_book.EnterLogAndPrint("Cannot access colors directory or not found. ("+colors_dir+").", logger.LogTypes.Error, err)
	}
	if !info.IsDir() {
		t.logg_book.EnterLogAndPrint("Colors directory not found for this theme - "+colors_dir, logger.LogTypes.Error, errors.New("colors directory not found for this theme"))
	}

	// loading schema
	schema_path := filepath.Join(t.themesDir, "colors", "schema")
	content, err := fldir.ReadFileAsString(schema_path)
	if err != nil {
		t.logg_book.EnterLogAndPrint("Error while reading schema - "+schema_path, logger.LogTypes.Error, err)
	}
	var schema map[string]string = make(map[string]string)
	for line := range strings.SplitSeq(content, "\n") {
		parts := strings.Split(line, "=")
		schema[parts[0]] = filepath.Join(t.homeDir, parts[1])
	}

	// logic
	entries, err := os.ReadDir(colors_dir)
	if err != nil {
		t.logg_book.EnterLogAndPrint("Error while reading colors directory - "+colors_dir, logger.LogTypes.Error, err)
	}
	for _, entry := range entries {
		// entry check
		if entry.IsDir() {
			t.logg_book.EnterLogAndPrint(fmt.Sprintf("Folder %s is skipped", filepath.Join(t.themesDir, entry.Name())), logger.LogTypes.Warning, nil)
		}

		target_path, ok := schema[entry.Name()]
		if !ok {
			t.logg_book.EnterLogAndPrint("unknown colors file found. Skipping - "+entry.Name(), logger.LogTypes.Warning, nil)
			continue
		}

		// applying
		if err := fldir.CopyFile(filepath.Join(colors_dir, entry.Name()), target_path); err != nil {
			t.logg_book.EnterLogAndPrint(err.Error(), logger.LogTypes.Error, err)
		}
	}
}

// stores whether config files are installed or not
func (t *Themer) common_state() bool {
	if fldir.IsPathExist(t.commonStatePath) {
		data, err := fldir.ReadFileAsString(t.commonStatePath)
		if err != nil {
			t.logg_book.EnterLogAndPrint("Error while reading common state - "+t.commonStatePath, logger.LogTypes.Error, err)
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
	if err := fldir.WriteStringToFile(content, t.commonStatePath); err != nil {
		t.logg_book.EnterLogAndPrint(err.Error(), logger.LogTypes.Error, err)
	}
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

	t.logg_book.EnterLogAndPrint("Rofi theme image pair not found.", logger.LogTypes.Error, errors.New("rofi theme image pair not found"))
	return ""
}

// version string parser
func (t *Themer) version_parser(version string) (float64, int) {
	version_parts := strings.Split(version, "-")
	version_fl, err := strconv.ParseFloat(strings.SplitN(version_parts[0], ".", 2)[1], 64)
	if err != nil {
		t.logg_book.Log("Error while parsing version", logger.LogTypes.Error, err)
	}
	sub_version, err := strconv.Atoi(version_parts[1])
	if err != nil {
		t.logg_book.Log("Error while parsing version", logger.LogTypes.Error, err)
	}
	return version_fl, sub_version
}
