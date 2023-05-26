package cli

import (
	"github.com/spf13/cobra"
)

var instanceCmdTable = make(cmdTable)

var instanceRootCmd = &cobra.Command{
	Use:     "instance",
	Aliases: []string{"instances"},
	Args:    cobra.NoArgs,
	Short:   "Manipulate instances of " + nameCLI,
	Run: func(cmd *cobra.Command, args []string) {
		isInteractive(true)
		var acceptedOpts []string
		if elementInSlice(instanceStatus(currentInstance), &[]string{"Exited", "Created"}) == -1 { // checks if the instance is running
			acceptedOpts = []string{"stats", "ping", "logs", "consoles", "users"}
			instanceCmdTable["stats"] = statInstanceRootCmd.Run
			instanceCmdTable["ping"] = pingInstanceRootCmd.Run
			instanceCmdTable["consoles"] = consoleInstanceRootCmd.Run
			instanceCmdTable["logs"] = logInstanceRootCmd.Run
			instanceCmdTable["users"] = userInstanceRootCmd.Run
		} else if instanceStatus(currentInstance) != "Created" {
			acceptedOpts = []string{"logs"}
			instanceCmdTable["logs"] = logInstanceRootCmd.Run
		}
		if len(allInstances()) > 1 {
			acceptedOpts = append(acceptedOpts, []string{"switch", "backup", "upgrade", "list", "new", "restore", "remove"}...)
			instanceCmdTable["switch"] = switchInstanceRootCmd.Run
			instanceCmdTable["backup"] = backupInstanceRootCmd.Run
			instanceCmdTable["upgrade"] = upgradeInstanceRootCmd.Run
			instanceCmdTable["list"] = listInstanceRootCmd.Run
			instanceCmdTable["remove"] = removeInstanceRootCmd.Run
			instanceCmdTable["new"] = newInstanceRootCmd.Run
			instanceCmdTable["restore"] = restoreInstanceRootCmd.Run
		} else {
			acceptedOpts = append(acceptedOpts, []string{"backup", "upgrade", "new", "restore"}...)
			instanceCmdTable["backup"] = backupInstanceRootCmd.Run
			instanceCmdTable["upgrade"] = upgradeInstanceRootCmd.Run
			instanceCmdTable["new"] = newInstanceRootCmd.Run
			instanceCmdTable["restore"] = restoreInstanceRootCmd.Run
		}
		if cmd.CalledAs() == "instance" {
			acceptedOpts = append(acceptedOpts, coloredExit)
		} else {
			acceptedOpts = append(acceptedOpts, []string{"back", coloredExit}...)
			instanceCmdTable["back"] = cmd.Run
		}
		instanceCmdTable[selectOpt(acceptedOpts, "")](cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(instanceRootCmd)
}
