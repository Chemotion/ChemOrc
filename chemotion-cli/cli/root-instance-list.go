package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func instanceList() {
	if currentState.debug {
		zboth.Debug().Msgf("Currently existing instances are :", strings.Join(allInstances(), " "))
	}
	if !currentState.quiet {
		confirmInstalled()
		fmt.Printf("The following instances of %s exist:\n", nameCLI)
		for _, inst := range allInstances() {
			fmt.Println(inst)
		}
	}
}

var listInstanceRootCmd = &cobra.Command{
	Use:   "list",
	Args:  cobra.NoArgs,
	Short: "Get a list of all instances of " + nameCLI,
	Run: func(cmd *cobra.Command, args []string) {
		logWhere()
		instanceList()
	},
}

func init() {
	instanceRootCmd.AddCommand(listInstanceRootCmd)
}
