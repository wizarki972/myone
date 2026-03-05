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

		current, err := fldir.ReadFileAsString(t.currentThemeNamePath)
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

	if (local_v < repo_v) || (local_v == repo_v && local_sv < repo_sv) {
		t.Download()
		t.Install()
		return
	} else {
		fmt.Println("Themes are already up-to-date.")
	}
}

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

func (t *Themer) Install() {
	t.themePlaceholderValues = map[string]string{
		"${SCRIPTS_DIRECTORY_PATH}":   filepath.Join(t.homeDir, common.SCRIPTS_DIR),
		"${CURRENT_WALLPAPER_PATH}":   filepath.Join(t.homeDir, common.CURRENT_WALLPAPER_ENTRY_PATH),
		"${ALL_WALLS_DIRECTORY_PATH}": filepath.Join(t.homeDir, common.ALL_WALLS_DIR),
		"${ROFI_IMAGE}":               t.get_rofi_image(),
		"${SCREEN_WIDTH}":             strconv.Itoa(display.GetScreenResolution()[0]),
		"${SCREEN_HEIGHT}":            strconv.Itoa(display.GetScreenResolution()[1]),
	}

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
	t.Dependency_check()

	// t.copy_files(themepath, "")
	fldir.WriteStringToFile(t.ThemeName, t.currentThemeNamePath)
}

func (t *Themer) Dependency_check() {
	path := filepath.Join(t.themesDir, "deps.lst")
	if !fldir.IsPathExist(path) {
		return
	}

	content, err := fldir.ReadFileAsString(path)
	if err != nil {
		panic(err)
	}

	var packages strings.Builder
	for line := range strings.SplitSeq(content, "\n") {
		pkg := strings.TrimSpace(line)
		if !bootstrap.Is_dependency_installed(pkg) {
			packages.WriteString(pkg)
			packages.WriteString(" ")
		}
	}

	// check
	if packages.Len() == 0 {
		slog.Info("SKIPPING DEPENDENCY CHECK :: All dependencies are already installed")
		return
	}

	if err := bootstrap.Dependency_install(packages.String()); err != nil {
		slog.Error(err.Error())
	}
}

func (t *Themer) Apply_Theme() {
	if !t.common_state() {
		t.place_common_files()
	}
	t.place_theme_dependent_files()
	t.apply_colors()
	fldir.WriteStringToFile(t.ThemeName, t.currentThemeNamePath)
}

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
	for _, line := range strings.Split(content, "\n") {
		parts := strings.Split(line, "=")
		schema[strings.TrimSpace(parts[0])] = filepath.Join(t.homeDir, strings.TrimSpace(parts[1]))
	}

	// logic
	entries, err := os.ReadDir(colors_dir)
	if err != nil {
		panic(err)
	}
	for _, entry := range entries {
		// entry check
		target_path, ok := schema[entry.Name()]
		if !ok {
			slog.Warn("unknown colors file found. Skipping")
			continue
		}

		// applying
		fldir.CopyFile(filepath.Join(colors_dir, entry.Name()), target_path)
	}
}

func (t *Themer) place_theme_dependent_files() {
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
	if err := t.place_files_logic(td_path, "", true); err != nil {
		slog.Warn(err.Error())
	}

}

func (t *Themer) place_common_files() {
	common_dir := filepath.Join(t.themesDir, "common")
	if !fldir.IsPathExist(common_dir) {
		slog.Info("Theme not found, trying to update themes...")
		t.Download()
	}

	if err := t.place_files_logic(common_dir, "", false); err != nil {
		panic(err)
	}

	t.set_common_state(true)
}

func (t *Themer) place_files_logic(path, suffix string, force_fill bool) error {
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
			t.place_files_logic(entry_path, filepath.Join(suffix, entry.Name()), force_fill)
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

func (t *Themer) common_state() bool {
	if fldir.IsPathExist(t.commonStatePath) {
		data, err := fldir.ReadFileAsString(t.commonStatePath)
		if err != nil {
			panic(err)
		}
		if strings.TrimSpace(data) == "1" {
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
