package cli

import (
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

func dropIntoConsole(givenName string, consoleName string) {
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
	if status == "Up" {
		zboth.Info().Msgf("Entering %s for instance `%s`.", consoleName, givenName)
		if _, err, _ := gotoFolder(givenName), commandExec.Run(), gotoFolder("workdir"); err == nil {
			zboth.Debug().Msgf("Successfuly closed %s for `%s`.", consoleName, givenName)
		} else {
			zboth.Fatal().Err(err).Msgf("%s ended with exit message: %s.", consoleName, err.Error())
		}
	} else {
		zboth.Warn().Err(toError("instance is %s", status)).Msgf("Cannot start a %s for `%s`. Instance is not running.", consoleName, givenName)
	}
}

var shellConsoleInstanceRootCmd = &cobra.Command{
	Use:     "shell",
	Aliases: []string{"bash"},
	Short:   "Drop into a shell (bash) console",
	Args:    cobra.NoArgs,
	Run: func(_ *cobra.Command, _ []string) {
		dropIntoConsole(currentInstance, "shell")
	},
}

var railsConsoleInstanceRootCmd = &cobra.Command{
	Use:     "rails",
	Aliases: []string{"ruby", "railsc"},
	Short:   "Drop into a Ruby on Rails console",
	Args:    cobra.NoArgs,
	Run: func(_ *cobra.Command, _ []string) {
		dropIntoConsole(currentInstance, "railsc")
	},
}

var psqlConsoleInstanceRootCmd = &cobra.Command{
	Use:     "psql",
	Aliases: []string{"postgresql", "sql", "postgres", "PostgreSQL"},
	Short:   "Drop into a PostgreSQL console",
	Args:    cobra.NoArgs,
	Run: func(_ *cobra.Command, _ []string) {
		dropIntoConsole(currentInstance, "psql")
	},
}

func init() {
	consoleInstanceRootCmd.AddCommand(shellConsoleInstanceRootCmd)
	consoleInstanceRootCmd.AddCommand(railsConsoleInstanceRootCmd)
	consoleInstanceRootCmd.AddCommand(psqlConsoleInstanceRootCmd)
}
