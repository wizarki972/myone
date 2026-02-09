package themer

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/wizarki972/myone/internal/utils/cmds"
	"github.com/wizarki972/myone/internal/utils/config"
	themes_config "github.com/wizarki972/myone/internal/utils/config/themes"
	"github.com/wizarki972/myone/internal/utils/fldir"
	"github.com/wizarki972/myone/internal/utils/user"
)

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
	var err error

	themes_path := config.GetDirPathFor("themes")
	fldir.CreateDirectory(themes_path)

	cache_dir := filepath.Join(config.GetDirPathFor("cache"), "themes")
	cache_path := filepath.Join(config.GetDirPathFor("cache"), "themes/mythemes.zip")
	fldir.CreateDirectory(cache_dir)

	slog.Info("Downloading themes...")
	command := "curl -L https://github.com/wizarki972/mythemes/archive/refs/heads/main.zip -o " + cache_path
	if err = cmds.ExecComamndWithError(command); err != nil {
		slog.Error("Failed to download themes")
		os.Exit(1)
	}

	slog.Info("Removing currently installed themes...")
	if err := os.RemoveAll(themes_path); err != nil {
		slog.Error("Failed to remove themes ==> " + err.Error())
	}
	fldir.CreateDirectory(themes_path)

	slog.Info("Installing themes...")
	command = fmt.Sprintf("unzip -o %s -d %s && mv %s/mythemes-main/* %s", cache_path, cache_dir, cache_dir, themes_path)
	if err = cmds.ExecComamndWithError(command); err != nil {
		slog.Error("Failed during installing downloaded themes")
		os.Exit(1)
	}

	slog.Info("Cleaning Up...")
	if err := os.RemoveAll(cache_dir); err != nil {
		slog.Error("Failed to remove cache")
	}
	slog.Info("You can install the downloaded themes now...")
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
