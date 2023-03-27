package cli

import (
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

func dropIntoConsole(givenName string, consoleName string) { // see if can be moved to callVirtualizer way of calling virtualizer
	commandExec := exec.Command(virtualizer, []string{"compose", "exec", "eln", "chemotion", consoleName}...)
	commandExec.Stdin, commandExec.Stdout, commandExec.Stderr = os.Stdin, os.Stdout, os.Stderr
	status := instanceStatus(givenName)
	switch consoleName { // use proper name when printing to user
	case "shell":
		consoleName = "shell"
	case "railsc":
		consoleName = "Rails console"
	case "psql":
		consoleName = "postgreSQL console"
	}
	if !(elementInSlice(status, &[]string{"Created", "Exited"}) > 0) {
		zboth.Info().Msgf("Entering %s for instance `%s`.", consoleName, givenName)
		if _, err, _ := gotoFolder(givenName), commandExec.Run(), gotoFolder("workdir"); err == nil {
			zboth.Debug().Msgf("Successfuly closed %s for `%s`.", consoleName, givenName)
		} else {
			zboth.Fatal().Err(err).Msgf("%s ended with exit message: %s.", consoleName, err.Error())
		}
	} else {
		zboth.Warn().Err(toError("instance is %s", status)).Msgf("Cannot start a %s for `%s`. Instance is %s.", consoleName, givenName, status)
	}
}

var consoleInstanceRootCmd = &cobra.Command{
	Use:       "console",
	Aliases:   []string{"consoles"},
	Short:     "Allow users to interact with an instance's command line interface",
	ValidArgs: []string{"shell", "railsc", "psql"},
	Run: func(cmd *cobra.Command, args []string) {
		var selected string
		if ownCall(cmd) && len(args) != 1 {
			if isInteractive(true) {
				acceptedOpts := []string{"shell", "ruby on rails", "postgreSQL"}
				if ownCall(cmd) {
					acceptedOpts = append(acceptedOpts, "exit")
				} else {
					acceptedOpts = append(acceptedOpts, []string{"back", "exit"}...)
				}
				selected = selectOpt(acceptedOpts, "")
			}
		} else {
			selected = args[0]
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
		case "exit":
			os.Exit(0)
		default:
			zboth.Info().Msgf("console expects one of the following: %.", strings.Join(cmd.ValidArgs, ", "))
		}

	},
}

func init() {
	instanceRootCmd.AddCommand(consoleInstanceRootCmd)
}
