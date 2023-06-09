package cli

import (
	"strings"

	"github.com/chigopher/pathlib"
	"github.com/spf13/cobra"
)

func instanceBackup(givenName, portion string) {
	calledCmd := toSprintf("--profile execution run --rm -e BACKUP_WHAT=%s executor chemotion backup", portion)
	var backupFile pathlib.Path
	if strings.HasSuffix(conf.GetString(joinKey(instancesWord, currentInstance, "image")), "eln-1.3.1p220712") {
		backupFile = downloadFile(backupshURL, "temp.backup.sh")
		var err error
		var output string
		var out []byte
		commands := []string{"create --name backcon ptrxyz/chemotion:eln-1.3.1p220712", "cp temp.backup.sh backcon:/embed/scripts/backup.sh", "commit backcon ptrxyz/chemotion:eln-1.3.1p220712", "rm backcon"}
		for _, comm := range commands {
			out, err = execShell(toSprintf("%s %s", virtualizer, comm))
			output = toSprintf("%s %s", output, out)
			if err != nil {
				zboth.Fatal().Err(err).Msgf("command failed: %s %s", virtualizer, comm)
			}
		}
		zboth.Debug().Msgf(output)
		backupFile.Remove()
	}
	if _, successBackUp, _ := gotoFolder(givenName), callVirtualizer(composeCall+calledCmd), gotoFolder("work.dir"); successBackUp {
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
					switch selectOpt([]string{"database and data", "database", "data", coloredExit}, "What would you like to backup?") {
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
			if elementInSlice(status, &[]string{"Exited", "Created"}) != -1 { // i.e. status == "Exited" || status == "Created"
				_, _, _ = gotoFolder(currentInstance), callVirtualizer(composeCall+"stop"), gotoFolder("work.dir")
			}
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
