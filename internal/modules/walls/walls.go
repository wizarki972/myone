package walls

import (
	"bufio"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/wizarki972/myone/internal/common"
	"github.com/wizarki972/myone/internal/utils/cmds"
	"github.com/wizarki972/myone/internal/utils/fldir"
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

func NewWall() *Wall {
	return &Wall{
		wallDir:            WALLS_DIR,
		indexPath:          filepath.Join(WALLS_DIR, "index.txt"),
		local_indices:      make(map[string]*index),
		is_local_refreshed: false,
		repo_indices:       make(map[string]*index),
		is_repo_refreshed:  false,
	}
}

type Wall struct {
	wallDir   string
	indexPath string

	local_indices      map[string]*index
	is_local_refreshed bool
	repo_indices       map[string]*index
	is_repo_refreshed  bool
}

func (w *Wall) RefreshLocalIndices() {
	if w.is_local_refreshed {
		return
	}

	local_index_path := w.indexPath
	if !fldir.IsPathExist(local_index_path) {
		fldir.CreateDirectory(w.wallDir)
		if _, err := os.Create(local_index_path); err != nil {
			panic(err)
		}
		return
	}

	indices, err := fldir.ReadFileAsString(local_index_path)
	if err != nil {
		panic(err)
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
			panic(err)
		}
		w.local_indices[strings.TrimSpace(parts[1])] = &index{
			Version: version,
			Name:    strings.TrimSpace(parts[1]),
			ZipName: strings.TrimSpace(parts[2]),
		}
	}
	w.is_local_refreshed = true
}

func (w *Wall) RefreshRepoIndices() {
	if w.is_repo_refreshed {
		return
	}

	indices := fldir.ReadTextFileFromURL(INDEX_URL, false, "")

	scanner := bufio.NewScanner(strings.NewReader(indices))
	for scanner.Scan() {
		line := scanner.Text()
		if len(strings.TrimSpace(line)) == 0 {
			continue
		}

		parts := strings.Split(line, " = ")
		version, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
		if err != nil {
			panic(err)
		}
		w.repo_indices[strings.TrimSpace(parts[1])] = &index{
			Version: version,
			Name:    strings.TrimSpace(parts[1]),
			ZipName: strings.TrimSpace(parts[2]),
		}
	}
	w.is_repo_refreshed = true
}

func (w *Wall) List() {
	w.RefreshLocalIndices()
	w.RefreshRepoIndices()

	fmt.Println("PACKS:")
	for key, value := range w.repo_indices {
		_, ok := w.local_indices[key]

		switch {
		case ok && w.local_indices[key].Version < value.Version:
			fmt.Printf("* %s - %.2f [UPDATE AVAILABLE current version %.2f]\n", key, value.Version, w.local_indices[key].Version)
		case ok:
			fmt.Printf("* %s - %.2f [INSTALLED]\n", key, value.Version)
		default:
			fmt.Printf("* %s - %.2f\n", key, value.Version)
		}
	}
}

// func (w *Wall) ListDownloadables() {
// 	w.RefreshRepoIndices()
// 	w.RefreshLocalIndices()

// 	fmt.Println("PACKS FOR DOWNLOAD:")
// 	for key, value := range w.repo_indices {
// 		_, ok := w.local_indices[key]
// 		if ok && w.local_indices[key].Version < value.Version {
// 			fmt.Printf("%.2f - %s [UPDATE AVAILABLE current version %.2f]\n", value.Version, key, w.local_indices[key].Version)
// 		}

// 		if !ok {
// 			fmt.Printf("%.2f - %s\n", value.Version, key)
// 		}
// 	}
// }

// func (w *Wall) ListInstalled() {
// 	w.RefreshLocalIndices()
// 	w.RefreshRepoIndices()

// 	fmt.Println("INSTALLED WALLPAPER PACKAGES:")
// 	for key, value := range w.local_indices {
// 		if w.repo_indices[key].Version > value.Version {
// 			fmt.Printf("%.2f - %s [UPDATE AVAILABLE new version %.2f]\n", value.Version, key, w.repo_indices[key].Version)
// 			continue
// 		}
// 		fmt.Printf("%.2f - %s\n", value.Version, key)
// 	}
// }

func (w *Wall) Remove(pack_name string) {
	w.RefreshLocalIndices()
	w.RefreshRepoIndices()

	if err := os.RemoveAll(filepath.Join(w.wallDir, pack_name)); err != nil {
		panic(err)
	}
	delete(w.local_indices, pack_name)
	w.WriteIndex()
}

func (w *Wall) Install(pack_name string) {
	w.RefreshRepoIndices()
	w.RefreshLocalIndices()

	// Pack's existence
	pack, ok := w.repo_indices[capitalizeFirst(pack_name)]
	if !ok {
		slog.Error(fmt.Sprintf("%s pack not available", pack_name))
		os.Exit(1)
	}

	// DOWNLOADING WALL PACK
	cache_path := filepath.Join(fldir.GetHomeDir(), common.CACHE_DIR, "walls", pack.ZipName)
	fldir.DownloadURL(ZIPS_DIR_URL+w.repo_indices[pack_name].ZipName, cache_path, true)

	// UNZIPPING PACK
	fmt.Println("EXTRACTING WALLPAPERS...")
	destination := filepath.Join(w.wallDir, pack_name)
	fldir.CreateDirectory(destination)
	fldir.Unzip(cache_path, destination)

	// ADDING INDEX
	w.local_indices[pack_name] = w.repo_indices[pack_name]
	w.WriteIndex()

	// CLEANING UP
	if err := os.RemoveAll(filepath.Dir(cache_path)); err != nil {
		fmt.Println("ERROR: Failed to clean up cache")
	}

	fmt.Printf("INSTALLED %s Wallpaper pack", pack_name)
}

func (w *Wall) WriteIndex() {
	var b strings.Builder
	for _, v := range w.local_indices {
		fmt.Fprintf(&b, "%.2f = %s = %s\n", v.Version, v.Name, v.ZipName)
	}

	fldir.WriteStringToFile(b.String(), w.indexPath)
}

func (w *Wall) ShowWallpaperChangeMenu() {
	w.RefreshLocalIndices()

	home := fldir.GetHomeDir()

	// PACK MENU
	rofi_input := rofiWallMenuBuilder(w.wallDir, "dir")
	command := fmt.Sprintf("printf '%s' | rofi -dmenu -theme %s/.config/rofi/themes/wallpapers.rasi", rofi_input, home)
	selected_pack, err := cmds.ExecCommand(command)
	if err != nil {
		panic(err)
	}

	// WALLS MENU
	pack_dir := filepath.Join(w.wallDir, strings.TrimSpace(string(selected_pack)))
	rofi_input = rofiWallMenuBuilder(pack_dir, "")
	cmd := exec.Command(
		"rofi",
		"-dmenu",
		"-show-icons",
		"-i",
		"-theme",
		filepath.Join(home, ".config/rofi/themes/wallpapers.rasi"),
	)
	cmd.Stdin = strings.NewReader(rofi_input)
	selection, err := cmd.Output()
	if err != nil {
		panic(err)
	}

	// SETTING WALL
	current_wall_path := filepath.Join(home, common.BASE_DIR, CURRENT_WALL_NAME)
	fldir.CopyFile(filepath.Join(pack_dir, strings.TrimSpace(string(selection))), current_wall_path)
	command = fmt.Sprintf("swww img %s --transition-type fade --transition-duration 0.5", current_wall_path)
	if err = cmds.ExecComamndWithError(command); err != nil {
		panic(err)
	}
}

func rofiWallMenuBuilder(dir_path, mode string) string {
	entries, err := os.ReadDir(dir_path)
	if err != nil {
		panic(err)
	}

	var rofi_input = ""
	if len(entries) == 0 {
		command := "notify-send 'No Wallpapers' 'Install a wallpaper package.\nRun `myone wallpapers --list-repo` command to see available packages.'"
		cmds.ExecCommandNoFeedback(command)
		panic(errors.New("No wallpaper package is installed"))
	}

	for _, entry := range entries {
		if mode == "dir" {
			if entry.IsDir() {
				rofi_input += entry.Name() + "\n"
			}
		} else {
			ext := filepath.Ext(entry.Name())
			switch ext {
			case ".jpeg", ".jpg", ".png", ".gif":
				rofi_input += entry.Name() + "\x00icon\x1f" + filepath.Join(dir_path, entry.Name()) + "\n"
			}

		}
	}
	return rofi_input
}

func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	r, size := utf8.DecodeRuneInString(s)
	return strings.ToUpper(string(r)) + s[size:]
}
