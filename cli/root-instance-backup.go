package cli

import (
	"strings"

	"github.com/spf13/cobra"
)

func instanceBackup(givenName, portion string) {
	calledCmd := "exec --env BACKUP_WHAT=" + portion + "executor chemotion backup"
	if strings.HasSuffix(conf.GetString(joinKey(instancesWord, currentInstance, "image")), "eln-1.3.1p220712") {
		calledCmd = "exec --env BACKUP_WHAT=" + portion + "executor bash -c \"curl " + backupshURL + " --output /embed/scripts/backup.sh && chemotion backup\""
	}
	if _, successBackUp, _ := gotoFolder(givenName), callVirtualizer(composeCall+calledCmd), gotoFolder("workdir"); successBackUp {
		zboth.Info().Msgf("Backup successful.")
	} else {
		zboth.Fatal().Err(toError("backup failed")).Msgf("Backup process failed.")
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
