package config

import (
	"path/filepath"

	"github.com/wizarki972/myone/internal/utils/user"
)

const BASE_DIR = ".local/share/myone"
const CACHE_BASE_DIR = ".cache/myone"
const CONFIG_DIR = ".config/myone"

var PseudoPaths = map[string]string{
	"config":     filepath.Join(CONFIG_DIR, "config.toml"),
	"cache":      CACHE_BASE_DIR,
	"scripts":    filepath.Join(BASE_DIR, "bin"),
	"themes":     filepath.Join(BASE_DIR, "themes"),
	"git_clones": filepath.Join(CACHE_BASE_DIR, "git_clones"),
	"build":      filepath.Join(CACHE_BASE_DIR, "build"),
}

func GetDirPathFor(value string) string {
	home := user.GetHomeDir()

	return filepath.Join(home, PseudoPaths[value])
}
