package logout

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/wizarki972/myone/internal/modules/display"
	"github.com/wizarki972/myone/internal/utils/cmds"
	"github.com/wizarki972/myone/internal/utils/fldir"
	"github.com/wizarki972/myone/internal/utils/user"
)

func GetLogoutValues(layout int) (map[string]string, error) {
	var width, height, hyprBorder, scale int

	width, height, scale = display.GetScreenresolution()
	scale = scale * 100
	hyprBorder = display.GetHyprBorder()
	fontSize := fmt.Sprintf("%d", height*2/100)
	buttonRadius := fmt.Sprintf("%d", hyprBorder*8)
	activeButtonRadius := fmt.Sprintf("%d", hyprBorder*5)

	logoutValues := map[string]string{
		"${fontSize}":             fontSize,
		"${button_radius}":        buttonRadius,
		"${active_button_radius}": activeButtonRadius,
		"${HOME}":                 user.GetHomeDir(),
	}

	switch layout {
	case 1:
		margin := fmt.Sprintf("%d", height*28/scale)
		hover := fmt.Sprintf("%d", height*23/scale)
		logoutValues["${margin}"] = margin
		logoutValues["${hover}"] = hover
	case 2:
		x_margin := fmt.Sprintf("%d", width*35/scale)
		y_margin := fmt.Sprintf("%d", height*25/scale)
		x_hover := fmt.Sprintf("%d", width*32/scale)
		y_hover := fmt.Sprintf("%d", height*20/scale)

		logoutValues["${x_margin}"] = x_margin
		logoutValues["${y_margin}"] = y_margin
		logoutValues["${x_hover}"] = x_hover
		logoutValues["${y_hover}"] = y_hover
	}

	return logoutValues, nil
}

func Logout(layout int) {
	var logoutValues map[string]string
	var cols int
	var err error

	// this command is little different from other error checks
	if err = cmds.ExecComamndWithError("pkill wlogout"); err == nil {
		return
	}

	home := user.GetHomeDir()
	layoutPath := fmt.Sprintf("%s/.config/wlogout/layout_%d", home, layout)
	stylesPath := fmt.Sprintf("%s/.config/wlogout/style_%d.css", home, layout)

	stylesContent, err := fldir.ReadFileAsString(stylesPath)
	if err != nil {
		panic(err)
	}

	if logoutValues, err = GetLogoutValues(layout); err != nil {
		panic(err)
	}

	for old, new := range logoutValues {
		stylesContent = strings.ReplaceAll(stylesContent, old, new)
	}

	switch layout {
	case 1:
		cols = 6
	case 2:
		cols = 2
	}

	cssPath := filepath.Join(home, ".cache/myone/logout/style.css")
	fldir.WriteFile(stylesContent, cssPath)

	command := fmt.Sprintf("wlogout -b %d -c 0 -r 0 -m 0 --layout %s --css %s --protocol layer-shell", cols, layoutPath, cssPath)
	if err = cmds.ExecComamndWithError(command); err != nil {
		panic(err)
	}
}
