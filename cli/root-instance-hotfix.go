package cli

import "github.com/spf13/cobra"

// hotfix applies hotfix on a given instance
// not implemented yet: hidden from help and from the interactive menu, kept so that existing calls do not fail
var hotfixInstanceRootCmd = &cobra.Command{
	Use:    "hotfix",
	Args:   cobra.NoArgs,
	Hidden: true,
	Short:  "Apply hotfix on an instance of " + nameProject + " (not implemented yet)",
	Run: func(cmd *cobra.Command, _ []string) {
		zboth.Info().Msgf("This functionality is yet to be implemented.")
	},
}

func init() {
	instanceRootCmd.AddCommand(hotfixInstanceRootCmd)
}
