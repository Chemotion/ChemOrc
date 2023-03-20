package cli

import (
	"github.com/spf13/cobra"
	"golang.org/x/exp/maps"
)

var advancedCmdTable = make(cmdTable)

// Backbone for system-related commands
var advancedRootCmd = &cobra.Command{
	Use:       "advanced",
	Short:     "Perform advanced actions related to system and " + nameCLI,
	Args:      cobra.NoArgs,
	ValidArgs: maps.Keys(advancedCmdTable),
	PersistentPreRun: func(cmd *cobra.Command, _ []string) {
		if cmd.Flag("selected-instance").Changed {
			zboth.Warn().Msgf("The `-i` flag is not supported for the `advanced` command and its subcommands.")
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		isInteractive(true)
		acceptedOpts := []string{"info"}
		advancedCmdTable["info"] = infoAdvancedRootCmd.Run
		if updateRequired() {
			acceptedOpts = append(acceptedOpts, "update cli")
			advancedCmdTable["update cli"] = updateSelfAdvancedRootCmd.Run
		}
		if cmd.Use == cmd.CalledAs() { // || elementInSlice(cmd.CalledAs(), &cmd.Aliases) > -1 { { // there are no aliases at the moment
			acceptedOpts = append(acceptedOpts, []string{"uninstall", "exit"}...)
			advancedCmdTable["uninstall"] = uninstallAdvancedRootCmd.Run
		} else {
			acceptedOpts = append(acceptedOpts, []string{"uninstall", "back", "exit"}...)
			advancedCmdTable["uninstall"] = uninstallAdvancedRootCmd.Run
			advancedCmdTable["back"] = cmd.Run
		}
		advancedCmdTable[selectOpt(acceptedOpts, "")](cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(advancedRootCmd)
}
