package pkg

import (
	"errors"
	"fmt"
	"strings"

	"github.com/wizarki972/myone/internal/common"
	"github.com/wizarki972/myone/internal/utils/cmds"
	"github.com/wizarki972/myone/internal/utils/fldir"
)

const DEPENDENCIES = "go hyprland wireplumber blueman waybar rofi brightnessctl wiremix nwg-displays nwg-look nautilus wl-clipboard kitty swaync swayosd flameshot awww wlogout"

// Installs the given package names using pacman
func PkgInstall(pkg_name string) error {
	command := fmt.Sprintf("sudo pacman -Sy --needed --noconfirm %s", pkg_name)
	if !cmds.IsInteractiveShell() {
		cmds.ExecCommandInInInteractiveShell(fmt.Sprintf("%s\nFollowing dependencies are needed, \n\t%s\nDo you want to install it?", common.MYONE_ASCII, pkg_name), "MyOne-Dependency-Install", command, true, false)
	} else if _, err := cmds.ExecCommand(command, true, false); err != nil {
		return err
	}
	return nil
}

// Checks whether a certain package is installed or not.
func IsPkgInstalled(pkg_name string) bool {
	pkg_name = strings.TrimSpace(pkg_name)

	// check
	if len(pkg_name) == 0 {
		// slog.Warn("Enter a package name to check package's installation status")
		return false
	}

	// logic
	command := fmt.Sprintf("pacman -Q %s", pkg_name)
	out, err := cmds.ExecCommand(command, false, true)
	if err != nil {
		// slog.Warn(pkg_name + " package not installed")
		return false
	}

	return strings.HasPrefix(out, pkg_name)
}

// Installs linux packages for arch using pacman.
// It reads the package names from a file
func InstallPkgsFromFile(path string) error {
	if !fldir.IsPathExist(path) {
		return errors.New("dependency list file not found.")
	}

	content, err := fldir.ReadFileAsString(path)
	if err != nil {
		return err
	}

	var packages strings.Builder
	for line := range strings.SplitSeq(content, "\n") {
		if !IsPkgInstalled(line) {
			packages.WriteString(line + " ")
		}
	}

	if packages.Len() == 0 {
		// slog.Info("No new package(s) name found in the file.")
		return nil
	}

	return PkgInstall(packages.String())
}

// installs missing dependency
func Dependency_check() error {
	var packages strings.Builder
	for pkg := range strings.SplitSeq(DEPENDENCIES, " ") {
		if !IsPkgInstalled(pkg) {
			packages.WriteString(pkg + " ")
		}
	}

	if packages.Len() == 0 {
		// slog.Info("All dependencies are installed.")
		return nil
	}

	return PkgInstall(packages.String())
}
