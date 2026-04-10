package cmds

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"

	"golang.org/x/term"
)

func Exec_cmd(command string, feedback, output, detach bool) (string, error) {
	cmd := exec.Command("sh", "-c", command)

	var buf bytes.Buffer
	if detach {
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setsid: true,
		}
		cmd.Stderr = nil
		cmd.Stdin = nil
		cmd.Stdout = nil
	} else {
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
	}

	if err := cmd.Run(); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func Exec_cmd_bytes(command string, output bool) ([]byte, error) {
	var buf bytes.Buffer

	cmd := exec.Command("sh", "-c", command)
	if output {
		cmd.Stdout = &buf
	}
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func ExecCommandInInteractiveShell(msg, envs, title, command string, ask_permission, detach bool) {
	var cmd *exec.Cmd
	if ask_permission {
		cmd = exec.Command("bash", "-c", fmt.Sprintf("printf '%s [y/N]: ' && read ans && [[ '$ans' =~ ^[Yy]$ ]] && %s kitty --title %s -e sh -c \"%s && printf 'Press any key to exit...' && read\"", msg, envs, title, command))
	} else {
		cmd = exec.Command("bash", "-c", fmt.Sprintf("%s kitty --title %s -e sh -c \"%s && printf 'Press any key to exit...' && read\"", envs, title, command))
	}

	if detach {
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setsid: true,
		}
		cmd.Stderr = nil
		cmd.Stdout = nil
		cmd.Stdin = nil
	}

	if err := cmd.Run(); err != nil {
		panic(err)
	}
}

func Is_interactive_shell() bool {
	return term.IsTerminal(int(os.Stdin.Fd())) &&
		term.IsTerminal(int(os.Stdout.Fd()))
}
