package screenshot

import (
	"github.com/wizarki972/myone/internal/utils/cmds"
	"github.com/wizarki972/myone/internal/utils/user"
)

// A wrapper around flameshot to suit my needs
// when initial setup - set autostart to false
// also set useGrimAdapter=true in ~/.config/flameshot/flameshot.ini

func OpenGUI() {
	command := "flameshot gui --path " + user.GetXDGDir("PICTURES")
	if err := cmds.ExecComamndWithError(command); err != nil {
		panic(err)
	}
}
