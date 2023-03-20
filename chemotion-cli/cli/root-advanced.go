package cli

import (
	"github.com/spf13/cobra"
)

// Backbone for system-related commands
var advancedRootCmd = &cobra.Command{
	Use:   "advanced {info|uninstall}",
	Short: "Perform advanced actions related to system and " + nameCLI,
	Run: func(cmd *cobra.Command, args []string) {
		logWhere()
		confirmInstalled()
		confirmInteractive()
		acceptedOpts := []string{"info", "uninstall", "exit"}
		selected := selectOpt(acceptedOpts)
		switch selected {
		case "info":
			infoAdvancedRootCmd.Run(cmd, args)
		case "uninstall":
			uninstallAdvancedRootCmd.Run(cmd, args)
		case "exit":
			zlog.Debug().Msg("Chose to exit")
		}
	},
}

func init() {
	rootCmd.AddCommand(advancedRootCmd)
}
