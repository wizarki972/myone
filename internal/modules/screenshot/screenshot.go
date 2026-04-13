package screenshot

import (
	"github.com/wizarki972/myone/internal/utils/cmds"
	"github.com/wizarki972/myone/internal/utils/fldir"
)

// A wrapper around flameshot to suit my needs
// when initial setup - set autostart to false
// also set useGrimAdapter=true in ~/.config/flameshot/flameshot.ini

// runs flameshot screenshot utility
func OpenGUI() {
	command := "flameshot gui --path " + fldir.GetXDGDir("PICTURES")
	cmds.ExecCommandDetached(command)
}
