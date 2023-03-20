package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func instanceList() {
	allInstances := allInstances()
	if conf.GetBool(joinKey(stateWord, "debug")) {
		zboth.Debug().Msgf("Currently existing instances are :", strings.Join(allInstances, " "))
	}
	if isInteractive(false) {
		fmt.Printf("The following instances of %s exist:\n", nameCLI)
		for _, inst := range allInstances {
			fmt.Println(inst)
		}
	}
}

var listInstanceRootCmd = &cobra.Command{
	Use:   "list",
	Args:  cobra.NoArgs,
	Short: "Get a list of all instances of " + nameCLI,
	Run: func(_ *cobra.Command, _ []string) {
		instanceList()
	},
}

func init() {
	instanceRootCmd.AddCommand(listInstanceRootCmd)
}
