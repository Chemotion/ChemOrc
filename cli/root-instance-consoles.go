package cli

import (
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func dropIntoConsole(givenName string, consoleName string) { // see if can be moved to callVirtualizer way of calling virtualizer
	status, properName := instanceStatus(givenName), consoleName
	switch consoleName { // use proper name when printing to user
	case "shell":
		properName = "shell"
	case "railsc":
		properName = "Rails console"
	case "psql":
		properName = "postgreSQL console"
	}
	if !(elementInSlice(status, &[]string{"Created", "Exited"}) > 0) {
		zboth.Info().Msgf("Entering %s for instance `%s`.", properName, givenName)
		if _, success, _ := gotoFolder(givenName), callVirtualizer(toSprintf("compose exec %s chemotion %s", primaryService, consoleName)), gotoFolder("work.dir"); success {
			zboth.Debug().Msgf("Successfuly closed %s for `%s`.", properName, givenName)
		} else {
			zboth.Warn().Msgf("%s ended with an error.", properName)
		}
	} else {
		zboth.Warn().Err(toError("instance is %s", status)).Msgf("Cannot start a %s for `%s`. Instance is %s.", properName, givenName, status)
	}
}

var consoleInstanceRootCmd = &cobra.Command{
	Use:       "console",
	Aliases:   []string{"consoles"},
	Short:     "Allow users to interact with an instance's command line interface",
	ValidArgs: []string{"shell", "railsc", "psql"},
	Run: func(cmd *cobra.Command, args []string) {
		var selected string
		if ownCall(cmd) {
			if selected = "multiple arguments"; len(args) == 1 {
				selected = args[0]
			}
		} else {
			if isInteractive(true) {
				acceptedOpts := []string{"shell", "ruby on rails", "postgreSQL"}
				if ownCall(cmd) {
					acceptedOpts = append(acceptedOpts, coloredExit)
				} else {
					acceptedOpts = append(acceptedOpts, []string{"back", coloredExit}...)
				}
				selected = selectOpt(acceptedOpts, "")
			}
		}
		switch selected {
		case "shell", "bash", "sh":
			dropIntoConsole(currentInstance, "shell")
		case "ruby on rails", "railsc", "ruby":
			dropIntoConsole(currentInstance, "railsc")
		case "postgreSQL", "psql", "postgres", "postgresql":
			dropIntoConsole(currentInstance, "psql")
		case "back":
			cmd.Run(cmd, args)
		case coloredExit:
			os.Exit(0)
		case "multiple arguments":
			zboth.Warn().Msgf("console expects only ONE argument of the following: %s.", strings.Join(cmd.ValidArgs, ", "))
		default:
			zboth.Info().Msgf("console expects one of the following: %s.", strings.Join(cmd.ValidArgs, ", "))
		}
	},
}

func init() {
	instanceRootCmd.AddCommand(consoleInstanceRootCmd)
}
