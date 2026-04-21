package themer

import (
	"fmt"

	"github.com/wizarki972/myone/internal/utils/cmds"
	"github.com/wizarki972/myone/internal/utils/logger"
)

// refreshes waybar, swaync after theme change
func (t *Themer) refreshDesktop() {
	var processes = []string{"waybar", "swaync"}

	for _, name := range processes {
		command := fmt.Sprintf("pkill -x %s && %s", name, name)
		_, err := cmds.ExecCommand(command, false, false)
		if err != nil {
			t.logg_book.EnterLogAndPrint("failed while refreshing"+name+"this process.", logger.LogTypes.Error, err)
		}
	}
}
