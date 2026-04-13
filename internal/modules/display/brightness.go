package display

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/wizarki972/myone/internal/utils/cmds"
)

// sends a signal to monitor daemon to change the brightness
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

// Changes brightness
func DefaultChangeBrightness(value string) {
	command := "brightnessctl s " + strings.TrimSpace(value)
	cmds.ExecCommandDetached(command)
}
