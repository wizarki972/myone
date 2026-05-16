package release

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/wizarki972/myone/internal/common"
	"github.com/wizarki972/myone/internal/utils/cmds"
	"github.com/wizarki972/myone/internal/utils/fldir"
	"github.com/wizarki972/myone/internal/utils/logger"
)

const REPO_NAME = "myone"

func SelfUpdate(loggBook *logger.LogBook) {
	release, err := GetLatestRelease(REPO_NAME)
	if err != nil {
		loggBook.EnterLogAndPrint("Failed to get latest github release.", logger.LogTypes.Error, err)
	}

	isNewer, err := IsNewer(release, common.GetMajorVersion(), common.GetMinorVersion(), common.GetPatchVersion())
	if err != nil {
		loggBook.EnterLogAndPrint(err.Error(), logger.LogTypes.Error, err)
	}

	if isNewer {
		if !cmds.IsInteractiveShell() {
			fmt.Println("inter")
			if err := cmds.ExecCommandInInInteractiveShell("", "MyOne-Update", "myone --update", false, true); err != nil {
				loggBook.EnterLogAndPrint(err.Error(), logger.LogTypes.Error, err)
			}
		} else {
			fmt.Println("internon")
			fmt.Print(common.MYONE_ASCII)

			// getting user consent
			fmt.Printf("\nUpdate available %s ==> %s\n", common.GetVersionString(), strings.TrimPrefix(release.TagName, "v"))
			fmt.Print("Do you wish to update? [Y/n]: ")

			var response string
			fmt.Scanln(&response)
			response = strings.ToLower(response)
			if response == "" || response == "y" || response == "yes" {
				tempPath, err := DownloadLatestRelease(release)
				if err != nil {
					loggBook.EnterLogAndPrint(err.Error(), logger.LogTypes.Error, err)
				}
				tempDir := filepath.Dir(tempPath)
				if err := fldir.Unzip(tempPath, tempDir); err != nil {
					loggBook.EnterLogAndPrint(err.Error(), logger.LogTypes.Error, err)
				}

				// COMMAND
				cmd := exec.Command("bash", "-c", "make full_install")
				cmd.Dir = filepath.Join(tempDir, "myone-main")
				cmd.Stdout = os.Stdout
				cmd.Stdin = os.Stdin
				cmd.Stderr = os.Stderr

				if err := cmd.Run(); err != nil {
					// os.RemoveAll(tempDir)
					loggBook.EnterLogAndPrint("Error while running 'make full_install', in the downloaded source tree.", logger.LogTypes.Error, err)

				}

				// cleaning up cache
				os.RemoveAll(tempDir)

			} else {
				os.Exit(1)
			}
		}
	} else {
		fmt.Println("nei")
		loggBook.EnterLogAndPrint("Already on the latest version.", logger.LogTypes.Info, nil)
	}
}
