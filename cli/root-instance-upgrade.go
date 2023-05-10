package cli

import (
	"strings"
	"time"

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
	name := getInternalName(givenName)
	// download the new compose (in the working directory)
	newComposeFile := downloadFile(composeURL, workDir.String())
	newCompose := parseCompose(newComposeFile.String())
	newImage := newCompose.GetString(joinKey("services", primaryService, "image"))
	// get port from old compose
	oldComposeFile := workDir.Join(instancesWord, name, chemotionComposeFilename)
	oldCompose := parseCompose(oldComposeFile.String())
	if oldCompose.GetStringSlice(joinKey("services", primaryService, "ports"))[0] != toSprintf("%d:%d", firstPort, firstPort) {
		if err := changeKey(newComposeFile.String(), joinKey("services", primaryService, "ports[0]"), oldCompose.GetStringSlice(joinKey("services", primaryService, "ports"))[0]); err != nil {
			newComposeFile.Remove()
			zboth.Fatal().Err(err).Msgf("Failed to update the port in downloaded compose file. This is necessary for future use. The file was removed.")
		}
	}
	// backup the old compose file
	if err := oldComposeFile.Rename(workDir.Join(instancesWord, name, toSprintf("old.%s.%s", time.Now().Format("060102150405"), chemotionComposeFilename))); err == nil {
		zboth.Info().Msgf("The old compose file is now called %s:", oldComposeFile.String())
	} else {
		newComposeFile.Remove()
		zboth.Fatal().Err(err).Msgf("Failed to remove the new compose file. Check log. ABORT!")
	}
	if err := newComposeFile.Rename(workDir.Join(instancesWord, name, chemotionComposeFilename)); err != nil {
		zboth.Fatal().Err(err).Msgf("Failed to rename the new compose file: %s. Check log. ABORT!", newComposeFile.String())
	}
	// shutdown existing instance's docker
	if _, success, _ = gotoFolder(givenName), callVirtualizer(composeCall+"down --remove-orphans"), gotoFolder("workdir"); !success {
		zboth.Fatal().Err(toError("compose down failed")).Msgf("Failed to stop %s. Check log. ABORT!", givenName)
	}
	if success {
		if _, success, _ = gotoFolder(givenName), callVirtualizer(toSprintf("volume rm %s_chemotion_app", name)), gotoFolder("workdir"); !success {
			zboth.Fatal().Err(toError("volume removal failed")).Msgf("Failed to remove old app volume. Check log. ABORT!")
		}
	}
	if success {
		commandStr := toSprintf(composeCall + "up --no-start")
		zboth.Info().Msgf("Starting %s with command: %s", virtualizer, commandStr)
		if _, success, _ = gotoFolder(givenName), callVirtualizer(commandStr), gotoFolder("workdir"); success {
			conf.Set(joinKey(instancesWord, givenName, "image"), newImage)
			writeConfig(false)
			zboth.Info().Msgf("Instance upgraded successfully!")
		} else {
			zboth.Fatal().Err(toError("%s failed", commandStr)).Msgf("Failed to initialize upgraded %s. Check log. ABORT!", givenName)
		}

	}
}

var upgradeInstanceRootCmd = &cobra.Command{
	Use:   "upgrade",
	Args:  cobra.NoArgs,
	Short: "Upgrade (the selected) instance of " + nameCLI,
	Run: func(cmd *cobra.Command, _ []string) {
		var pull, backup, upgrade bool = false, false, true
		var use string = composeURL
		if ownCall(cmd) {
			use = cmd.Flag("use").Value.String()
			pull = toBool(cmd.Flag("pull-only").Value.String())
			upgrade = !pull
		}
		if !pull && isInteractive(false) {
			switch selectOpt([]string{"all actions: pull image, backup and upgrade", "preparation: pull image and backup", "upgrade only (if already prepared)", "pull image only", "exit"}, "What do you want to do") {
			case "all actions: pull image, backup and upgrade":
				pull, backup, upgrade = true, true, true
			case "preparation: pull image and backup":
				pull, backup, upgrade = true, true, false
			case "upgrade only (if already prepared)":
				pull, backup, upgrade = false, false, true
			case "pull image only":
				pull, backup, upgrade = true, false, false
			}
		}
		if pull {
			pullImages(use)
		}
		if backup {
			instanceBackup(currentInstance, "both")
		}
		if upgrade {
			if instanceStatus(currentInstance) == "Up" {
				zboth.Fatal().Err(toError("upgrade fail; instance is up")).Msgf("Cannot upgrade an instance that is currently running. Please turn it off before continuing.")
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
