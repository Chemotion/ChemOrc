package cli

import (
	"github.com/spf13/cobra"
	"golang.org/x/exp/maps"
)

var instanceCmdTable = make(cmdTable)

var instanceRootCmd = &cobra.Command{
	Use:       "instance",
	ValidArgs: maps.Keys(instanceCmdTable),
	Args:      cobra.NoArgs,
	Short:     "Manipulate instances of " + nameCLI,
	Run: func(cmd *cobra.Command, args []string) {
		isInteractive(true)
		var acceptedOpts []string
		if elementInSlice(instanceStatus(currentInstance), &[]string{"Exited", "Created"}) == -1 { // checks if the instance is running
			acceptedOpts = []string{"stats", "ping", "logs", "consoles"}
			instanceCmdTable["stats"] = statInstanceRootCmd.Run
			instanceCmdTable["ping"] = pingInstanceRootCmd.Run
			instanceCmdTable["consoles"] = consoleInstanceRootCmd.Run
			instanceCmdTable["logs"] = logInstanceRootCmd.Run
		} else if instanceStatus(currentInstance) != "Created" {
			acceptedOpts = []string{"logs"}
			instanceCmdTable["logs"] = logInstanceRootCmd.Run
		}
		if len(allInstances()) > 1 {
			acceptedOpts = append(acceptedOpts, []string{"switch", "backup", "upgrade", "list", "new", "remove"}...)
			instanceCmdTable["switch"] = switchInstanceRootCmd.Run
			instanceCmdTable["backup"] = backupInstanceRootCmd.Run
			instanceCmdTable["upgrade"] = upgradeInstanceRootCmd.Run
			instanceCmdTable["list"] = listInstanceRootCmd.Run
			instanceCmdTable["remove"] = removeInstanceRootCmd.Run
			instanceCmdTable["new"] = newInstanceRootCmd.Run
		} else {
			acceptedOpts = append(acceptedOpts, []string{"backup", "upgrade", "new"}...)
			instanceCmdTable["backup"] = backupInstanceRootCmd.Run
			instanceCmdTable["upgrade"] = upgradeInstanceRootCmd.Run
			instanceCmdTable["new"] = newInstanceRootCmd.Run
		}
		if cmd.CalledAs() == "instance" {
			acceptedOpts = append(acceptedOpts, "exit")
		} else {
			acceptedOpts = append(acceptedOpts, []string{"back", "exit"}...)
			instanceCmdTable["back"] = cmd.Run
		}
		instanceCmdTable[selectOpt(acceptedOpts, "")](cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(instanceRootCmd)
}
