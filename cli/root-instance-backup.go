package cli

import (
	"github.com/spf13/cobra"
)

func instanceBackup(givenName, portion string) {
	// deliver payload
	//TODO: include version check before delivering payload
	gotoFolder(givenName)
	var err, msg string
	status := instanceStatus(givenName)
	if successStart := callVirtualizer(composeCall + "start eln"); successStart {
		if successCurl := callVirtualizer(composeCall + "exec eln curl https://raw.githubusercontent.com/harivyasi/chemotion/chemotion-cli/chemotion-cli/payload/backup.sh --output /embed/scripts/backup.sh"); successCurl {
			if successBackUp := callVirtualizer(composeCall + "exec --env BACKUP_WHAT=" + portion + " eln chemotion backup"); successBackUp {
				zboth.Info().Msgf("Backup successful.")
			} else {
				msg = "Backup process failed."
				err = "backup failed"
			}
		} else {
			err = "backup.sh update failed"
			msg = "Could not fix the broken `backup.sh`. Can't create backup."
		}
		if status != "Up" { // if instance was not Up prior to start then stop it now
			callVirtualizer(composeCall + "stop") // need to be low-level because only one service is running
		}
	} else {
		err = "starting eln service failed"
		msg = "Could not backup unless it starts. Can't create backup."
	}
	gotoFolder("workdir")
	if err != "" {
		zboth.Fatal().Err(toError(err)).Msgf(msg)
	}
}

var backupInstanceRootCmd = &cobra.Command{
	Use:   "backup",
	Short: "Create a backup of the data associated to an instance of " + nameProject,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, _ []string) {
		backup, status := true, instanceStatus(currentInstance)
		if status == "Up" {
			zboth.Warn().Err(toError("instance running")).Msgf("The instance called %s is running. Backing up a running instance is not a good idea.", currentInstance)
			if isInteractive(false) {
				backup = selectYesNo("Continue", false)
			}
		}
		if status == "Created" {
			zboth.Warn().Err(toError("instance never run")).Msgf("The instance called %s was created but never turned on. Backing up such an instance is not a good idea.", currentInstance)
			if isInteractive(false) {
				backup = selectYesNo("Continue", false)
			}
		}
		if backup {
			portion := "both"
			if ownCall(cmd) {
				if toBool(cmd.Flag("db").Value.String()) && !toBool(cmd.Flag("data").Value.String()) {
					portion = "db"
				}
				if toBool(cmd.Flag("data").Value.String()) && !toBool(cmd.Flag("db").Value.String()) {
					portion = "data"
				}
			} else {
				if isInteractive(false) {
					switch selectOpt([]string{"database and data", "database", "data", "exit"}, "What would you like to backup?") {
					case "database and data":
						portion = "both"
					case "database":
						portion = "db"
					case "data":
						portion = "data"
					}
				}
			}
			instanceBackup(currentInstance, portion)
		} else {
			zboth.Debug().Msgf("Backup operation cancelled.")
		}
	},
}

func init() {
	backupInstanceRootCmd.Flags().Bool("db", false, "backup only database")
	backupInstanceRootCmd.Flags().Bool("data", false, "backup only data")
	instanceRootCmd.AddCommand(backupInstanceRootCmd)
}
