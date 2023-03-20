package cli

import (
	"github.com/spf13/cobra"
)

func instanceStart(givenName string) {
	status := instanceStatus(givenName)
	if status == "Up" {
		zboth.Warn().Msgf("The instance called %s is already running.", givenName)
	} else {
		if _, success, _ := gotoFolder(givenName), callVirtualizer("compose up -d"), gotoFolder("workdir"); success {
			var (
				seconds int
				ft      string
			)
			if seconds = 20; status == "Created" {
				seconds = 60
			}
			if firstRun {
				ft = " for the first time"
			}
			zboth.Info().Msgf("Starting instance called %s. Please give it %d seconds to initialize%s.", givenName, seconds, ft)
			waitProgressBar(seconds, []string{"Starting", givenName})
			zboth.Info().Msgf("Successfully started instance called %s.", givenName)
		} else {
			zboth.Fatal().Msgf("Failed to start instance called %s.", givenName)
		}
	}
}

func instanceStop(givenName string) {
	status := instanceStatus(givenName)
	if status == "Up" {
		if _, success, _ := gotoFolder(givenName), callVirtualizer("compose stop"), gotoFolder("workdir"); success {
			zboth.Info().Msgf("Successfully stopped instance called %s.", givenName)
		} else {
			zboth.Fatal().Msgf("Failed to stop instance called %s.", givenName)
		}
	} else {
		zboth.Warn().Msgf("Cannot stop instance %s. It seems to be %s.", givenName, status)
	}
}

var onRootCmd = &cobra.Command{
	Use:   "on",
	Args:  cobra.NoArgs,
	Short: "Start (the selected instance of) chemotion",
	Run: func(cmd *cobra.Command, args []string) {
		logWhere()
		confirmInstalled()
		instanceStart(currentState.name)
	},
}

var offRootCmd = &cobra.Command{
	Use:   "off",
	Args:  cobra.NoArgs,
	Short: "Stop (the selected instance of) chemotion",
	Run: func(cmd *cobra.Command, args []string) {
		logWhere()
		confirmInstalled()
		instanceStop(currentState.name)
	},
}

func init() {
	rootCmd.AddCommand(onRootCmd)
	rootCmd.AddCommand(offRootCmd)
}
