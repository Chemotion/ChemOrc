package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var _advanced_pull_use_ string

var pullimageAdvancedRootCmd = &cobra.Command{
	Use:   "pull-image",
	Args:  cobra.NoArgs,
	Short: fmt.Sprintf("pull latest image of %s from the internet", nameCLI),
	Run: func(cmd *cobra.Command, args []string) {
		logWhere()
		confirmInstalled()
		tempCompose := getCompose(_advanced_pull_use_)
		services := getSubHeadings(&tempCompose, "services")
		for _, service := range services {
			zboth.Info().Msgf("Pulling image for the service called %s", service)
			if success := callVirtualizer(fmt.Sprintf("pull %s", tempCompose.GetString(joinKey("services", service, "image")))); !success {
				zboth.Warn().Err(fmt.Errorf("pull failed")).Msgf("Failed to pull image for the service called %s", service)
			}
		}
	},
}

func init() {
	pullimageAdvancedRootCmd.Flags().StringVar(&_advanced_pull_use_, "use", composeURL, "URL or filepath of the compose file to use when pulling the instance")
	advancedRootCmd.AddCommand(pullimageAdvancedRootCmd)
}
