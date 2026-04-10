package themer

import (
	"fmt"
	"log/slog"

	"github.com/wizarki972/myone/internal/utils/cmds"
)

func (t *Themer) refresh_desktop() {
	var processes = []string{"waybar", "swaync"}

	for _, name := range processes {
		command := fmt.Sprintf("pkill -x %s && %s", name, name)
		_, err := cmds.Exec_cmd(command, false, false, true)
		if err != nil {
			slog.Error(fmt.Sprintf("failed while refreshing %s this process.", name))
			panic(err)
		}
	}
}
