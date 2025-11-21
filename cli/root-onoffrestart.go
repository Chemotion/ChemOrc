package cli

import (
	"strings"
	"time"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

// show (and then remove) a progress bar that waits for an instance to start
func waitStartSpinner(waitForSeconds int, givenName string) (waitTime int) {
	bar := progressbar.NewOptions(
		-1,
		progressbar.OptionSetDescription(toSprintf("Starting %s...", givenName)),
		progressbar.OptionSetPredictTime(false),
		progressbar.OptionClearOnFinish(),
		progressbar.OptionSetRenderBlankState(true),
		progressbar.OptionSetVisibility(true),
		progressbar.OptionSpinnerType(51),
	)
	startTime := time.Now()
	for {
		time.Sleep(1 * time.Second)
		response := instancePing(givenName)
		timeSince := int(time.Since(startTime).Seconds())
		if response == "200 OK" {
			waitTime = timeSince
			bar.Finish()
			return
		}
		if strings.Contains(response, "x509") {
			waitTime = -1
			bar.Finish()
			zboth.Warn().Err(toError(response)).Msgf("Ping failed because: `certificate signed by unknown authority`.")
			return
		}
		bar.Add(timeSince)
		if timeSince >= waitForSeconds {
			waitTime = -1
			bar.Finish()
			return
		}
	}
}

func instanceStart(givenName string) {
	status := instanceStatus(givenName)
	if status == "Up" {
		zboth.Warn().Msgf("The instance called %s is already running.", givenName)
	} else {
		if errCreateFolder := modifyContainer(givenName, "mkdir -p", "shared/pullin", ""); !errCreateFolder {
			zboth.Fatal().Err(toError("create shared/pullin failed")).Msgf("Failed to create folder inside the respective container.")
		}
		if _, success, _ := gotoFolder(givenName), callVirtualizer(composeCall+"up -d"), gotoFolder("work.dir"); success {
			waitFor := 120 // in seconds
			if status == "Exited" {
				waitFor = 60 // in seconds
			}
			zboth.Info().Msgf("Pinging instance called %s.", givenName) // because user sees the spinner
			waitTime := waitStartSpinner(waitFor, givenName)
			if waitTime >= 0 {
				var timeTaken string
				if waitTime > 0 {
					timeTaken = toSprintf(" in %d seconds", waitTime)
				}
				zboth.Info().Msgf("Successfully started instance called %s%s at %s.", givenName, timeTaken, conf.GetString(joinKey(instancesWord, givenName, "accessAddress")))
			} else {
				zboth.Fatal().Err(toError("ping timeout after %d seconds", waitTime)).Msgf("Failed to ping the instance called %s. Try accessing it yourself. Also, please check logs using `%s instance %s`.", givenName, commandForCLI, logInstanceRootCmd.Use)
			}
		} else {
			zboth.Fatal().Msgf("Failed to start instance called %s.", givenName)
		}
	}
}

func instanceStop(givenName string) {
	status := instanceStatus(givenName)
	if elementInSlice(status, &[]string{"Exited", "Created"}) == -1 {
		if _, success, _ := gotoFolder(givenName), callVirtualizer(composeCall+"stop"), gotoFolder("work.dir"); success {
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
