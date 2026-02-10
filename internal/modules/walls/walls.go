package walls

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/wizarki972/myone/internal/utils/cmds"
	"github.com/wizarki972/myone/internal/utils/config"
	themes_config "github.com/wizarki972/myone/internal/utils/config/themes"
	"github.com/wizarki972/myone/internal/utils/fldir"
	"github.com/wizarki972/myone/internal/utils/user"
)

// EXPLORE FOR APPENDING STRINGS IN LOOP - WriteIndex

const ZIPS_DIR_URL = "https://raw.githubusercontent.com/wizarki972/mywalls/main/zips/"
const INDEX_URL = "https://raw.githubusercontent.com/wizarki972/mywalls/main/index.txt"
const VERSION_URL = "https://raw.githubusercontent.com/wizarki972/mywalls/main/VERSION"

type index struct {
	Version float64
	Name    string
	ZipName string
}

func NewWall() *Wall {
	return &Wall{
		indexPath:          filepath.Join(config.GetDirPathFor("walls"), "index.txt"),
		local_indices:      make(map[string]*index),
		is_local_refreshed: false,
		repo_indices:       make(map[string]*index),
		is_repo_refreshed:  false,
	}
}

type Wall struct {
	indexPath          string
	local_indices      map[string]*index
	is_local_refreshed bool
	repo_indices       map[string]*index
	is_repo_refreshed  bool
}

func (w *Wall) RefreshLocalIndices() {
	if w.is_local_refreshed {
		return
	}

	local_index_path := filepath.Join(config.GetDirPathFor("walls"), "index.txt")
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

func (w *Wall) ListDownloadables() {
	w.RefreshRepoIndices()
	w.RefreshLocalIndices()

	fmt.Println("PACKS FOR DOWNLOAD:")
	for key, value := range w.repo_indices {
		_, ok := w.local_indices[key]
		if ok && w.local_indices[key].Version < value.Version {
			fmt.Printf("%.2f - %s [UPDATE AVAILABLE current version %.2f]\n", value.Version, key, w.local_indices[key].Version)
		}

		if !ok {
			fmt.Printf("%.2f - %s\n", value.Version, key)
		}
	}
}

func (w *Wall) ListInstalled() {
	w.RefreshLocalIndices()
	w.RefreshRepoIndices()

	fmt.Println("INSTALLED WALLPAPER PACKAGES:")
	for key, value := range w.local_indices {
		if w.repo_indices[key].Version > value.Version {
			fmt.Printf("%.2f - %s [UPDATE AVAILABLE new version %.2f]\n", value.Version, key, w.repo_indices[key].Version)
			continue
		}
		fmt.Printf("%.2f - %s\n", value.Version, key)
	}
}

func (w *Wall) Remove(pack_name string) {
	w.RefreshLocalIndices()
	w.RefreshRepoIndices()

	if err := os.RemoveAll(filepath.Join(config.GetDirPathFor("walls"), pack_name)); err != nil {
		panic(err)
	}
	delete(w.local_indices, pack_name)
	w.WriteIndex()
}

func (w *Wall) Install(pack_name string) {
	w.RefreshRepoIndices()
	w.RefreshLocalIndices()

	// DOWNLOADING WALL PACK
	cache_path := filepath.Join(config.GetDirPathFor("cache"), "walls", w.repo_indices[pack_name].ZipName)
	fldir.DownloadURL(ZIPS_DIR_URL+w.repo_indices[pack_name].ZipName, cache_path)

	// UNZIPPING PACK
	fmt.Println("EXTRACTING WALLPAPERS...")
	destination := filepath.Join(config.GetDirPathFor("walls"), pack_name)
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
	// var content = ""
	// for _, index := range w.local_indices {
	// 	content += fmt.Sprintf("%.2f = %s = %s", index.Version, index.Name, index.ZipName)
	// }

	var b strings.Builder
	for _, v := range w.local_indices {
		fmt.Fprintf(&b, "%.2f = %s = %s\n", v.Version, v.Name, v.ZipName)
	}

	fldir.WriteFile(b.String(), w.indexPath)
}

func (w *Wall) ShowWallpaperChangeMenu() {
	w.RefreshLocalIndices()

	// PACK MENU
	rofi_input := rofiWallMenuBuilder(config.GetDirPathFor("walls"), "dir")
	command := fmt.Sprintf("printf '%s' | rofi -dmenu -theme %s/.config/rofi/themes/wallpapers.rasi", rofi_input, user.GetHomeDir())
	selected_pack, err := cmds.ExecCommand(command)
	if err != nil {
		panic(err)
	}

	// WALLS MENU
	pack_dir := filepath.Join(config.GetDirPathFor("walls"), strings.TrimSpace(string(selected_pack)))
	rofi_input = rofiWallMenuBuilder(pack_dir, "")
	cmd := exec.Command(
		"rofi",
		"-dmenu",
		"-show-icons",
		"-i",
		"-theme",
		filepath.Join(user.GetHomeDir(), ".config/rofi/themes/wallpapers.rasi"),
	)
	cmd.Stdin = strings.NewReader(rofi_input)
	selection, err := cmd.Output()
	if err != nil {
		panic(err)
	}

	// SETTING WALL
	fldir.CopyFile(filepath.Join(pack_dir, strings.TrimSpace(string(selection))), filepath.Join(config.GetDirPathFor("base"), themes_config.CURRENT_WALL_NAME))
	command = fmt.Sprintf("swww img %s --transition-type fade --transition-duration 0.5", filepath.Join(config.GetDirPathFor("base"), themes_config.CURRENT_WALL_NAME))
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
