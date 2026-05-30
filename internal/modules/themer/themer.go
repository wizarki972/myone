package themer

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wizarki972/myone/internal/common"
	"github.com/wizarki972/myone/internal/utils/fldir"
	"github.com/wizarki972/myone/internal/utils/logger"
	"github.com/wizarki972/myone/internal/utils/pkg"
	"github.com/wizarki972/myone/internal/utils/release"
)

// Maybe make choosing and installing themes - in future
const THEMES_REPO_NAME = "mythemes"
const DEFAULT_THEME = "tokyonight"

// Themer struct generator
func NewThemer(themeName string, loggBook *logger.LogBook) *Themer {
	home := fldir.GetHomeDir()

	var t = &Themer{
		release:         nil,
		isLocalVerFound: false,
		loggBook:        loggBook,
		homeDir:         home,
		themesDir:       filepath.Join(home, common.THEMES_DIR),

		commonStatePath:      filepath.Join(home, common.COMMON_PLACED_STATE_PATH),
		currentThemeNamePath: filepath.Join(home, common.CURRENT_THEME_NAME_ENTRY_PATH),
	}
	t.versionPath = filepath.Join(t.themesDir, "VERSION")
	t.CheckFiles()

	// if a theme name is given
	if len(themeName) > 0 {
		if themeName == "default" {
			t.ThemeName = DEFAULT_THEME
		} else {
			t.ThemeName = themeName
		}
		return t
	}

	// current theme
	if fldir.IsPathExist(t.currentThemeNamePath) {
		var err error
		t.ThemeName, err = fldir.ReadFileAsString(t.currentThemeNamePath)
		if err != nil {
			t.loggBook.EnterLogAndPrint("Cannot get current theme name from the path - "+t.currentThemeNamePath, logger.LogTypes.Error, err)
		}

		// second check for errors
		if len(strings.TrimSpace(t.ThemeName)) == 0 {
			t.loggBook.EnterLogAndPrint("Theme name from the path - "+t.currentThemeNamePath+" is empty. Using default theme.", logger.LogTypes.Warning, nil)
			t.ThemeName = DEFAULT_THEME
		}
		return t
	} else {
		t.ThemeName = DEFAULT_THEME
		return t
	}
}

type Themer struct {
	ThemeName string
	Version   struct {
		Major int
		Minor int
		Patch int
	}
	areFilesFound   bool
	isLocalVerFound bool
	versionPath     string
	homeDir         string
	themesDir       string

	commonStatePath      string
	currentThemeNamePath string

	loggBook               *logger.LogBook
	release                *release.Release
	themePlaceholderValues map[string]string
}

// Updates the config/theme files if a new version is available
func (t *Themer) Update() {
	if t.release == nil {
		t.FetchRelease()
	}

	if !t.isLocalVerFound {
		t.FetchLocalVersion()
	}

	if t.areFilesFound {
		isNewer, err := release.IsNewer(t.release, t.Version.Major, t.Version.Minor, t.Version.Patch)
		if err != nil {
			t.loggBook.EnterLogAndPrint(err.Error(), logger.LogTypes.Error, err)
		}

		if isNewer {
			t.Download()
			t.Install()
			return
		}
		t.loggBook.EnterLogAndPrint("Theme files are up-to-date.", logger.LogTypes.Info, nil)
	} else {
		t.loggBook.EnterLogAndPrint("No files found, downloading them...", logger.LogTypes.Warning, nil)
		t.Download()
		t.Install()
	}
}

// Downloads files from the repo
func (t *Themer) Download() {
	// DOWNLOADING LATEST RELEASE
	tempPath, err := release.DownloadLatestRelease(t.release)
	if err != nil {
		t.loggBook.EnterLogAndPrint(err.Error(), logger.LogTypes.Error, err)
	}

	// REMOVING CURRENTLY INSTALLED FILES
	t.loggBook.EnterLogAndPrint("Preparing path for extraction...", logger.LogTypes.Info, nil)
	if err := os.RemoveAll(t.themesDir); err != nil {
		t.loggBook.EnterLogAndPrint("Failed to remove old themes.", logger.LogTypes.Error, err)
	}
	if err := fldir.CreateDirectory(t.themesDir); err != nil {
		t.loggBook.EnterLogAndPrint(err.Error(), logger.LogTypes.Error, err)
	}

	// EXTRACTING
	t.loggBook.EnterLogAndPrint("Extracting files...", logger.LogTypes.Info, nil)
	if err := fldir.Unzip(tempPath, t.themesDir); err != nil {
		t.loggBook.EnterLogAndPrint("Failed to unzip the downloaded files.", logger.LogTypes.Error, err)
	}
	t.areFilesFound = true

	// CLEARING CACHE
	t.loggBook.EnterLogAndPrint("Cleaning up...", logger.LogTypes.Info, nil)
	if err := os.RemoveAll(tempPath); err != nil {
		t.loggBook.EnterLogAndPrint("Failed to clear cache.", logger.LogTypes.Warning, nil)
	}
}

// Installs the downloaded config/theme files
func (t *Themer) Install() {
	if !t.areFilesFound {
		t.loggBook.EnterLogAndPrint("Cannot download theme/config files.", logger.LogTypes.Error, errors.New("cannot download theme/config files"))
		return
	}
	t.generatePlaceholderValues()

	// placing files
	t.placeCommonFiles()

	// theme dependent file
	t.placeThemeDependentFiles()

	// applying colors
	t.apply_colors()

	// dependency check
	dep_lst_path := filepath.Join(t.themesDir, "deps.lst")
	if fldir.IsPathExist(dep_lst_path) {
		t.loggBook.EnterLogAndPrint("Performing dependency check for the themes...", logger.LogTypes.Info, nil)
		if err := pkg.InstallPkgsFromFile(dep_lst_path); err != nil {
			t.loggBook.EnterLogAndPrint(err.Error(), logger.LogTypes.Error, err)
		}
	}

	// writing current theme name to a file
	if err := fldir.WriteStringToFile(t.ThemeName, t.currentThemeNamePath); err != nil {
		t.loggBook.EnterLogAndPrint(err.Error(), logger.LogTypes.Error, err)
	}

	// to refresh desktop after update
	t.refreshDesktop()
}

// applies themes
func (t *Themer) ApplyTheme() {
	if !t.areFilesFound {
		t.loggBook.EnterLogAndPrint("No files found. downloading them...", logger.LogTypes.Warning, nil)
		t.Download()
		t.Install()
	}

	// placeholder values
	t.generatePlaceholderValues()

	// checking common state
	if !t.common_state() {
		t.placeCommonFiles()
	}

	// placing theme based files
	t.placeThemeDependentFiles()

	// applying colors
	t.apply_colors()

	// writing current theme name to a file
	if err := fldir.WriteStringToFile(t.ThemeName, t.currentThemeNamePath); err != nil {
		t.loggBook.EnterLogAndPrint(err.Error(), logger.LogTypes.Error, err)
	}

	// to refresh desktop after update
	t.refreshDesktop()
}

// changes colors files based on the theme
func (t *Themer) apply_colors() {
	colors_dir := filepath.Join(t.themesDir, "colors", t.ThemeName)

	// Checks
	info, err := os.Stat(colors_dir)
	if err != nil {
		t.loggBook.EnterLogAndPrint("Cannot access colors directory or not found. ("+colors_dir+").", logger.LogTypes.Error, err)
	}
	if !info.IsDir() {
		t.loggBook.EnterLogAndPrint("Colors directory not found for this theme - "+colors_dir, logger.LogTypes.Error, errors.New("colors directory not found for this theme"))
	}

	// loading schema
	schema_path := filepath.Join(t.themesDir, "colors", "schema")
	content, err := fldir.ReadFileAsString(schema_path)
	if err != nil {
		t.loggBook.EnterLogAndPrint("Error while reading schema - "+schema_path, logger.LogTypes.Error, err)
	}
	var schema map[string]string = make(map[string]string)
	for line := range strings.SplitSeq(content, "\n") {
		parts := strings.Split(line, "=")
		schema[parts[0]] = filepath.Join(t.homeDir, parts[1])
	}

	// logic
	entries, err := os.ReadDir(colors_dir)
	if err != nil {
		t.loggBook.EnterLogAndPrint("Error while reading colors directory - "+colors_dir, logger.LogTypes.Error, err)
	}
	for _, entry := range entries {
		// entry check
		if entry.IsDir() {
			t.loggBook.EnterLogAndPrint(fmt.Sprintf("Folder %s is skipped", filepath.Join(t.themesDir, entry.Name())), logger.LogTypes.Warning, nil)
		}

		target_path, ok := schema[entry.Name()]
		if !ok {
			t.loggBook.EnterLogAndPrint("unknown colors file found. Skipping - "+entry.Name(), logger.LogTypes.Warning, nil)
			continue
		}

		// applying
		if err := fldir.CopyFile(filepath.Join(colors_dir, entry.Name()), target_path); err != nil {
			t.loggBook.EnterLogAndPrint(err.Error(), logger.LogTypes.Error, err)
		}
	}
}

// stores whether config files/initial setup is complete or not
func (t *Themer) common_state() bool {
	data, err := fldir.ReadFileAsString(t.commonStatePath)
	if errors.Is(err, os.ErrNotExist) {
		return false
	}
	if err != nil {
		t.loggBook.EnterLogAndPrint("Error while reading common state - "+t.commonStatePath, logger.LogTypes.Error, err)
	}
	if data == "1" {
		return true
	}
	return false
}

// sets common state - indicates whether files/initial setup is complete or not
func (t *Themer) set_common_state(state bool) {
	var content string
	if state {
		content = "1"
	} else {
		content = "0"
	}
	if err := fldir.WriteStringToFile(content, t.commonStatePath); err != nil {
		t.loggBook.EnterLogAndPrint(err.Error(), logger.LogTypes.Error, err)
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

	t.loggBook.EnterLogAndPrint("Rofi theme image pair not found.", logger.LogTypes.Error, errors.New("rofi theme image pair not found"))
	return ""
}

func (t *Themer) FetchLocalVersion() {
	if !t.areFilesFound {
		t.loggBook.EnterLogAndPrint("No themes installed.", logger.LogTypes.Error, errors.New("no themes installed"))
	}

	if t.areFilesFound {
		versionStr, err := fldir.ReadFileAsString(t.versionPath)
		if err != nil {
			t.loggBook.EnterLogAndPrint("Cannot read version file - "+t.versionPath, logger.LogTypes.Error, err)
		}

		t.Version.Major, t.Version.Minor, t.Version.Patch, err = release.VersionParser(versionStr)
		if err != nil {
			t.loggBook.EnterLogAndPrint(err.Error(), logger.LogTypes.Error, err)
		}
		t.isLocalVerFound = true
		return
	}

	t.loggBook.EnterLogAndPrint("Cannot fetch local themes/config files version from path - "+t.versionPath+", make sure that they are already downloaded.", logger.LogTypes.Error, errors.New("cannot fetch local themes/config files version from path - "+t.versionPath+", make sure that they are already downloaded"))
}

func (t *Themer) CheckFiles() {
	entries, err := os.ReadDir(t.themesDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			t.areFilesFound = false
			return
		}
		t.loggBook.EnterLogAndPrint(err.Error(), logger.LogTypes.Error, err)
	}
	t.areFilesFound = len(entries) > 7
}

func (t *Themer) FetchRelease() {
	if t.release != nil {
		return
	}
	var err error
	t.release, err = release.GetLatestRelease(THEMES_REPO_NAME)
	if err != nil {
		t.loggBook.EnterLogAndPrint("Cannot find latest release from github.", logger.LogTypes.Error, err)
	}
}
