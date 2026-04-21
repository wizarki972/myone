package cmds

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"golang.org/x/term"
)

func ExecCommand(command string, feedback, output bool) (string, error) {
	cmd := exec.Command("bash", "-c", command)

	var buf bytes.Buffer
	switch {
	case feedback && output:
		multi := io.MultiWriter(os.Stdout, &buf)
		cmd.Stdout = multi
		cmd.Stderr = os.Stderr
	case feedback:
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	case output:
		cmd.Stdout = &buf
	}
	if err := cmd.Run(); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func ExecCommandDetached(command string) {
	cmd := exec.Command("bash", "-c", command)

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}
	cmd.Stderr = nil
	cmd.Stdin = nil
	cmd.Stdout = nil

	if err := cmd.Start(); err != nil {
		panic(err)
	}

}

func ExecCommandBytes(command string, output bool) ([]byte, error) {
	var buf bytes.Buffer

	cmd := exec.Command("bash", "-c", command)
	if output {
		cmd.Stdout = &buf
	}
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func ExecCommandInInInteractiveShell(msg, title, command string, ask_user_permission, detach bool) {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("==> [ERROR] Cannot get user home directory.")
		os.Exit(1)
	}
	kittyConfig := filepath.Join(home, ".config/kitty/kitty_popup.conf")
	var cmd *exec.Cmd
	if ask_user_permission {
		script := fmt.Sprintf(`printf '%s [y/N]: '; read ans; if [[ "$ans" =~ ^[yY]$ ]]; then %s; fi; printf '\nPress any key to exit...'; read`, msg, command)
		cmd = exec.Command("kitty", "--hold", "-c", kittyConfig, "--title", title, "-e", "bash", "-c", script)
	} else {
		script := fmt.Sprintf("%s && printf 'Press any key to exit...' && read", command)
		cmd = exec.Command("kitty", "-c", kittyConfig, "--title", title, "-e", "bash", "-c", script)
	}

	if detach {
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setsid: true,
		}
		cmd.Stderr = nil
		cmd.Stdout = nil
		cmd.Stdin = nil

		if err := cmd.Start(); err != nil {
			panic(err)
		}
	} else if err := cmd.Run(); err != nil {
		panic(err)
	}
}

func IsInteractiveShell() bool {
	return term.IsTerminal(int(os.Stdin.Fd())) &&
		term.IsTerminal(int(os.Stdout.Fd()))
}
