package walls

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/wizarki972/myone/internal/common"
	"github.com/wizarki972/myone/internal/utils/cmds"
	"github.com/wizarki972/myone/internal/utils/fldir"
	"github.com/wizarki972/myone/internal/utils/logger"
)

// EXPLORE FOR APPENDING STRINGS IN LOOP - WriteIndex
const ZIPS_DIR_URL = "https://raw.githubusercontent.com/wizarki972/mywalls/main/zips/"
const INDEX_URL = "https://raw.githubusercontent.com/wizarki972/mywalls/main/index.txt"
const VERSION_URL = "https://raw.githubusercontent.com/wizarki972/mywalls/main/VERSION"

var WALLS_DIR = filepath.Join(fldir.GetHomeDir(), common.BASE_DIR, "walls")

type index struct {
	Version float64
	Name    string
	ZipName string
}

// generates Wall struct
func NewWall(logg_book *logger.LogBook) *Wall {
	return &Wall{
		logg_book:          logg_book,
		wallDir:            WALLS_DIR,
		indexPath:          filepath.Join(WALLS_DIR, "index.txt"),
		local_indices:      make(map[string]*index),
		is_local_refreshed: false,
		repo_indices:       make(map[string]*index),
		is_repo_refreshed:  false,
	}
}

type Wall struct {
	logg_book *logger.LogBook
	wallDir   string
	indexPath string

	local_indices      map[string]*index
	is_local_refreshed bool
	repo_indices       map[string]*index
	is_repo_refreshed  bool
}

// loads locally installed pack info
func (w *Wall) RefreshLocalIndices() {
	if w.is_local_refreshed {
		return
	}

	local_index_path := w.indexPath
	if !fldir.IsPathExist(local_index_path) {
		fldir.CreateDirectory(w.wallDir)
		if _, err := os.Create(local_index_path); err != nil {
			w.logg_book.EnterLogAndPrint("Failed to create local index path.", logger.LogTypes.Error, err)
		}
		return
	}

	indices, err := fldir.ReadFileAsString(local_index_path)
	if err != nil {
		w.logg_book.EnterLogAndPrint("Failed to read local index path.", logger.LogTypes.Error, err)
	}

	scanner := bufio.NewScanner(strings.NewReader(indices))
	for scanner.Scan() {
		line := scanner.Text()
		if len(strings.TrimSpace(line)) == 0 {
			continue
		}

		parts := strings.Split(line, "=")
		version, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
		if err != nil {
			w.logg_book.EnterLogAndPrint("Failed to parse float from string.", logger.LogTypes.Error, err)
		}
		w.local_indices[strings.ToLower(strings.TrimSpace(parts[1]))] = &index{
			Version: version,
			Name:    strings.TrimSpace(parts[1]),
			ZipName: strings.TrimSpace(parts[2]),
		}
	}
	w.is_local_refreshed = true
}

// loads info of availbale packs in repo
func (w *Wall) RefreshRepoIndices() {
	if w.is_repo_refreshed {
		return
	}

	indices, err := fldir.ReadTextFileFromURL(INDEX_URL, false, "")
	if err != nil {
		w.logg_book.EnterLogAndPrint(err.Error(), logger.LogTypes.Error, err)
	}

	scanner := bufio.NewScanner(strings.NewReader(indices))
	for scanner.Scan() {
		line := scanner.Text()
		if len(strings.TrimSpace(line)) == 0 {
			continue
		}

		parts := strings.Split(line, " = ")
		version, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
		if err != nil {
			w.logg_book.EnterLogAndPrint("Failed to parse float from string.", logger.LogTypes.Error, err)
		}
		w.repo_indices[strings.ToLower(strings.TrimSpace(parts[1]))] = &index{
			Version: version,
			Name:    strings.TrimSpace(parts[1]),
			ZipName: strings.TrimSpace(parts[2]),
		}
	}
	w.is_repo_refreshed = true
}

// lists all packs installed and not installed with an indicator
func (w *Wall) List() {
	w.RefreshLocalIndices()
	w.RefreshRepoIndices()

	fmt.Println("PACKS:")
	for key, value := range w.repo_indices {
		_, ok := w.local_indices[key]

		switch {
		case ok && w.local_indices[key].Version < value.Version:
			fmt.Printf("* %s - %.2f [UPDATE AVAILABLE current version %.2f]\n", value.Name, value.Version, w.local_indices[key].Version)
		case ok:
			fmt.Printf("* %s - %.2f [INSTALLED]\n", value.Name, value.Version)
		default:
			fmt.Printf("* %s - %.2f\n", value.Name, value.Version)
		}
	}
}

// remove an installed wall pack
func (w *Wall) Remove(pack_name string) {
	w.RefreshLocalIndices()
	w.RefreshRepoIndices()

	pack_name_lc := strings.ToLower(pack_name)
	pack, ok := w.local_indices[pack_name_lc]
	if !ok {
		w.logg_book.EnterLogAndPrint("Wallpaper package not found to remove.", logger.LogTypes.Error, errors.New("pack not found"))
	}

	if err := os.RemoveAll(filepath.Join(w.wallDir, pack.Name)); err != nil {
		w.logg_book.EnterLogAndPrint("Failed to remove wallpaper package.", logger.LogTypes.Error, errors.New("pack not found"))
	}
	delete(w.local_indices, pack_name_lc)
	w.WriteIndex()
}

// install a wall pack from repo
func (w *Wall) Install(pack_name string) {
	w.RefreshRepoIndices()
	w.RefreshLocalIndices()

	pack_name_lc := strings.ToLower(pack_name)
	// Pack's existence
	pack, ok := w.repo_indices[pack_name_lc]
	if !ok {
		w.logg_book.EnterLogAndPrint("Wallpaper package not found in repo.", logger.LogTypes.Error, errors.New("pack not found"))
	}

	// DOWNLOADING WALL PACK
	cache_path := filepath.Join(fldir.GetHomeDir(), common.CACHE_DIR, "walls", pack.ZipName)
	if err := fldir.DownloadURL(ZIPS_DIR_URL+w.repo_indices[pack_name_lc].ZipName, cache_path, true); err != nil {
		w.logg_book.EnterLogAndPrint(err.Error(), logger.LogTypes.Error, err)
	}

	// UNZIPPING PACK
	w.logg_book.EnterLogAndPrint("Extracting Wallpaper pack...", logger.LogTypes.Info, nil)
	destination := filepath.Join(w.wallDir, pack.Name)
	if err := os.RemoveAll(destination); err != nil {
		w.logg_book.EnterLogAndPrint("Failed to remove old wallpaper package of "+pack.Name+".", logger.LogTypes.Error, errors.New("pack not found"))
	}
	fldir.CreateDirectory(destination)
	if err := fldir.Unzip(cache_path, destination); err != nil {
		w.logg_book.EnterLogAndPrint(err.Error(), logger.LogTypes.Error, err)
	}

	// ADDING INDEX
	w.local_indices[pack_name_lc] = w.repo_indices[pack_name_lc]
	w.WriteIndex()

	// CLEANING UP
	w.logg_book.EnterLogAndPrint("Clearing cache...", logger.LogTypes.Info, nil)
	if err := os.RemoveAll(filepath.Dir(cache_path)); err != nil {
		w.logg_book.EnterLogAndPrint("Failed to clean up cache.", logger.LogTypes.Warning, nil)
	}

	w.logg_book.EnterLogAndPrint("Installed "+pack.Name+" wallpack.", logger.LogTypes.Info, nil)
}

// writes the index of locally installed wall packs
func (w *Wall) WriteIndex() {
	var b strings.Builder
	for _, v := range w.local_indices {
		fmt.Fprintf(&b, "%.2f=%s=%s\n", v.Version, v.Name, v.ZipName)
	}

	fldir.WriteStringToFile(b.String(), w.indexPath)
}

// shows the wallpaper change menu using rofi
func (w *Wall) ShowWallpaperChangeMenu() {
	w.RefreshLocalIndices()

	home := fldir.GetHomeDir()

	// PACK MENU
	rofi_input, err := w.rofiWallMenuBuilder(w.wallDir, "dir")
	if err != nil {
		w.logg_book.EnterLogAndPrint("Error while building rofi menu for wallpaper packs.", logger.LogTypes.Error, err)
	}
	command := fmt.Sprintf("printf '%s' | rofi -dmenu -theme %s/.config/rofi/clipboard.rasi", rofi_input, home)
	selected_pack, err := cmds.ExecCommand(command, false, true)
	if err != nil {
		w.logg_book.EnterLogAndPrint("Failed to execute command - "+command, logger.LogTypes.Error, err)
	}

	// WALLS MENU
	pack_dir := filepath.Join(w.wallDir, strings.TrimSpace(string(selected_pack)))
	rofi_input, err = w.rofiWallMenuBuilder(pack_dir, "")
	if err != nil {
		w.logg_book.EnterLogAndPrint("Error while building rofi menu for wallpapers.", logger.LogTypes.Error, err)
	}
	cmd := exec.Command(
		"rofi",
		"-dmenu",
		"-show-icons",
		"-i",
		"-theme",
		filepath.Join(home, ".config/rofi/wallpapers.rasi"),
	)
	cmd.Stdin = strings.NewReader(rofi_input)
	selection, err := cmd.Output()
	if err != nil {
		w.logg_book.EnterLogAndPrint("Failed to execute rofi wallpaper menu command.", logger.LogTypes.Error, err)
	}

	// SETTING WALL
	current_wall_path := filepath.Join(home, common.CURRENT_WALLPAPER_ENTRY_PATH)
	fldir.CopyFile(filepath.Join(pack_dir, strings.TrimSpace(string(selection))), current_wall_path)
	command = fmt.Sprintf("awww img %s --transition-type fade --transition-duration 0.5", current_wall_path)
	if _, err = cmds.ExecCommand(command, false, false); err != nil {
		w.logg_book.EnterLogAndPrint("Failed to execute command - "+command, logger.LogTypes.Error, err)
	}
}

// reads the walls directory and creates a list for rofi to display
func (w *Wall) rofiWallMenuBuilder(dir_path, mode string) (string, error) {
	entries, err := os.ReadDir(dir_path)
	if err != nil {
		return "", err
	}

	if len(entries) == 0 {
		command := "notify-send 'No Wallpapers' 'Install a wallpaper package.\nRun `myone wallpapers --list-repo` command to see available packages.'"
		cmds.ExecCommand(command, false, false)
		return "", errors.New("No wallpaper package is installed")
	}

	var rofi_input strings.Builder
	for _, entry := range entries {
		if mode == "dir" {
			if entry.IsDir() {
				rofi_input.WriteString(entry.Name() + "\n")
			}
		} else {
			ext := filepath.Ext(entry.Name())
			switch ext {
			case ".jpeg", ".jpg", ".png", ".gif":
				rofi_input.WriteString(entry.Name() + "\x00icon\x1f" + filepath.Join(dir_path, entry.Name()) + "\n")
			}

		}
	}
	if rofi_input.Len() == 0 {
		w.logg_book.Log("No wallpapers found in the directory, install a wallpack.", logger.LogTypes.Error, errors.New("no wallpapers found in the directory, install a wallpack"))
	}
	return rofi_input.String(), nil
}
