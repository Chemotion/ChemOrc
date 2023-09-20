package cli

import (
	"strings"
	"time"

	"github.com/chigopher/pathlib"
	vercompare "github.com/hashicorp/go-version"
	"github.com/spf13/cobra"
)

func upgradeRequired() (toUpgrade []string) {
	if conf.IsSet(joinKey(stateWord, "latest_eln")) {
		if latestVersion, err := vercompare.NewVersion(conf.GetString(joinKey(stateWord, "latest_eln"))); err == nil {
			for _, givenName := range allInstances() {
				_, imageName, _ := strings.Cut(conf.GetString(joinKey(instancesWord, givenName, "image")), "/")
				if _, imageVersion, found := strings.Cut(imageName, "-"); found {
					if imVer, err := vercompare.NewVersion(imageVersion); err == nil {
						if latestVersion.GreaterThan(imVer) {
							toUpgrade = append(toUpgrade, givenName)
						}
					}
				}
			}
		}
	}
	return
}

func pullImages(use string) {
	tempCompose := parseCompose(use)
	services := getSubHeadings(&tempCompose, "services")
	if len(services) == 0 {
		zboth.Warn().Err(toError("no services found")).Msgf("Please check that %s is a valid compose file with named services.", tempCompose.ConfigFileUsed())
	}
	for _, service := range services {
		zboth.Info().Msgf("Pulling image for the service called %s", service)
		if success := callVirtualizer(toSprintf("pull %s", tempCompose.GetString(joinKey("services", service, "image")))); !success {
			zboth.Warn().Err(toError("pull failed")).Msgf("Failed to pull image for the service called %s", service)
		}
	}
}

func instanceUpgrade(givenName, use string) {
	var success bool = true
	var newComposeFile pathlib.Path
	name := getInternalName(givenName)
	// download or copy the new compose
	if existingFile(use) {
		dest := workDir.Join(toSprintf("%s.%s", getNewUniqueID(), chemotionComposeFilename))
		if err := copyfile(use, dest.String()); err == nil {
			newComposeFile = *dest
		} else {
			zboth.Fatal().Err(err).Msgf("Failed to copy the suggested compose file: %s. This is necessary for future use.", use)
		}
	} else {
		newComposeFile = downloadFile(use, workDir.Join(toSprintf("%s.%s", getNewUniqueID(), chemotionComposeFilename)).String())
	}
	// get name of new image, and do a kind-of sanity check on the proposed file
	newCompose := parseCompose(newComposeFile.String())
	newImage := newCompose.GetString(joinKey("services", primaryService, "image"))
	if newImage == "" {
		zboth.Fatal().Err(toError("not understood compose file")).Msgf("Failed to identify %s image in the compose file.", primaryService)
	}
	// get port from old compose
	oldComposeFile := workDir.Join(instancesWord, name, chemotionComposeFilename)
	oldCompose := parseCompose(oldComposeFile.String())
	if oldCompose.IsSet(joinKey("services", primaryService, "ports")) {
		if err := changeExposedPort(newComposeFile.String(), oldCompose.GetStringSlice(joinKey("services", primaryService, "ports"))[0][:4]); err != nil {
			newComposeFile.Remove()
			zboth.Fatal().Err(err).Msgf("Failed to update the port in downloaded compose file. This is necessary for future use. The file was removed.")
		}
	} else {
		zboth.Warn().Msgf("The entry %s was not found in the existing %s, remember to configure the new file as per your settings.", joinKey("services", primaryService, "ports"), chemotionComposeFilename)
	}
	// backup the old compose file
	if err := oldComposeFile.Rename(workDir.Join(instancesWord, name, toSprintf("old.%s.%s", time.Now().Format("060102150405"), chemotionComposeFilename))); err == nil {
		zboth.Info().Msgf("The old compose file is now called: %s.", oldComposeFile.String())
	} else {
		if errRemove := newComposeFile.Remove(); errRemove != nil {
			zboth.Warn().Err(err).Msgf("Failed to create backup of the old compose file.")
			zboth.Fatal().Err(errRemove).Msgf("Failed to remove the new compose file. Check log. ABORT!")
		}
		zboth.Fatal().Err(err).Msgf("Failed to create backup of the old compose file.")
	}
	// move the new file in place of the old one
	if err := newComposeFile.Rename(workDir.Join(instancesWord, name, chemotionComposeFilename)); err != nil {
		if errRestore := oldComposeFile.Rename(workDir.Join(instancesWord, name, chemotionComposeFilename)); errRestore == nil {
			zboth.Fatal().Err(err).Msgf("Failed to rename the new compose file: %s. Check log. ABORT!", newComposeFile.String())
		} else {
			zboth.Warn().Err(err).Msgf("Failed to rename the new compose file: %s.", newComposeFile.String())
			zboth.Fatal().Err(errRestore).Msgf("Failed to restore the old compose file %s. Instance will fail to restart. Rename it manually. ABORT!", oldComposeFile.String())
		}
	}
	var err error = nil
	var msg string
	// shutdown existing instance's docker
	if _, success, _ = gotoFolder(givenName), callVirtualizer(composeCall+"down --remove-orphans"), gotoFolder("work.dir"); !success {
		err = toError("compose down failed")
		msg = toSprintf("Failed to stop %s. Check log. ABORT!", givenName)
	}
	// remove the existing volume
	if success {
		if _, success, _ = gotoFolder(givenName), callVirtualizer(toSprintf("volume rm %s_chemotion_app", name)), gotoFolder("work.dir"); !success {
			err = toError("volume removal failed")
			msg = "Failed to remove old app volume. Check log. ABORT!"
		}
	}
	// build the new container and update the config file
	if success {
		commandStr := toSprintf(composeCall + "up --no-start")
		zboth.Info().Msgf("Starting %s with command: %s", virtualizer, commandStr)
		if _, success, _ = gotoFolder(givenName), callVirtualizer(commandStr), gotoFolder("work.dir"); success {
			conf.Set(joinKey(instancesWord, givenName, "image"), newImage)
			writeConfig(false)
			zboth.Info().Msgf("Instance upgraded successfully!")
			func() {
				// fixes bugs in extended compose
				// update image in extended compose
				// remove extra entry from extended compose
				// rewrite labels
				if _, success, _ = gotoFolder(givenName), callVirtualizer("pull mikefarah/yq"), gotoFolder("work.dir"); success {
					var result []byte
					gotoFolder(givenName)
					extendedCompose := parseCompose(cliComposeFilename)
					if extendedCompose.IsSet("networks.chemotion.labels") {
						labels := extendedCompose.GetStringMapString("networks.chemotion.labels")
						if strings.HasPrefix(labels["net.chemotion.cli.project"], givenName) {
							if result, err = execShell(toSprintf("cat %s | %s run -i --rm mikefarah/yq 'del(.networks.*.labels)'", cliComposeFilename, virtualizer)); err == nil {
								yamlFile := pathlib.NewPath(cliComposeFilename)
								err = yamlFile.WriteFile(result)
								if err == nil {
									extendedCompose = parseCompose(cliComposeFilename)
								}
							}
						}
					}
					extendedCompose.Set(joinKey("services", "executor", "image"), newImage)
					if extendedCompose.IsSet(joinKey("networks", "chemotion")) {
						// reset labels on services and volumes for future identification
						sections := []string{"services", "volumes"}
						compose := parseCompose(chemotionComposeFilename)
						for _, section := range sections {
							subheadings := getSubHeadings(&compose, section) // subheadings are the names of the services and volumes
							for _, k := range subheadings {
								extendedCompose.Set(joinKey(section, k, "labels"), []string{toSprintf("net.chemotion.cli.project=%s", name)})
							}
						}
					}
					err = extendedCompose.WriteConfigAs(cliComposeFilename)
					gotoFolder("work.dir")
					if err != nil {
						zboth.Warn().Err(err).Msgf("Failed to correct the file %s for instance %s.", cliComposeFilename, givenName)
					}
				} else {
					zboth.Warn().Err(toError("failed to pull `yq`")).Msgf("Failed to pull `yq` image.")
				}
			}() // to be removed in version 3
		} else {
			err = toError("%s failed", commandStr)
			msg = toSprintf("Failed to initialize upgraded %s. Check log. ABORT!", givenName)
		}
	}
	if err != nil {
		if errRestore := oldComposeFile.Rename(workDir.Join(instancesWord, name, chemotionComposeFilename)); errRestore == nil {
			zboth.Fatal().Err(err).Msg(msg)
		} else {
			zboth.Warn().Err(err).Msgf(msg)
			zboth.Fatal().Err(errRestore).Msgf("Failed to restore the old compose file %s. Instance will fail to restart. Rename it manually. ABORT!", oldComposeFile.String())
		}
	}
}

var upgradeInstanceRootCmd = &cobra.Command{
	Use:   "upgrade",
	Args:  cobra.NoArgs,
	Short: "Upgrade (the selected) instance of " + nameCLI,
	Run: func(cmd *cobra.Command, _ []string) {
		var pull, backup, stop, upgrade bool = false, false, false, true
		var use string = composeURL
		if ownCall(cmd) {
			use = cmd.Flag("use").Value.String()
			pull = toBool(cmd.Flag("pull-only").Value.String())
			upgrade = !pull
		}
		if !pull && isInteractive(false) {
			_, currentVersion, _ := strings.Cut(conf.GetString(joinKey(instancesWord, currentInstance, "image")), "-")
			use = getComposeAddressToUse(currentVersion, "upgrade to")
			switch selectOpt([]string{"all actions: pull image, backup and upgrade", "preparation: pull image and backup", "upgrade only (if already prepared)", "pull image only", coloredExit}, "What do you want to do") {
			case "all actions: pull image, backup and upgrade":
				pull, backup, stop, upgrade = true, true, true, true
			case "preparation: pull image and backup":
				pull, backup, stop, upgrade = true, true, false, false
			case "upgrade only (if already prepared)":
				pull, backup, stop, upgrade = false, false, false, true
			case "pull image only":
				pull, backup, stop, upgrade = true, false, false, false
			}
		}
		if pull {
			pullImages(use)
		}
		if backup {
			instanceBackup(currentInstance, "both")
		}
		if stop {
			instanceStop(currentInstance)
		}
		if upgrade {
			status := instanceStatus(currentInstance)
			if elementInSlice(status, &[]string{"Exited", "Created"}) == -1 {
				zboth.Fatal().Err(toError("upgrade fail; instance is %s", status)).Msgf("Cannot upgrade an instance that is not properly shut down. Please turn it off before continuing.")
			}
			instanceUpgrade(currentInstance, use)
		}
	},
}

func init() {
	upgradeInstanceRootCmd.Flags().String("use", composeURL, "URL or filepath of the compose file to use for upgrading")
	upgradeInstanceRootCmd.Flags().Bool("pull-only", false, "Pull image for use in upgrade, don't do the upgrade")
	instanceRootCmd.AddCommand(upgradeInstanceRootCmd)
}
