package cli

import (
	"fmt"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

var uninstallAdvancedRootCmd = &cobra.Command{
	Use:   "uninstall (accepts no flags)",
	Args:  cobra.NoArgs,
	Short: fmt.Sprintf("uninstall %s completely", nameCLI),
	Run: func(cmd *cobra.Command, args []string) {
		if currentState.quiet {
			fmt.Println("For security reasons, this command will not run in silent mode.")
			zboth.Fatal().Msgf("For security reasons, this command will not run in silent mode.")
		}
		zerolog.SetGlobalLevel(zerolog.DebugLevel) // uninstall operates in debug mode
		fmt.Println("Uninstall operates in debug mode!")
		logWhere()
		confirmInstalled()
		confirmInteractive()
		if selectYesNo("Are you sure you want to uninstall "+nameCLI, false) {
			chosen := conf.GetString(selector_key)
			instances := append(allInstances(), chosen)
			skip := true
			for _, inst := range instances {
				if inst == chosen && skip { // contraption to make sure that the chosen instance is deleted last
					skip = false
					continue
				}
				zboth.Info().Msgf("Removing instance called %s.", inst)
				_root_instance_remove_force_ = true
				if !instanceRemove(inst) {
					zboth.Fatal().Err(fmt.Errorf("uninstalled failed")).Msgf("Uninstall failed while trying to remove %s", inst)
					break
				}
			}
			if err := workDir.Join(instancesFolder).RemoveAll(); err != nil {
				zboth.Warn().Err(err).Msgf("Failed to delete the `%s` folder.", instancesFolder)
			}
			if err := workDir.Join(conf.ConfigFileUsed()).Remove(); err != nil {
				zboth.Warn().Err(err).Msgf("Failed to delete the configuration file: %s.", conf.ConfigFileUsed())
			}
			zboth.Info().Msgf("%s was successfully uninstalled.", nameCLI)
			if selectYesNo("Do you want to remove the log file as well", false) {
				if err := workDir.Join(logFilename).Remove(); err != nil {
					zboth.Warn().Err(err).Msgf("Failed to delete the log file: %s.", logFilename)
				}
			}
		} else {
			zboth.Info().Msgf("Nothing was done.")
		}
	},
}

func init() {
	advancedRootCmd.AddCommand(uninstallAdvancedRootCmd)
}
