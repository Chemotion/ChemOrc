package cli

import "github.com/spf13/cobra"

// hotfix applies hotfix on a given instance
var hotfixInstanceRootCmd = &cobra.Command{
	Use:   "hotfix",
	Args:  cobra.NoArgs,
	Short: "Apply hotfix on an instance of " + nameProject,
	Run: func(cmd *cobra.Command, _ []string) {
		zboth.Info().Msgf("This functionality is yet to be implemented.")
	},
}

func init() {
	instanceRootCmd.AddCommand(hotfixInstanceRootCmd)
}
