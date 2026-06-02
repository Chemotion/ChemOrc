package cli

import (
	"strings"

	"github.com/spf13/cobra"
)

func resetInstance(givenName string) (err error) {
	status := instanceStatus(givenName)
	if elementInSlice(status, &[]string{"Exited", "Created"}) != -1 { // checks if the instance is off
		target := getInternalName(givenName) + "_chemotion_app"
		command := toSprintf("%s %s volumes --quiet", virtualizer, composeCall)
		var out []byte
		gotoFolder(givenName)
		out, err = execShell(command)
		gotoFolder("work.dir")
		if err == nil {
			err = toError("volume name mismatch/indeterminable")
			volumes := strings.Split(string(out), "\n")
			for _, volume := range volumes {
				if strings.Contains(volume, "app") {
					if volume == target {
						if _, success, _ := gotoFolder(givenName), callVirtualizer(composeCall+"down"), gotoFolder("work.dir"); !success {
							err = toError("failed to `down` the instance %s because %s", givenName, err.Error())
							return
						}
						zboth.Info().Msgf("Resetting instance by removing volume %s", target)
						command = toSprintf("%s volume rm %s", virtualizer, target)
						out, err = execShell(command)
						if strings.TrimSpace(string(out)) == target {
							zboth.Info().Msgf("Successfully removed volume %s", target)
							if _, success, _ := gotoFolder(givenName), callVirtualizer(composeCall+"up --no-start"), gotoFolder("work.dir"); !success {
								zboth.Fatal().Err(toError("compose up failed")).Msgf("Failed to re-establish the instance of %s. Check log. ABORT! Tread with caution!", nameProject)
							}
						} else {
							if err == nil {
								err = toError("volume removal failed")
							} else {
								err = toError("volume removal failed because: %s", err.Error())
							}
						}
						return
					}
				}
			}
			return
		}
	} else {
		zboth.Warn().Msgf("Cannot reset instance %s. It seems to be %s. It must be Exited state i.e. turned-off before resetting.", givenName, status)
		err = toError("instance %s is %s", givenName, status)
	}
	return
}

var resetInstanceRootCmd = &cobra.Command{
	Use:   "reset",
	Args:  cobra.NoArgs,
	Short: "Reset an instance of " + nameProject,
	Run: func(cmd *cobra.Command, _ []string) {
		if elementInSlice(instanceStatus(currentInstance), &[]string{"Exited", "Created"}) != -1 { // checks if the instance is off
			if err := resetInstance(currentInstance); err != nil {
				zboth.Fatal().Err(err).Msgf("Failed to reset the instance called %s.", currentInstance)
			} else {
				zboth.Info().Msgf("Successfully reset the instance called %s.", currentInstance)
			}
		} else {
			zboth.Warn().Msgf("Cannot reset instance that is not Exited or Created.")
		}
	},
}

func init() {
	instanceRootCmd.AddCommand(resetInstanceRootCmd)
}
