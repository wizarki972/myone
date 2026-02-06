package cmds

import "os/exec"

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
