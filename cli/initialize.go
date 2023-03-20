package cli

import (
	"os"
	"time"

	"github.com/chigopher/pathlib"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

// Initializes logging. Ignores values in the configuration as configuration is loaded after this initialization.
func initLog() {
	// lowest level reading of the debug and quiet flags
	// alas, it works only with command line flags, otherwise
	// we have to wait for the values to be read in from the config file
	// this low-level reading has to be done because logging begins before reading the config file.
	if zerolog.SetGlobalLevel(zerolog.InfoLevel); elementInSlice("--debug", &os.Args) > 0 || elementInSlice("-d", &os.Args) > 0 || elementInSlice("-qd", &os.Args) > 0 || elementInSlice("-dq", &os.Args) > 0 {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
	// start logging
	if logFile, err := workDir.Join(logFilename).OpenFile(os.O_APPEND | os.O_CREATE | os.O_WRONLY); err == nil {
		zlog = zerolog.New(logFile).With().Timestamp().Logger()
		if elementInSlice("-q", &os.Args) > 0 || elementInSlice("--quiet", &os.Args) > 0 || elementInSlice("-qd", &os.Args) > 0 || elementInSlice("-dq", &os.Args) > 0 {
			zboth = zlog // in this case, both the loggers point to the same file and there should be no console output
		} else {
			console := zerolog.ConsoleWriter{Out: os.Stdout}
			console.FormatErrFieldName = func(_ any) string { return "" }  // we don't want error to be shown in the console
			console.FormatErrFieldValue = func(_ any) string { return "" } // PartsExclude doesn't seem to work!
			multi := zerolog.MultiLevelWriter(logFile, console)
			zboth = zerolog.New(multi).With().Timestamp().Logger()
		}
		zlog.Debug().Msgf("%s started. Successfully initialized logging", nameCLI)
	} else {
		minimalConsoleWriter := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()
		minimalConsoleWriter.Fatal().Err(err).Msg("Can't write log file. ABORT!") // minimal console writer
	}
}

func initFlags() {
	zlog.Debug().Msg("Start: initialize flags")
	// flag 1: instance, i.e. name of the instance to operate upon
	// terminal overrides config-file, default is read from the config file
	rootCmd.PersistentFlags().StringVarP(&currentInstance, "selected-instance", "i", "", toSprintf("select an existing instance of %s when starting", nameCLI))
	// flag 2: config, the configuration file
	// config as a flag cannot be read from the configuration file because that creates a circular dependency, default name is hard-coded
	rootCmd.PersistentFlags().StringVarP(&configFile, "config-file", "f", defaultConfigFilepath, "path to the configuration file")
	// flag 3: quiet, i.e. should the CLI run in interactive mode
	// terminal overrides config-file, default is false
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, toSprintf("use %s in scripted mode i.e. without an interactive prompt", commandForCLI))
	// flag 4: debug, i.e. should debug messages be logged
	// terminal overrides config-file, default is false
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "enable logging of debug messages")
	zlog.Debug().Msg("End: initialize flags")
}

func upgradeThisTool(transition string) (success bool) {
	switch transition {
	case "0.1_to_0.2":
		if success = selectYesNo("It seems you are upgrading from version 0.1.x of the tool to 0.2.x. Is this true?", true); success {
			newConfig := viper.New()
			newConfig.Set("version", versionYAML)
			newConfig.Set(joinKey(stateWord, selectorWord), conf.GetString(selectorWord))
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
				env := viper.New()
				env.SetConfigType("env")
				env.SetConfigFile(workDir.Join(instancesWord, name, "shared", "pullin", ".env").String())
				if err := env.ReadInConfig(); err == nil {
					newConfig.Set(joinKey(instancesWord, givenName, "environment"), env.AllSettings())
				} else {
					zboth.Warn().Err(err).Msgf("Failed to read the .env file, using existing information to create reasonable entries. Please check the created file manually!")
				}
				if env.IsSet("URL_HOST") && env.IsSet("URL_PROTOCOL") {
					newConfig.Set(joinKey(instancesWord, givenName, "accessaddress"), env.GetString("URL_PROTOCOL")+"://"+env.GetString("URL_HOST"))
				} else {
					newConfig.Set(joinKey(instancesWord, givenName, "accessaddress"), conf.GetString(joinKey(instancesWord, givenName, "protocol")+"://"+conf.GetString(joinKey(instancesWord, givenName, "address"))))
				}
				if !existingFile(workDir.Join(instancesWord, name, extenedComposeFilename).String()) {
					extendedCompose := createExtendedCompose(name, workDir.Join(instancesWord, name, defaultComposeFilename).String())
					// write out the extended compose file
					if _, err, _ := gotoFolder(givenName), extendedCompose.WriteConfigAs(extenedComposeFilename), gotoFolder("workdir"); err == nil {
						zboth.Info().Msgf("Written extended file %s in the above step.", extenedComposeFilename)
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
						success = true
					} else {
						zboth.Fatal().Err(errRenameNew).Msgf("Failed to rename the new configuration file. It is available at: %s", configFile+".new")
					}
				} else {
					zboth.Fatal().Err(errRenameOld).Msgf("Failed to rename existing configuration file. New one is available at: %s", configFile+".new")
				}
			} else {
				zboth.Fatal().Err(errWrite).Msgf("Failed to write the new configuration file. Old one is still available at %s for use with version 0.1.x of this tool.", oldConfigPath.String())
			}
		}
	}
	return
}

// Viper is used to load values from config file. Cobra is the basis of our command line interface.
// This function uses Viper to set flags on Cobra.
// (See how cool this sounds, make sure you pick fun project names!)
func initConf() {
	zlog.Debug().Msg("Start: initialize configuration")
	// check status of the config flag
	// if changed and specified file is not found then exit
	// otherwise set the value of configFile on viper
	// then use the path as determined by viper and set as the value of configFile
	if configFile != defaultConfigFilepath && !existingFile(configFile) { // the flag was set but file is missing
		zboth.Fatal().Err(toError("specified config file not found")).Msgf("Please ensure that the file you specify using --config/-f flag does exist.")
	}
	conf.SetConfigFile(configFile)
	zlog.Debug().Msg("Attempting to read configuration file")
	// if the flag is not changed, check for the posibility of first run
	if existingFile(configFile) {
		firstRun = false
		// Try and read the configuration file, then unmarshal it
		if err := conf.ReadInConfig(); err == nil {
			if conf.IsSet(joinKey(stateWord, selectorWord)) && conf.IsSet(instancesWord) {
				if currentInstance == "" { // i.e. the flag was not set
					if errUnmarshal := conf.UnmarshalKey(joinKey(stateWord, selectorWord), &currentInstance); errUnmarshal != nil {
						zboth.Fatal().Err(toError("unmarshal failed")).Msgf("Failed to unmarshal the mandatory key %s in the file: %s.", joinKey(stateWord, selectorWord), configFile)
					}
				}
				if !conf.IsSet(joinKey(instancesWord, currentInstance)) {
					zboth.Fatal().Err(toError("unmarshal failed")).Msgf("Failed to find the description for instance `%s` in the file: %s.", currentInstance, configFile)
				}
			} else {
				if conf.IsSet(joinKey(selectorWord)) {
					if upgradeThisTool("0.1_to_0.2") {
						zboth.Info().Msgf("Upgrade was successful. Please restart this tool.")
						os.Exit(0)
					}
				}
				zboth.Fatal().Err(toError("unmarshal failed")).Msgf("Failed to find the mandatory keys `%s`, `%s` and `%s` in the file: %s.", stateWord, selectorWord, instancesWord, configFile)
			}
		} else {
			zboth.Fatal().Err(err).Msgf("Failed to read configuration file: %s. ABORT!", configFile)
		}
	}
	zlog.Debug().Msgf("End: initialize configuration; Config found?: %t; is inside container?: %t", !firstRun, isInContainer)
}

// bind the command line flags to the configuration
func bindFlags() {
	zlog.Debug().Msg("Start: bind flags")
	if err := conf.BindPFlag(joinKey(stateWord, selectorWord), rootCmd.Flag("selected-instance")); err != nil {
		zboth.Warn().Err(err).Msgf("Failed to bind flag: %s. Will ignore command line input.", "selected-instance")
	}
	for _, flag := range []string{"debug", "quiet"} {
		if err := conf.BindPFlag(joinKey(stateWord, flag), rootCmd.Flag(flag)); err != nil {
			zboth.Warn().Err(err).Msgf("Failed to bind flag: %s. Will ignore command line input.", flag)
		}
		if !conf.IsSet(joinKey(stateWord, flag)) {
			conf.Set(joinKey(stateWord, flag), false)
		}
	}
	zlog.Debug().Msg("End: bind flags")
}
