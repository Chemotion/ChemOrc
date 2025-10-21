package cli

import (
	"os"
	"strings"
	"time"

	"github.com/chigopher/pathlib"
	color "github.com/mitchellh/colorstring"
	"github.com/spf13/viper"
)

func applyPatch(patchName string) (success bool) {
	var applied []string
	if conf.IsSet((joinKey(stateWord, patchWord))) {
		applied = conf.GetStringSlice((joinKey(stateWord, patchWord)))
		zboth.Debug().Msgf("The following patches have been applied: %s.", strings.Join(applied, ", "))
	} else {
		conf.Set(joinKey(stateWord, patchWord), applied)
		zboth.Debug().Msg("No patch has been applied so far.")
	}
	if elementInSlice(patchName, &applied) != -1 {
		success = true
	} else {
		switch patchName {
		case "fix-173-ketcher":
			zboth.Debug().Msgf("Applying patch: %s", patchName)
			// patch for ELN version 1.7.3 docker-compose.yml file
			success = true
			for _, givenName := range allInstances() {
				gotoFolder(givenName)
				var result []byte
				var err error
				compose := parseAndPullCompose(chemotionComposeFilename, false)
				if compose.IsSet("services.ketchersvc.image") {
					eln_image := compose.GetString("services.eln.image")
					ketcher_image := compose.GetString("services.ketchersvc.image")
					if eln_image == "ptrxyz/chemotion:eln-1.7.3" && ketcher_image == "ptrxyz/chemotion:ketchersvc-1.7.3" {
						if result, err = execShell(toSprintf("cat %s | %s run -i --rm mikefarah/yq '.services.ketchersvc.image = \"ptrxyz/chemotion:ketchersvc-1.7.2\"'", chemotionComposeFilename, virtualizer)); err == nil {
							yamlFile := pathlib.NewPath(chemotionComposeFilename)
							err = yamlFile.WriteFile(result)
						}
						if err == nil {
							zboth.Info().Msg(color.Color(toSprintf("[red][bold]Instance %s has been patched. Please restart it ASAP.", givenName)))
						} else {
							success = false
							zboth.Warn().Err(err).Msgf("Failed to update %s for %s. Patch not completely successful.", chemotionComposeFilename, givenName)
						}
					}
				}
				gotoFolder("work.dir")
			}
		}
		if success {
			zboth.Debug().Msgf("Successfully applied patch: %s", patchName)
			applied = append(applied, patchName)
			conf.Set((joinKey(stateWord, patchWord)), applied)
			if existingFile(conf.ConfigFileUsed()) {
				if err := writeConfig(false); err != nil {
					zboth.Warn().Err(err).Msgf("Failed to rewrite config file. You will need to add `%s: - %s` to the %s file manually.", joinKey(stateWord, patchWord), patchName, conf.ConfigFileUsed())
				}
			}
		}
	}
	return
}

func upgradeThisTool(transition string) (success bool) {
	switch transition {
	case "0.1_to_0.2":
		if success = selectYesNo("It seems you are upgrading from version 0.1.x of the tool to 0.2.x. Is this true?", true); success {
			newConfig := viper.New()
			newConfig.Set("version", versionConfig)
			newConfig.Set(joinKey(stateWord, selectorWord), conf.GetString("selected"))
			newConfig.Set(joinKey(stateWord, "debug"), false)
			newConfig.Set(joinKey(stateWord, "quiet"), false)
			newConfig.Set(joinKey(stateWord, "version"), versionCLI)
			instances := getSubHeadings(&conf, instancesWord)
			newConfig.Set(instancesWord, instances)
			for _, givenName := range instances {
				name := conf.GetString(joinKey(instancesWord, givenName, "name"))
				newConfig.Set(joinKey(instancesWord, givenName, "name"), name)
				newConfig.Set(joinKey(instancesWord, givenName, "kind"), conf.GetString(joinKey(instancesWord, givenName, "kind")))
				newConfig.Set(joinKey(instancesWord, givenName, "port"), conf.GetInt(joinKey(instancesWord, givenName, "port")))
				compose := viper.New()
				compose.SetConfigFile(workDir.Join(instancesWord, name, chemotionComposeFilename).String())
				compose.ReadInConfig()
				newConfig.Set(joinKey(instancesWord, givenName, "image"), compose.GetString(joinKey("services", "eln", "image")))
				details := make(map[string]string)
				details["name"] = name
				port, address, portStr := conf.GetInt(joinKey(instancesWord, givenName, "port")), conf.GetString(joinKey(instancesWord, givenName, "address")), ""
				if (port >= 4000 && port < 4264) && address != "localhost" { // is an automatically allocated port
					portStr = ""
				} else {
					portStr = toSprintf(":%d", port)
				}
				details["accessAddress"] = conf.GetString(joinKey(instancesWord, givenName, "protocol")) + "://" + address + portStr
				newConfig.Set(joinKey(instancesWord, givenName, "accessAddress"), details["accessAddress"])
				if !existingFile(workDir.Join(instancesWord, name, cliComposeFilename).String()) {
					extendedCompose := createExtendedCompose(details, workDir.Join(instancesWord, name, chemotionComposeFilename).String())
					// write out the extended compose file
					if _, err, _ := gotoFolder(givenName), extendedCompose.WriteConfigAs(cliComposeFilename), gotoFolder("work.dir"); err == nil {
						zboth.Info().Msgf("Written extended file %s in the above step.", cliComposeFilename)
					} else {
						zboth.Fatal().Err(err).Msgf("Failed to write the extended compose file to its repective folder. This is necessary for future use.")
					}
				}
			}
			oldConfigPath := pathlib.NewPath(conf.ConfigFileUsed())
			if errWrite := newConfig.WriteConfigAs("new." + defaultConfigFilepath); errWrite == nil {
				zboth.Debug().Msgf("New configuration file  `%s`.", "new."+defaultConfigFilepath)
				if errRenameOld := oldConfigPath.RenameStr(toSprintf("old.%s.%s", time.Now().Format("060102150405"), defaultConfigFilepath)); errRenameOld == nil {
					zboth.Debug().Msgf("Renamed old configuration file to %s", oldConfigPath.String())
					if errRenameNew := workDir.Join("new." + defaultConfigFilepath).RenameStr(conf.ConfigFileUsed()); errRenameNew == nil {
						zboth.Info().Msgf("Successfully written new configuration file at %s.", conf.ConfigFileUsed())
						oldConfigPath.Remove()
						zboth.Info().Msgf("Upgrade was successful. Please restart this tool.")
						os.Exit(0)
					} else {
						zboth.Fatal().Err(errRenameNew).Msgf("Failed to rename the new configuration file. It is available at: %s", "new."+defaultConfigFilepath)
					}
				} else {
					zboth.Fatal().Err(errRenameOld).Msgf("Failed to rename existing configuration file. New one is available at: %s", "new."+defaultConfigFilepath)
				}
			} else {
				zboth.Fatal().Err(errWrite).Msgf("Failed to write the new configuration file. Old one is still available at %s for use with version 0.1.x of this tool.", oldConfigPath.String())
			}
		}
	}
	return
}
