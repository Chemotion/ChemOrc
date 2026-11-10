package cli

import (
	"path/filepath"

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
			for flag, path := range map[string]string{"db": db, "data": data} {
				if path != "" {
					if err := fileValidate(path); err != nil {
						zboth.Fatal().Err(err).Msgf("Cannot use %s as the --%s backup file: %s", path, flag, err.Error())
					}
				}
			}
		}
		if (db == "" || data == "") && isInteractive(true) {
			if db == "" {
				db = getString("Please enter the absolute path of your database backup file (e.g. /path/to/backup.sql.gz)", []string{}, fileValidate)
			}
			if data == "" {
				data = getString("Please enter the absolute path of your data backup file (e.g. /path/to/backup.data.tar.gz)", []string{}, fileValidate)
			}
		}
		// the paths were checked from the current folder, but `docker cp` runs from the instance's folder: make them absolute
		for _, path := range []*string{&db, &data} {
			if abs, err := filepath.Abs(*path); err == nil {
				*path = abs
			} else {
				zboth.Fatal().Err(err).Msgf("Cannot determine the absolute path of %s.", *path)
			}
		}
		details := make(map[string]string)
		if create := processInstanceCreateCmd(cmd, details); create {
			var success bool
			if success = instanceCreate(details); success {
				zboth.Info().Msgf("Successfully created a new production instance. Now restoring the backup into it")
				container := toSprintf("%s-%s-%d", details["name"], primaryService, rollNum)
				// paths are passed as single arguments so that they may contain spaces
				if _, success, _ = gotoFolder(details["givenName"]), callVirtualizerArgs("cp", data, container+":/backup/backup.data.tar.gz"), gotoFolder("work.dir"); success {
					if _, success, _ = gotoFolder(details["givenName"]), callVirtualizerArgs("cp", db, container+":/backup/backup.sql.gz"), gotoFolder("work.dir"); success {
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
				if isInteractive(false) {
					if selectYesNo(toSprintf("Do you want to remove this instance called %s because the restoration process failed?", details["givenName"]), true) {
						if err := instanceRemove(details["givenName"], true); err != nil {
							zboth.Fatal().Err(err).Msgf("Failed to remove the instance called %s.", details["givenName"])
						}
					}
				} else {
					// no prompt in quiet mode: keep the instance rather than delete anything unasked,
					// and point to the command that can actually remove it (`instance remove` refuses the only and the selected instance)
					givenName := details["givenName"]
					var how string
					switch {
					case len(allInstances()) == 1:
						how = toSprintf("As it is the only instance, remove it with `%s %s %s` (interactive only; this removes the whole installation)", commandForCLI, advancedRootCmd.Use, uninstallAdvancedRootCmd.Use)
					case givenName == currentInstance:
						how = toSprintf("As it is the selected instance, first select another one with `%s instance %s -n <name>`, then remove it with `%s instance %s -n %s --force -q`", commandForCLI, switchInstanceRootCmd.Use, commandForCLI, removeInstanceRootCmd.Use, givenName)
					default:
						how = toSprintf("Remove it with `%s instance %s -n %s --force -q`", commandForCLI, removeInstanceRootCmd.Use, givenName)
					}
					zboth.Warn().Msgf("The instance called %s was kept. %s if it is not needed.", givenName, how)
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
