package user

import "github.com/wizarki972/myone/internal/utils/cmds"

func GetXDGDir(name string) string {
	output, err := cmds.ExecCommand("xdg-user-dir " + name)
	if err != nil {
		panic(err)
	}
	return string(output)
}
