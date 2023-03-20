package cli

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

func advancedUninstall(removeLogfile bool) {
	existingInstances := allInstances()
	existingInstances[elementInSlice(currentInstance, &existingInstances)] = existingInstances[len(existingInstances)-1]
	existingInstances[len(existingInstances)-1] = currentInstance // move currentInstance to the end of the queue for deletion
	for _, inst := range existingInstances {
		zboth.Info().Msgf("Removing instance called %s.", inst)
		if err := instanceRemove(inst, true); err != nil {
			zboth.Warn().Err(err).Msgf(err.Error())
			zboth.Fatal().Err(toError("uninstalled failed")).Msgf("Uninstall failed while trying to remove %s", inst)
		}
	}
	if err := workDir.Join(instancesWord).RemoveAll(); err != nil {
		zboth.Warn().Err(err).Msgf("Failed to delete the `%s` folder.", instancesWord)
	}
	if err := workDir.Join(conf.ConfigFileUsed()).Remove(); err != nil {
		zboth.Warn().Err(err).Msgf("Failed to delete the configuration file: %s.", conf.ConfigFileUsed())
	}
	zboth.Info().Msgf("%s was successfully uninstalled.", nameCLI)
	if removeLogfile {
		if err := workDir.Join(defaultLogFilename).Remove(); err != nil {
			zboth.Warn().Err(err).Msgf("Failed to delete the log file: %s.", defaultLogFilename)
		}
	}
}

var uninstallAdvancedRootCmd = &cobra.Command{
	Use:   "uninstall",
	Args:  cobra.NoArgs,
	Short: toSprintf("uninstall %s completely", nameCLI),
	Run: func(_ *cobra.Command, _ []string) {
		if isInteractive(false) {
			zerolog.SetGlobalLevel(zerolog.DebugLevel) // uninstall operates in debug mode
			zboth.Debug().Msgf("Uninstall operates in debug mode!")
			if selectYesNo("Are you sure you want to uninstall "+nameCLI, false) {
				switch selectOpt([]string{"yes", "no", "exit"}, "Do you want to keep the log file after successful uninstallation") {
				case "exit":
					// ideally this case is handled in the selectOpt function, here as a safety precaution
					os.Exit(0)
				case "yes":
					advancedUninstall(false)
				case "no":
					advancedUninstall(true)
				}
			} else {
				zboth.Info().Msgf("Nothing was done.")
				if !conf.GetBool(joinKey(stateWord, "debug")) {
					zerolog.SetGlobalLevel(zerolog.InfoLevel)
				}
			}
		} else {
			zboth.Fatal().Err(toError("uninstall in silent mode")).Msgf("For security reasons, this command will not run in silent mode.")
		}
	},
}

func init() {
	advancedRootCmd.AddCommand(uninstallAdvancedRootCmd)
}
