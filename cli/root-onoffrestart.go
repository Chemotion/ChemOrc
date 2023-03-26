package cli

import (
	"fmt"
	"time"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

// show (and then remove) a progress bar that waits for an instance to start
func waitStartSpinner(seconds int, givenName string) (waitTime int) {
	bar := progressbar.NewOptions(
		-1,
		progressbar.OptionSetDescription(toSprintf("Starting %s...", givenName)),
		progressbar.OptionSetPredictTime(false),
		progressbar.OptionClearOnFinish(),
		progressbar.OptionSetRenderBlankState(true),
		progressbar.OptionSetVisibility(true),
		progressbar.OptionSpinnerType(51),
	)
	for i := 0; i < seconds; i++ {
		if instancePing(givenName) == "200 OK" {
			bar.Finish()
			fmt.Println()
			waitTime = i
			return
		}
		bar.Add(1)
		time.Sleep(1 * time.Second)
	}
	bar.Finish()
	waitTime = -1
	fmt.Println()
	return
}

func instanceStart(givenName string) {
	status := instanceStatus(givenName)
	if status == "Up" {
		zboth.Warn().Msgf("The instance called %s is already running.", givenName)
	} else {
		if errCreateFolder := modifyContainer(givenName, "mkdir -p", "shared/pullin", ""); !errCreateFolder {
			zboth.Fatal().Err(toError("create shared/pullin failed")).Msgf("Failed to create folder inside the respective container.")
		}
		if _, success, _ := gotoFolder(givenName), callVirtualizer(composeCall+"up -d"), gotoFolder("workdir"); success {
			waitFor := 120 // in seconds
			if status == "Exited" {
				waitFor = 30
			}
			zboth.Info().Msgf("Starting instance called %s.", givenName) // because user sees the spinner
			waitTime := waitStartSpinner(waitFor, givenName)
			if waitTime >= 0 {
				var timeTaken string
				if waitTime > 0 {
					timeTaken = toSprintf(" in %d seconds", waitTime)
				}
				zboth.Info().Msgf("Successfully started instance called %s%s at %s.", givenName, timeTaken, conf.GetString(joinKey(instancesWord, givenName, "accessAddress")))
			} else {
				zboth.Fatal().Err(toError("ping timeout after %d seconds", waitTime)).Msgf("Failed to start instance called %s. Please check logs using `%s instance %s`.", givenName, commandForCLI, logInstanceRootCmd.Use)
			}
		} else {
			zboth.Fatal().Msgf("Failed to start instance called %s.", givenName)
		}
	}
}

func instanceStop(givenName string) {
	status := instanceStatus(givenName)
	if status == "Up" {
		if _, success, _ := gotoFolder(givenName), callVirtualizer(composeCall+"stop"), gotoFolder("workdir"); success {
			zboth.Info().Msgf("Successfully stopped instance called %s.", givenName)
		} else {
			zboth.Fatal().Msgf("Failed to stop instance called %s.", givenName)
		}
	} else {
		zboth.Warn().Msgf("Cannot stop instance %s. It seems to be %s.", givenName, status)
	}
}

func instanceRestart(givenName string) {
	instanceStop(givenName)
	instanceStart(givenName)
}

var restartRootCmd = &cobra.Command{
	Use:   "restart [-i <instance_name>]",
	Args:  cobra.NoArgs,
	Short: "Restart the selected instance of " + nameProject,
	Run: func(_ *cobra.Command, _ []string) {
		instanceRestart(currentInstance)
	},
	// TODO: add a force restart flag
}

var onRootCmd = &cobra.Command{
	Use:   "on [-i <instance_name>]",
	Args:  cobra.NoArgs,
	Short: "Start the selected instance of " + nameProject,
	Run: func(_ *cobra.Command, _ []string) {
		instanceStart(currentInstance)
	},
}

var offRootCmd = &cobra.Command{
	Use:   "off [-i <instance_name>]",
	Args:  cobra.NoArgs,
	Short: "Stop the selected instance of " + nameProject,
	Run: func(_ *cobra.Command, _ []string) {
		instanceStop(currentInstance)
	},
}

func init() {
	rootCmd.AddCommand(onRootCmd)
	rootCmd.AddCommand(offRootCmd)
	rootCmd.AddCommand(restartRootCmd)
}
