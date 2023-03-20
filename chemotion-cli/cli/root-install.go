package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// command to install a new container of Chemotion
var installRootCmd = &cobra.Command{
	Use:    "install",
	Args:   cobra.NoArgs,
	Short:  "Initialize the configuration file and install the first instance of " + nameCLI,
	Hidden: !firstRun,
	Run: func(cmd *cobra.Command, args []string) {
		logWhere()
		if firstRun {
			if currentState.quiet || newInstanceInteraction() {
				if currentState.quiet {
					zboth.Info().Msgf("You chose do first run of %s in quiet mode. Will go ahead and install it!", nameCLI)
				}
				if success := instanceCreate(_root_instance_new_name_, _root_instance_new_use_, "Production", _root_instance_new_address_); success {
					zboth.Info().Msgf("All done! Now you can do `%s on` and `%s off` to start/stop %s.", rootCmd.Name(), rootCmd.Name(), nameCLI)
				}
			}
		} else {
			zboth.Fatal().Err(fmt.Errorf("config file found")).Msgf("This option `%s` is only available for initial installation. Use `%s %s %s` if you wish to create more instances of %s.", cmd.Name(), rootCmd.Name(), instanceRootCmd.Name(), newInstanceRootCmd.Name(), nameCLI)
		}
	},
}

func init() {
	rootCmd.AddCommand(installRootCmd)
	installRootCmd.Flags().StringVar(&_root_instance_new_name_, "name", instanceDefault, "Name of the first instance to create")
	installRootCmd.Flags().StringVar(&_root_instance_new_use_, "use", composeURL, "URL or filepath to use for creating the instance")
	installRootCmd.Flags().StringVar(&_root_instance_new_address_, "address", addressDefault, "Web-address (or hostname) for accessing the instance")
	installRootCmd.Flags().StringVar(&_root_instance_new_env_, "env", "", ".env file for the first instance")
}
