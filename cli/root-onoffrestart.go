package cli

import (
	"context"
	"strings"
	"time"

	spinner "charm.land/huh/v2/spinner"
	"github.com/spf13/cobra"
)

func isInstancePingable(ctx context.Context) (err error) {
	givenName := ctx.Value("instance").(string)
	var response string
	for {
		select {
		case <-ctx.Done():
			err = ctx.Err()
			return
		default:
			time.Sleep(1 * time.Second)
			response = instancePing(givenName) // dynamically checks the started instance name
			if response == "200 OK" {
				err = nil
				return
			}
			if strings.Contains(response, "x509") {
				err = toError("ping failed because: certificate signed by unknown authority")
				return
			}
			err = toError(response)
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
		waitFor := 120 // in seconds
		if status == "Exited" {
			waitFor = 60 // in seconds
		}
		startTime := time.Now()
		if _, success, _ := gotoFolder(givenName), callVirtualizer(composeCall+"up -d"), gotoFolder("work.dir"); !success {
			zboth.Fatal().Msgf("Failed to start instance called %s.", givenName)
		}
		// wait for instance to be pingable
		ctx := context.Background()
		ctx = context.WithValue(ctx, "instance", givenName)
		ctx, cancel := context.WithTimeout(ctx, time.Duration(waitFor)*time.Second)
		defer cancel()
		err := spinner.New().Title(toSprintf(" Pinging instance called %s.", givenName)).ActionWithErr(isInstancePingable).Context(ctx).Type(spinner.Points).Run()
		waitTime := int(time.Since(startTime).Seconds())
		if err != nil {
			if err.Error() == "ping failed because: certificate signed by unknown authority" {
				zboth.Warn().Msgf("Ping failed because of an SSL error. This might be because the instance is using a self-signed certificate. Please check if you can access the instance at %s. If you can, then you can ignore this warning.", conf.GetString(joinKey(instancesWord, givenName, "accessAddress")))
			} else {
				zboth.Fatal().Err(toError("ping timeout after %d seconds", waitTime)).Msgf("Failed to ping the instance called %s. Try accessing it yourself. Also, please check logs using `%s instance %s`.", givenName, commandForCLI, logInstanceRootCmd.Use)
			}
		} else {
			zboth.Info().Msgf("Successfully started instance called %s in %d seconds at %s.", givenName, waitTime, conf.GetString(joinKey(instancesWord, givenName, "accessAddress")))
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
