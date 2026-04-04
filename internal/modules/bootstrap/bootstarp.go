package bootstrap

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/wizarki972/myone/internal/common"
	"github.com/wizarki972/myone/internal/utils/cmds"
)

func Dependency_install(pkg_name string) error {
	// check
	if len(pkg_name) == 0 {
		return errors.New("enter the package name to install")
	}

	// command
	command := fmt.Sprintf("sudo pacman -Sy --needed --noconfirm %s", pkg_name)
	if !cmds.Is_interactive_shell() {
		cmds.ExecForSudo(command)
	} else {
		cmds.ExecCommandInInteractiveShell(fmt.Sprintf("%s\nFollowing dependencies are needed, \n%s\nDo you want to install it?", common.MYONE_ASCII, pkg_name), "", "MyOne-Dependency-Install", command, true, false)
	}

	return nil
}

// rewrite this based on the themer dep check func
// func Dependecy_install_file(path string) {
// 	packages, err := fldir.ReadFileAsString(path)
// 	if err != nil {
// 		panic(err)
// 	}

// 	packages = strings.ReplaceAll(packages, "\n", " ")
// 	Dependency_install(packages)
// }

func Is_dependency_installed(pkg_name string) bool {
	pkg_name = strings.TrimSpace(pkg_name)

	// check
	if len(pkg_name) == 0 {
		slog.Warn("Enter the package name to check")
		return false
	}

	// logic
	command := fmt.Sprintf("pacman -Q %s", pkg_name)
	out, err := cmds.ExecCommand(command)
	if err != nil {
		// handle the errors later
		return false
	}

	return strings.HasPrefix(string(out), pkg_name)
}
