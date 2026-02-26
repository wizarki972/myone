package updater

import (
	"log/slog"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/wizarki972/myone/internal/utils/config"
	"github.com/wizarki972/myone/internal/utils/fldir"
)

const VERSION_URL = "https://raw.githubusercontent.com/wizarki972/myone/main/VERSION"
const DOWNLOAD_URL = "https://github.com/wizarki972/myone/archive/refs/heads/main.zip"

func get_latest() float64 {
	verStr := fldir.ReadTextFileFromURL(VERSION_URL, false, "")
	verStr = strings.SplitN(strings.Split(verStr, "-")[0], ".", 2)[1]
	out, err := strconv.ParseFloat(verStr, 64)
	if err != nil {
		panic(err)
	}
	return out
}

func is_latest() bool {
	return get_latest() == config.VERSION_INT
}

// Needs a second look
func Update() {
	if !is_latest() {
		cache_dir := config.GetDirPathFor("build")
		cache_path := filepath.Join(cache_dir, "repo.zip")

		slog.Info("Downloading Files...")
		fldir.DownloadURL(DOWNLOAD_URL, cache_path)
		fldir.Unzip(cache_path, cache_dir)

		// COMMAND
		cmd := exec.Command("sh", "-c", "make install")
		cmd.Dir = filepath.Join(cache_dir, "myone-main")
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setsid: true,
		}
		cmd.Stdout = nil
		cmd.Stdin = nil
		cmd.Stderr = nil

		if err := cmd.Run(); err != nil {
			panic(err)
		}
	}
}
