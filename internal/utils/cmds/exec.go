package cmds

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"golang.org/x/term"
)

func ExecCommandContext(ctx context.Context, command string, feedback, output bool) (string, error) {
	cmd := exec.CommandContext(ctx, command)

	out, err := commonStringOutputLogic(cmd, feedback, output)
	if ctx.Err() != nil {
		return "", ctx.Err()
	}
	return out, err
}

func ExecCommandContextBytes(ctx context.Context, command string, feedback, output bool) ([]byte, error) {
	cmd := exec.CommandContext(ctx, command)
	out, err := commonBytesOutputLogic(cmd, feedback, output)
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	return out, err
}

func ExecCommand(command string, feedback, output bool) (string, error) {
	cmd := exec.Command("bash", "-c", command)

	return commonStringOutputLogic(cmd, feedback, output)
}

func ExecCommandBytes(command string, output bool) ([]byte, error) {
	cmd := exec.Command("bash", "-c", command)
	return commonBytesOutputLogic(cmd, false, output)
}

func commonStringOutputLogic(cmd *exec.Cmd, feedback, ouput bool) (string, error) {
	buf, err := commonExecCommandLogic(cmd, feedback, ouput)
	return strings.TrimSpace(buf.String()), err
}

func commonBytesOutputLogic(cmd *exec.Cmd, feedback, ouput bool) ([]byte, error) {
	buf, err := commonExecCommandLogic(cmd, feedback, ouput)
	return buf.Bytes(), err
}

func commonExecCommandLogic(cmd *exec.Cmd, feedback, output bool) (bytes.Buffer, error) {
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
		return buf, err
	}

	return buf, nil
}

func ExecCommandDetached(command string) error {
	cmd := exec.Command("bash", "-c", command)

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}
	cmd.Stderr = nil
	cmd.Stdin = nil
	cmd.Stdout = nil

	if err := cmd.Start(); err != nil {
		return err
	}
	return nil
}

func ExecCommandInInInteractiveShell(msg, title, command string, ask_user_permission, detach bool) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return errors.New("cannot get user home directory")
	}
	kittyConfig := filepath.Join(home, ".config/kitty/kitty_popup.conf")
	var cmd *exec.Cmd
	if ask_user_permission {
		script := fmt.Sprintf(`printf '%s [y/N]: '; read ans; if [[ "$ans" =~ ^[yY]$ ]]; then %s; fi; printf '\nPress any key to exit...'; read`, msg, command)
		cmd = exec.Command("kitty", "-c", kittyConfig, "--title", title, "-e", "bash", "-c", script)
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
			return err
		}
	} else if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func IsInteractiveShell() bool {
	return term.IsTerminal(int(os.Stdin.Fd())) &&
		term.IsTerminal(int(os.Stdout.Fd()))
}
