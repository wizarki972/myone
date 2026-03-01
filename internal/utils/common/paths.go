package common

import (
	"path/filepath"

	"github.com/wizarki972/myone/internal/utils/fldir"
)

const BASE_DIR = ".local/share/myone"
const CONFIG_DIR = ".config/myone"
const CACHE_DIR = ".cache/myone"

var PseudoPaths = map[string]string{
	"base":    BASE_DIR,
	"config":  filepath.Join(CONFIG_DIR, "config.toml"),
	"scripts": filepath.Join(BASE_DIR, "scripts"),
}

func GetDirPathFor(value string) string {
	home := fldir.GetHomeDir()
	return filepath.Join(home, PseudoPaths[value])
}
