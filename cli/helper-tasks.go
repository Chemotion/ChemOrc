package cli

import (
	"os"
	"time"

	"github.com/chigopher/pathlib"
	"github.com/spf13/viper"
)

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
