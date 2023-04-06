package cli

import (
	"github.com/spf13/cobra"
)

func instanceRemove(givenName string, force bool) (err error) {
	name := getInternalName(givenName)
	// stop and delete instance
	if force {
		if _, success, _ := gotoFolder(givenName), callVirtualizer(composeCall+"kill"), gotoFolder("workdir"); success {
			zboth.Debug().Msgf("Successfully killed container of instance called %s.", givenName)
		} else {
			err = toError("failed to kill the containers associated with instance %s", givenName)
			return
		}
	}
	if _, success, _ := gotoFolder(givenName), callVirtualizer(composeCall+"down --remove-orphans --volumes"), gotoFolder("workdir"); success {
		zboth.Debug().Msgf("Successfully removed container of instance called %s.", givenName)
	} else {
		err = toError("failed to remove the containers associated with instance %s", givenName)
		return
	}
	// delete folders: shared folder and then named instance folder
	if deleteShared := modifyContainer(givenName, "rm -rf", "shared", ""); deleteShared {
		zboth.Debug().Msgf("Successfully removed `shared` folder associated with %s.", givenName)
		if deleteServices := modifyContainer(givenName, "rm -rf", "services", ""); deleteServices {
			zboth.Debug().Msgf("Successfully removed `services` folder associated with %s.", givenName)
		}
		if deleteFolder := workDir.Join(instancesWord, name).RemoveAll(); deleteFolder == nil {
			zboth.Debug().Msgf("Successfully removed named instance folder associated with %s.", givenName)
		} else {
			err = toError("failed to delete associated folder `%s` in `%s`", name, instancesWord)
			return
		}
	} else {
		err = toError("failed to remove the `shared` associated with instance %s; you may require admin priviledges to remove it", givenName)
		return
	}
	// delete entry in config
	configMap := conf.GetStringMap(instancesWord)
	delete(configMap, givenName)
	conf.Set(instancesWord, configMap)
	if err = writeConfig(false); err != nil {
		err = toError("fail to rewrite configuration file")
	}
	return
}

var removeInstanceRootCmd = &cobra.Command{
	Use:   "remove",
	Args:  cobra.NoArgs,
	Short: "Remove an existing instance of " + nameProject,
	Run: func(cmd *cobra.Command, _ []string) {
		if len(allInstances()) == 1 {
			zboth.Fatal().Err(toError("only one instance")).Msgf("Cannot delete the only instance. Use `%s %s %s` remove %s entirely", commandForCLI, advancedRootCmd.Use, uninstallAdvancedRootCmd.Use, commandForCLI)
		}
		var (
			givenName string
			force     bool
		)
		if ownCall(cmd) {
			if cmd.Flag("name").Changed {
				givenName = cmd.Flag("name").Value.String()
				if err := instanceValidate(givenName); err != nil {
					zboth.Fatal().Err(err).Msgf(err.Error())
				}
			} else {
				isInteractive(true)
				givenName = selectInstance("remove")
			}
		} else {
			if isInteractive(false) {
				givenName = selectInstance("remove")
			} else {
				zboth.Fatal().Err(toError("unexpected operation")).Msgf("Please repeat your actions with the `--debug` flag and report this error.")
			}
		}
		if givenName == currentInstance {
			zboth.Fatal().Err(toError("illegal operation")).Msgf("Cannot delete the currently selected instance. Use `%s instance %s` to switch selection to another instance before proceeding.", commandForCLI, switchInstanceRootCmd.Use)
		}
		status := instanceStatus(givenName)
		if elementInSlice(status, &[]string{"Exited", "Created"}) == -1 {
			zboth.Warn().Msgf("The instance %s is %s.", givenName, status)
			if ownCall(cmd) {
				force = toBool(cmd.Flag("force").Value.String())
				if !force && isInteractive(true) {
					force = selectYesNo("Force remove", false)
				}
			} else {
				if isInteractive(false) {
					force = selectYesNo("Force remove", false)
				} else {
					zboth.Fatal().Err(toError("unexpected operation")).Msgf("Please repeat your actions with the `--debug` flag and report this error.")
				}
			}
		}
		if err := instanceRemove(givenName, force); err != nil {
			zboth.Warn().Err(err).Msgf(err.Error())
		}
	},
}

func init() {
	instanceRootCmd.AddCommand(removeInstanceRootCmd)
	removeInstanceRootCmd.Flags().StringP("name", "n", "", "name of the instance to remove")
	removeInstanceRootCmd.Flags().Bool("force", false, "force remove an instance (very risky)")
}
