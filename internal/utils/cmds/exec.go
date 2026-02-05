package cmds

import "os/exec"

func ExecCommandNoOut(command string) {
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
