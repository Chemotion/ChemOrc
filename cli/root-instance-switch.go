package cli

import (
	color "github.com/mitchellh/colorstring"
	"github.com/spf13/cobra"
)

func instanceSwitch(givenName string) {
	conf.Set(joinKey(stateWord, selectorWord), givenName)
	if err := writeConfig(false); err == nil {
		currentInstance = givenName
		zboth.Info().Msgf("Instance being managed switched to %s.", color.Color(toSprintf("[green]%s", currentInstance)))
	} else {
		zboth.Fatal().Err(err).Msgf("Failed to update the selected instance.")
	}
}

var switchInstanceRootCmd = &cobra.Command{
	Use:   "switch",
	Short: "Switch to an instance of " + nameProject,
	Run: func(cmd *cobra.Command, _ []string) {
		if len(allInstances()) == 1 {
			zboth.Fatal().Err(toError("only one instance")).Msgf("You cannot switch because you only have one instance.")

		}
		if ownCall(cmd) {
			if cmd.Flag("name").Changed {
				givenName := cmd.Flag("name").Value.String()
				if err := instanceValidate(givenName); err != nil {
					zboth.Fatal().Err(err).Msg(err.Error())
				}
				instanceSwitch(givenName)
			} else {
				isInteractive(true)
				instanceSwitch(selectInstance("switch to"))
			}
		} else {
			if isInteractive(false) {
				instanceSwitch(selectInstance("switch to"))
			} else {
				zboth.Fatal().Err(toError("unexpected operation")).Msg("Please repeat your actions with the `--debug` flag and report this error.")
			}
		}
	},
}

func init() {
	instanceRootCmd.AddCommand(switchInstanceRootCmd)
	switchInstanceRootCmd.Flags().StringP("name", "n", "", "Name of instance to switch to.")
}
