package themer

import (
	"fmt"
	"log/slog"

	"github.com/wizarki972/myone/internal/utils/cmds"
)

// refreshes waybar, swaync after theme change
func (t *Themer) refreshDesktop() {
	var processes = []string{"waybar", "swaync"}

	for _, name := range processes {
		command := fmt.Sprintf("pkill -x %s && %s", name, name)
		_, err := cmds.ExecCommand(command, false, false)
		if err != nil {
			slog.Error(fmt.Sprintf("failed while refreshing %s this process.", name))
			panic(err)
		}
	}
}
