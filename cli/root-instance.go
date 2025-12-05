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
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		allIns := allInstances()
		numIns := len(allIns)
		if numIns == 1 && currentInstance != allIns[0] {
			zboth.Info().Msgf("Changing selected instance to %s as it is the only instance.", currentInstance)
			instanceSwitch(allIns[0])
		}
		if numIns == 0 && elementInSlice(cmd.Use, &[]string{"new", "restore"}) == -1 {
			getInternalName(currentInstance)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		isInteractive(true)
		var acceptedOpts []string
		allIns := allInstances()
		numIns := len(allIns)
		if numIns > 0 {
			instanceCmdTable["switch"] = switchInstanceRootCmd.Run
			if elementInSlice(instanceStatus(currentInstance), &[]string{"Exited", "Created"}) == -1 { // checks if the instance is running
				acceptedOpts = append(acceptedOpts, []string{"ping", "stats", "logs", "users", "consoles"}...)
				instanceCmdTable["ping"] = pingInstanceRootCmd.Run
				instanceCmdTable["stats"] = statInstanceRootCmd.Run
				instanceCmdTable["logs"] = logInstanceRootCmd.Run
				instanceCmdTable["users"] = userInstanceRootCmd.Run
				instanceCmdTable["consoles"] = consoleInstanceRootCmd.Run
			} else if instanceStatus(currentInstance) != "Created" {
				// confirm that it is not a brand new instance
				// hotfix should only be applied after the instance has been run at least once
				// logs are only available if the instance has been run at least once
				acceptedOpts = append(acceptedOpts, []string{"logs", "hotfix"}...)
				instanceCmdTable["hotfix"] = hotfixInstanceRootCmd.Run
				instanceCmdTable["logs"] = logInstanceRootCmd.Run
			}
			acceptedOpts = append(acceptedOpts, []string{"backup", "upgrade"}...)
			instanceCmdTable["backup"] = backupInstanceRootCmd.Run
			instanceCmdTable["upgrade"] = upgradeInstanceRootCmd.Run
			if numIns > 1 {
				acceptedOpts = append(acceptedOpts, "list")
				instanceCmdTable["list"] = listInstanceRootCmd.Run
			}
			acceptedOpts = append(acceptedOpts, []string{"new", "restore"}...)
			instanceCmdTable["new"] = newInstanceRootCmd.Run
			instanceCmdTable["restore"] = restoreInstanceRootCmd.Run
			if numIns > 1 {
				acceptedOpts = append(acceptedOpts, "remove")
				instanceCmdTable["remove"] = removeInstanceRootCmd.Run
			}
		} else {
			acceptedOpts = []string{"new", "restore"}
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
