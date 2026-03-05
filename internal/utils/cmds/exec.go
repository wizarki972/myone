package cmds

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"golang.org/x/term"
)

func ExecCommandNoFeedback(command string) {
	cmd := exec.Command("sh", "-c", command)
	_ = cmd.Run()
}

func ExecCommandWithOutput(command string) []byte {
	cmd := exec.Command("sh", "-c", command)
	out, err := cmd.CombinedOutput()
	if err != nil {
		panic(err)
	}
	return out
}

func ExecComamndWithError(command string) error {
	cmd := exec.Command("sh", "-c", command)
	_, err := cmd.CombinedOutput()
	return err
}

func ExecCommand(command string) ([]byte, error) {
	cmd := exec.Command("sh", "-c", command)
	output, err := cmd.CombinedOutput()
	return output, err
}

func ExecCommandInInteractiveShell(envs, title, command string, separate bool) {
	cmd := exec.Command("sh", "-c", fmt.Sprintf("%s kitty --title %s -e %s", envs, title, command))

	if separate {
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

func ExecForSudo(command string) error {
	cmd := exec.Command("sh", "-c", command)
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func Is_interactive_shell() bool {
	return term.IsTerminal(int(os.Stdin.Fd())) &&
		term.IsTerminal(int(os.Stdout.Fd()))
}

func Is_root() bool {
	return os.Geteuid() == 0
}

func Has_sudo() bool {
	cmd := exec.Command("sudo", "-n", "true")
	return cmd.Run() == nil
}
