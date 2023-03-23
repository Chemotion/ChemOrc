package cli

import (
	"github.com/spf13/cobra"
	"golang.org/x/exp/maps"
)

var consoleInstanceCmdTable = make(cmdTable)

var consoleInstanceRootCmd = &cobra.Command{
	Use:       "console",
	Aliases:   []string{"consoles"},
	Short:     "Allow users to interact with an instance's command line interface",
	ValidArgs: maps.Keys(consoleInstanceCmdTable),
	Run: func(cmd *cobra.Command, args []string) {
		isInteractive(true)
		acceptedOpts := []string{"shell", "ruby on rails", "postgresSQL"}
		consoleInstanceCmdTable["shell"] = shellConsoleInstanceRootCmd.Run
		consoleInstanceCmdTable["ruby on rails"] = railsConsoleInstanceRootCmd.Run
		consoleInstanceCmdTable["postgreSQL"] = psqlConsoleInstanceRootCmd.Run
		if ownCall(cmd) {
			acceptedOpts = append(acceptedOpts, "exit")
		} else {
			acceptedOpts = append(acceptedOpts, []string{"back", "exit"}...)
			consoleInstanceCmdTable["back"] = cmd.Run // FIXME: takes to the root menu instead of the parent menu
		}
		consoleInstanceCmdTable[selectOpt(acceptedOpts, "")](cmd, args)
	},
}

func init() {
	instanceRootCmd.AddCommand(consoleInstanceRootCmd)
}
