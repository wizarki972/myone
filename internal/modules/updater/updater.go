package updater

import (
	"fmt"
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

func Update() {

	// Getting latest version from repo
	verStr := fldir.ReadTextFileFromURL(VERSION_URL, false, "")
	verStr = strings.SplitN(strings.Split(verStr, "-")[0], ".", 2)[1]
	out, err := strconv.ParseFloat(verStr, 64)
	if err != nil {
		panic(err)
	}

	// If the version in repo is not the one installed then it will perform update/downgrade
	// This allows downgrading to last stable in case of bugs by rolling back to older releases in repo.
	if !(out == config.VERSION_INT) {
		// Cache directory paths
		cache_dir := filepath.Join(config.CACHE_BASE_DIR, "update")
		cache_path := filepath.Join(cache_dir, "repo.zip")

		// Need to change this line
		slog.Info("Downloading Files...")
		fldir.DownloadURL(DOWNLOAD_URL, cache_path, false)
		fldir.Unzip(cache_path, cache_dir)

		// COMMAND
		cmd := exec.Command("sh", "-c", "make install_pkexec")
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
	} else {
		fmt.Println("Already on the latest build.")
	}
}
