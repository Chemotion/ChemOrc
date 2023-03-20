package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func instanceBackup(givenName string) {
	// deliver payload
	//TODO: include version check before delivering payload
	gotoFolder(givenName)
	var err, msg string
	if successStart := callVirtualizer("compose start eln"); successStart {
		if successCurl := callVirtualizer("compose exec eln curl https://raw.githubusercontent.com/harivyasi/chemotion/chemotion-cli/chemotion-cli/payload/backup.sh --output /embed/scripts/backup.sh"); successCurl {
			if successBackUp := callVirtualizer("compose exec eln chemotion backup"); successBackUp {
				zboth.Info().Msgf("Backup successful.")
			} else {
				msg = "Backup process failed."
				err = "backup failed"
			}
		} else {
			err = "backup.sh update failed"
			msg = "Could not fix the broken `backup.sh`. Can't create backup."
		}
		callVirtualizer("compose stop eln")
	} else {
		err = "starting eln service failed"
		msg = "Could not backup unless it starts. Can't create backup."
	}
	gotoFolder("workdir")
	if err != "" {
		zboth.Fatal().Err(fmt.Errorf(err)).Msgf(msg)
	}
}

var backupInstanceRootCmd = &cobra.Command{
	Use:   "backup",
	Short: "Create a backup of the data associated to an instance of " + nameCLI,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		logWhere()
		confirmInstalled()
		backup := true
		if !currentState.quiet {
			status := instanceStatus(currentState.name)
			if status == "Up" {
				backup = selectYesNo(fmt.Sprintf("The instance called %s is running. Backing up a running instance is not a good idea. Continue", currentState.name), false)
			}
			if status == "Created" {
				backup = selectYesNo(fmt.Sprintf("The instance called %s was created but never turned on. Backing up such an instance is not a good idea. Continue", currentState.name), false)
			}
		}
		if backup {
			instanceBackup(currentState.name)
		} else {
			zlog.Debug().Msgf("Backup operation cancelled.")
		}
	},
}

func init() {
	instanceRootCmd.AddCommand(backupInstanceRootCmd)
}
