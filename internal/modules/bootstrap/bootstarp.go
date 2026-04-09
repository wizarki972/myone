package bootstrap

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/wizarki972/myone/internal/common"
	"github.com/wizarki972/myone/internal/utils/cmds"
	"github.com/wizarki972/myone/internal/utils/fldir"
)

// All dependencies needed by this package alone, not by the themes
const DEPENDENCIES = "starship go hyprland wireplumber blueman waybar rofi brightnessctl wiremix nwg-displays nwg-look nautilus wl-clipboard kitty swaync swayosd flameshot awww wlogout zsh"

// Installs the given package names using pacman
func Pkg_install(pkg_name string) {
	command := fmt.Sprintf("sudo pacman -Sy --needed --noconfirm %s", pkg_name)
	if !cmds.Is_interactive_shell() {
		cmds.ExecCommandInInteractiveShell(fmt.Sprintf("%s\nFollowing dependencies are needed, \n\t%s\nDo you want to install it?", common.MYONE_ASCII, pkg_name), "", "MyOne-Dependency-Install", command, true, false)
	} else {
		cmds.ExecCommandNoFeedback(command)
	}
}

// Checks whether a certain package is installed or not.
func Is_pkg_installed(pkg_name string) bool {
	pkg_name = strings.TrimSpace(pkg_name)

	// check
	if len(pkg_name) == 0 {
		slog.Warn("Enter a package name to check package's installation status")
		return false
	}

	// logic
	command := fmt.Sprintf("pacman -Q %s", pkg_name)
	out, err := cmds.ExecCommand(command)
	if err != nil {
		slog.Warn(pkg_name + " package not installed")
		return false
	}

	return strings.HasPrefix(string(out), pkg_name)
}

// Installs linux packages for arch using pacman.
// It reads the package names from a file
func Install_pkgs_from_file(path string) {
	if !fldir.IsPathExist(path) {
		slog.Warn("Dependency list file not found.")
		return
	}

	content, err := fldir.ReadFileAsString(path)
	if err != nil {
		panic(err)
	}

	var packages strings.Builder
	for line := range strings.SplitSeq(content, "\n") {
		pkg := strings.TrimSpace(line)
		if !Is_pkg_installed(pkg) {
			packages.WriteString(pkg + " ")
		}
	}

	if len(strings.TrimSpace(packages.String())) == 0 {
		slog.Warn("No new package(s) name found in the file.")
	}

	Pkg_install(packages.String())
}

// installs missing dependency
func Dependency_check() {
	var packages strings.Builder
	for pkg := range strings.SplitSeq(DEPENDENCIES, " ") {
		if !Is_pkg_installed(pkg) {
			packages.WriteString(pkg + " ")
		}
	}

	if packages.Len() == 0 {
		slog.Info("All dependencies are installed.")
		return
	}

	Pkg_install(packages.String())
}
