package cmd

import (
	"errors"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wizarki972/myone/internal/services"
	"github.com/wizarki972/myone/internal/utils/cmds"
	"github.com/wizarki972/myone/internal/utils/logger"
)

var battMon, displayMon, daemon bool

var servicesCMD = &cobra.Command{
	Use:   "services",
	Short: "start services from here...",
	RunE: func(cmd *cobra.Command, args []string) error {
		loggerInstance = handleLogg()
		loggerInstance.AddSubCommand("services")

		if battMon && displayMon {
			loggerInstance.EnterLogAndPrint("Cannot run both battery-monitor and display-monitor flag in a single command.", logger.LogTypes.Error, errors.New("cannot run both battery-monitor and display-monitor flag in a single command"))
		}

		if daemon {
			childArgs := []string{}
			childArgs = append(childArgs, os.Args[0])
			for _, arg := range os.Args[1:] {
				if arg != "-d" && arg != "--daemon" && arg != "-daemon" {
					childArgs = append(childArgs, arg)
				}
			}
			command := strings.Join(childArgs, " ")
			if err := cmds.ExecCommandDetached(command); err != nil {
				loggerInstance.EnterLogAndPrint(err.Error(), logger.LogTypes.Error, err)
			}
		} else {
			if battMon {
				loggerInstance.AddFlag("battery-monitor")
				bm := services.NewBattMon(loggerInstance, userConfig)
				bm.StartService()
			}
		}

		return nil
	},
}

func initializeServicesFlags() {
	servicesCMD.Flags().BoolVar(&battMon, "battery-monitor", false, "runs battery monitor service.")

	servicesCMD.Flags().BoolVarP(&daemon, "daemon", "d", false, "runs the process in the background as a daemon.")

	servicesCMD.Flags().BoolVar(&saveLog, "save-log", false, "saves the log based on the default path or path specified in config.\nNo need to use this flag, if you are using --log-path flag.")

	servicesCMD.Flags().StringVar(&logPath, "log-path", "", "saves the log to the provided path.")

	// rootCMD.AddCommand(servicesCMD)
}
