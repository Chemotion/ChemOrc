package cli

import (
	"github.com/spf13/cobra"
)

var restoreInstanceRootCmd = &cobra.Command{
	Use:   "restore",
	Args:  cobra.NoArgs,
	Short: "Restore an instance of " + nameCLI,
	Run: func(cmd *cobra.Command, _ []string) {
		var db, data string
		if ownCall(cmd) {
			db, data = cmd.Flag("db").Value.String(), cmd.Flag("data").Value.String()
		}
		if (db == "" || data == "") && isInteractive(true) {
			if db == "" {
				db = getString("Please enter the absolute path of your database backup file (e.g. /path/to/backup.sql.gz)", fileValidate)
			}
			if data == "" {
				data = getString("Please enter the absolute path of your data backup file (e.g. /path/to/backup.data.tar.gz)", fileValidate)
			}
		}
		details := make(map[string]string)
		if create := processInstanceCreateCmd(cmd, details); create {
			var success bool
			if success = instanceCreate(details); success {
				zboth.Info().Msgf("Successfully created a new production instance. Now restoring the backup into it")
				if _, success, _ = gotoFolder(details["givenName"]), callVirtualizer(toSprintf("cp %s %s-%s-%d:/backup/backup.data.tar.gz", data, details["name"], primaryService, rollNum)), gotoFolder("work.dir"); success {
					if _, success, _ = gotoFolder(details["givenName"]), callVirtualizer(toSprintf("cp %s %s-%s-%d:/backup/backup.sql.gz", db, details["name"], primaryService, rollNum)), gotoFolder("work.dir"); success {
						zboth.Info().Msgf("Backup files copied successfully. Now attempting restore them!")
						if _, success, _ = gotoFolder(details["givenName"]), callVirtualizer(composeCall+"--profile execution run --rm -e FORCE_DB_RESET=1 executor chemotion restore"), gotoFolder("work.dir"); success {
							zboth.Info().Msgf("Restoration completed successfully. Once switched on, `%s` can be found at: %s", details["givenName"], details["accessAddress"])
						}
						if _, successDataMv, _ := gotoFolder(details["givenName"]), callVirtualizer(composeCall+"--profile execution run --rm executor mv /backup/backup.data.tar.gz /backup/backup-first.data.tar.gz"), gotoFolder("work.dir"); successDataMv {
							_, _, _ = gotoFolder(details["givenName"]), callVirtualizer(composeCall+"--profile execution run --rm executor ln -s /backup/backup-first.data.tar.gz /backup/backup.data.tar.gz"), gotoFolder("work.dir")
						} else {
							zboth.Warn().Msgf("Failed to rename the data backup file. Symlinks to new backups in this new instance will not be created till this file is renamed.")
						}
						if _, successDBMv, _ := gotoFolder(details["givenName"]), callVirtualizer(composeCall+"--profile execution run --rm executor mv /backup/backup.sql.gz /backup/backup-first.sql.gz"), gotoFolder("work.dir"); successDBMv {
							_, _, _ = gotoFolder(details["givenName"]), callVirtualizer(composeCall+"--profile execution run --rm executor ln -s /backup/backup-first.sql.gz /backup/backup.sql.gz"), gotoFolder("work.dir")
						} else {
							zboth.Warn().Msgf("Failed to rename the database backup file. Symlinks to new backups in this new instance will not be created till this file is renamed.")
						}
					} else {
						zboth.Warn().Msgf("Failed to copy the database backup file %s into the instance.", db)
					}
				} else {
					zboth.Warn().Msgf("Failed to copy the data backup file %s into the instance.", data)
				}
			} else {
				zboth.Fatal().Msgf("Failed to create an instance to restore the files into!")
			}
			if success {
				_, _, _ = gotoFolder(details["givenName"]), callVirtualizer(composeCall+"stop"), gotoFolder("work.dir")
			} else {
				zboth.Warn().Msgf("Restoration process failed. Check LOG!")
				if selectYesNo(toSprintf("Do you want to remove this instance called %s because the restoration process failed?", details["givenName"]), true) {
					if err := instanceRemove(details["givenName"], true); err != nil {
						zboth.Fatal().Err(err).Msgf("Failed to remove the instance called %s.", details["givenName"])
					}
				}
			}
		}
	},
}

func init() {
	instanceRootCmd.AddCommand(restoreInstanceRootCmd)
	restoreInstanceRootCmd.Flags().StringP("name", "n", "must.be.given", "Name for the new instance")
	restoreInstanceRootCmd.Flags().String("suffix", getNewUniqueID(), "An assigned suffix, instead of a random one. Use only if necessary.")
	restoreInstanceRootCmd.Flags().String("use", composeURL, "URL or filepath of the compose file to use for creating the instance")
	restoreInstanceRootCmd.Flags().String("address", addressDefault, "Web-address (or hostname) for accessing the instance")
	restoreInstanceRootCmd.Flags().String("db", "", "Absolute path to the database backup file (e.g. /my/path/backup.sql.gz)")
	restoreInstanceRootCmd.Flags().String("data", "", "Absolute path to the database backup file (e.g. /my/path/backup.data.tar.gz)")
}
