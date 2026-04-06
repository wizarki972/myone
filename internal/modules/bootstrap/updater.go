package bootstrap

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/wizarki972/myone/internal/common"
	"github.com/wizarki972/myone/internal/utils/cmds"
	"github.com/wizarki972/myone/internal/utils/fldir"
)

const VERSION_URL = "https://raw.githubusercontent.com/wizarki972/myone/main/VERSION"
const DOWNLOAD_URL = "https://github.com/wizarki972/myone/archive/refs/heads/main.zip"

// NOTE
// MYONE_INTERNAL environment variable is used to separate a background update check from a command executed by the user

func is_latest() (bool, string) {
	// Getting latest version from repo
	ver_str := fldir.ReadTextFileFromURL(VERSION_URL, false, "")
	out, err := strconv.ParseFloat(strings.SplitN(ver_str, ".", 2)[1], 64)
	if err != nil {
		panic(err)
	}
	return out == common.GetVersionFloat(), ver_str
}

func Self_update() {
	ok, latest := is_latest()

	// If the version in repo is not the one installed then it will perform update/downgrade
	// This allows downgrading to last stable in case of bugs by rolling back to older releases in repo.
	if !ok {
		if !cmds.Is_interactive_shell() {
			cmds.ExecCommandInInteractiveShell("", "", "MyOne-Update", "myone --update", false, true)
		} else {
			fmt.Print(common.MYONE_ASCII)

			// getting user consent
			fmt.Printf("Update available %s ==> %s\n", common.VERSION, latest)
			fmt.Print("Do you wish to update? [Y/n]: ")

			var response string
			fmt.Scanln(&response)
			response = strings.ToLower(response)
			if response == "" || response == "y" || response == "yes" {
				// Cache directory paths
				cache_dir := filepath.Join(fldir.GetHomeDir(), common.CACHE_DIR, "update")
				cache_path := filepath.Join(cache_dir, "repo.zip")

				fmt.Println("DOWNLOADING...")
				fldir.DownloadURL(DOWNLOAD_URL, cache_path, false)
				fldir.Unzip(cache_path, cache_dir)

				// COMMAND
				cmd := exec.Command("sh", "-c", "make full_install")
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
