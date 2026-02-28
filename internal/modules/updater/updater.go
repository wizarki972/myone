package updater

import (
	"fmt"
	"os"
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

// NOTE
// MYONE_INTERNAL environment variable is used to separate a background update check from a command executed by the user

func isLatest() (bool, string) {
	// Getting latest version from repo
	ver_str := fldir.ReadTextFileFromURL(VERSION_URL, false, "")
	out, err := strconv.ParseFloat(strings.SplitN(strings.Split(ver_str, "-")[0], ".", 2)[1], 64)
	if err != nil {
		panic(err)
	}
	return out == config.VERSION_INT, ver_str
}

func Update(gui bool) {
	ok, latest := isLatest()

	// If the version in repo is not the one installed then it will perform update/downgrade
	// This allows downgrading to last stable in case of bugs by rolling back to older releases in repo.
	if !ok {
		if gui {
			cmd := exec.Command("sh", "-c", "MYONE_INTERNAL=0 kitty --title MyOne-Update -e myone --update")
			cmd.SysProcAttr = &syscall.SysProcAttr{
				Setsid: true,
			}
			cmd.Stderr = nil
			cmd.Stdout = nil
			cmd.Stdin = nil

			if err := cmd.Run(); err != nil {
				panic(err)
			}
		} else {
			fmt.Print(config.MYONE_ASCII)

			// getting user consent
			fmt.Printf("Update available %s ==> 0.%s\n", config.VERSION, latest)
			fmt.Print("Do you wish to update? [Y/n]: ")

			var response string
			fmt.Scanln(&response)
			response = strings.ToLower(response)
			if response == "" || response == "y" || response == "yes" {
				// Cache directory paths
				cache_dir := filepath.Join(config.CACHE_BASE_DIR, "update")
				cache_path := filepath.Join(cache_dir, "repo.zip")

				fmt.Println("DOWNLOADING...")
				fldir.DownloadURL(DOWNLOAD_URL, cache_path, false)
				fldir.Unzip(cache_path, cache_dir)

				// COMMAND
				cmd := exec.Command("sh", "-c", "make install")
				cmd.Dir = filepath.Join(cache_dir, "myone-main")
				cmd.Stdout = os.Stdout
				cmd.Stdin = os.Stdin
				cmd.Stderr = os.Stderr

				if err := cmd.Run(); err != nil {
					os.RemoveAll(cache_dir)
					panic(err)
				}

				// cleaning up cache
				os.RemoveAll(cache_dir)
			}
		}
	} else {
		fmt.Println("Already on the latest build.")
	}
}
