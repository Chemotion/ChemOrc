package cli

import (
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

var _instance_upgrade_use_ string

func instanceUpgrade(givenName, use string) {
	// first read the new compose file
	tempCompose := getCompose(use)
	name := getInternalName(givenName)
	// back up old compose file
	oldComposeFile := workDir.Join(instancesFolder, name, defaultComposeFilename)
	oldCompose := getCompose(oldComposeFile.String())
	// for some reason (no idea why), labels must be set before port
	setUniqueLabels(&tempCompose, name)
	tempCompose.Set(joinKey("services", "eln", "ports"), oldCompose.GetStringSlice(joinKey("services", "eln", "ports"))) // copy over the ports listing
	// shutdown existing instance's docker
	success := true
	if _, worked, _ := gotoFolder(givenName), callVirtualizer("compose down --remove-orphans"), gotoFolder("workdir"); !worked {
		success = worked
		zboth.Fatal().Err(fmt.Errorf("compose down failed")).Msgf("Failed to stop %s. Check log. ABORT!", givenName)
	}
	if success {
		if _, worked, _ := gotoFolder(givenName), callVirtualizer(fmt.Sprintf("volume rm %s_chemotion_app", name)), gotoFolder("workdir"); !worked {
			success = worked
			zboth.Fatal().Err(fmt.Errorf("volume removal failed")).Msgf("Failed to remove old app volume. Check log. ABORT!")
		}
	}
	if success {
		oldComposeFile.Rename(workDir.Join(instancesFolder, name, strconv.FormatInt(time.Now().Unix(), 10)+".old."+defaultComposeFilename))
		if err := tempCompose.WriteConfigAs(workDir.Join(instancesFolder, name, defaultComposeFilename).String()); err == nil {
			commandStr := fmt.Sprintf("compose -f %s up --no-start", defaultComposeFilename)
			zboth.Info().Msgf("Starting %s with command: %s", virtualizer, commandStr)
			if _, worked, _ := gotoFolder(givenName), callVirtualizer(commandStr), gotoFolder("workdir"); !worked {
				zboth.Fatal().Err(fmt.Errorf("%s failed", commandStr)).Msgf("Failed to initialize upgraded %s. Check log. ABORT!", givenName)
			}
		} else {
			zboth.Fatal().Err(fmt.Errorf("compose file write fail")).Msgf("Failed to write the new compose file. The old one is still available as %s", oldComposeFile.Name())
		}
	}
}

var upgradeInstanceRootCmd = &cobra.Command{
	Use:   "upgrade",
	Args:  cobra.NoArgs,
	Short: "Upgrade (the selected) instance of " + nameCLI,
	Run: func(cmd *cobra.Command, args []string) {
		logWhere()
		confirmInstalled()
		if instanceStatus(currentState.name) == "Up" {
			zboth.Fatal().Err(fmt.Errorf("upgrade fail; instance is up")).Msgf("Cannot upgrade an instance that is currently running. Please turn it off before continuing.")
		} else {
			if selectYesNo("Please be sure to backup before proceeding. Continue", false) {
				instanceUpgrade(currentState.name, _instance_upgrade_use_)
			}
		}
	},
}

func init() {
	upgradeInstanceRootCmd.Flags().StringVar(&_instance_upgrade_use_, "use", composeURL, "URL or filepath of the compose file to use for upgrading")
	instanceRootCmd.AddCommand(upgradeInstanceRootCmd)
}
