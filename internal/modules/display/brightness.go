package display

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/wizarki972/myone/internal/utils/cmds"
)

func ChangeBrightness(value string) {
	sock := filepath.Join(os.Getenv("XDG_RUNTIME_DIR"), "myone-display-monitor.sock")
	conn, err := net.Dial("unix", sock)
	if err != nil {
		fmt.Println("Check whether the monitor daemon is running")
		panic(err)
	}
	defer conn.Close()

	conn.Write([]byte(value + "\n"))
}

func DefaultChangeBrightness(value string) error {
	command := "brightnessctl s " + strings.TrimSpace(value)
	return cmds.ExecComamndWithError(command)
}
