package cli

import (
	"github.com/spf13/cobra"
)

var instanceRootCmd = &cobra.Command{
	Use:   "instance {status|switch|restart|new|remove}",
	Args:  cobra.NoArgs,
	Short: "Manipulate instances of " + nameCLI,
	Run: func(cmd *cobra.Command, args []string) {
		logWhere()
		confirmInstalled()
		confirmInteractive()
		acceptedOpts := []string{"status", "stats", "switch", "list", "restart", "new", "remove", "exit"} //, "status", "upgrade", "switch", "start", "pause", "stop", "restart", "delete"}
		switch selectOpt(acceptedOpts) {
		case "status":
			statusInstanceRootCmd.Run(cmd, args)
		case "stats":
			statInstanceRootCmd.Run(cmd, args)
		case "logs":
			logInstanceRootCmd.Run(cmd, args)
		case "switch":
			switchInstanceRootCmd.Run(cmd, args)
		case "list":
			listInstanceRootCmd.Run(cmd, args)
		case "restart":
			restartInstanceRootCmd.Run(cmd, args)
		case "new":
			newInstanceRootCmd.Run(cmd, args)
		case "remove":
			removeInstanceRootCmd.Run(cmd, args)
		case "exit":
			zlog.Debug().Msg("Chose to exit")
		}
	},
}

func init() {
	rootCmd.AddCommand(instanceRootCmd)
}
