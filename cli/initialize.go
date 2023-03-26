package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog"
)

// Initializes logging. Ignores values in the configuration as configuration is loaded after this initialization.
func initLog() {
	// lowest level reading of the debug and quiet flags
	// alas, it works only with command line flags, otherwise
	// we have to wait for the values to be read in from the config file
	// this low-level reading has to be done because logging begins before reading the config file.
	debug := elementInSlice("--debug", &os.Args) > 0
	quiet := elementInSlice("--quiet", &os.Args) > 0
	for _, arg := range os.Args[1:] { // only scan the arguments, not the calling-command itself
		if len(arg) > 1 && arg[0] == '-' && arg[0:2] != "--" {
			if strings.ContainsRune(arg[1:], 'q') {
				quiet = true
			}
			if strings.ContainsRune(arg[1:], 'd') {
				debug = true
			}
		}
	}
	// set debug level, depending on the flag
	if zerolog.SetGlobalLevel(zerolog.InfoLevel); debug {
		if !quiet {
			fmt.Printf("%s started in debug mode! ", nameCLI)
		}
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		if !quiet {
			fmt.Println("Logger set to Debug level!")
		}
	}
	// start logging
	if logFile, err := workDir.Join(logFilename).OpenFile(os.O_APPEND | os.O_CREATE | os.O_WRONLY); err == nil {
		zlog = zerolog.New(logFile).With().Timestamp().Logger()
		if quiet {
			zboth = zlog // in this case, both the loggers point to the same file and there should be no console output
		} else {
			console := zerolog.ConsoleWriter{Out: os.Stdout, FieldsExclude: []string{"error"}}
			multi := zerolog.MultiLevelWriter(logFile, console)
			zboth = zerolog.New(multi).With().Timestamp().Logger()
		}
		zboth.Debug().Msgf("%s started. Successfully initialized logging", nameCLI)
	} else {
		// Use a minimalistic console writer
		fmt.Printf("Can't write log file. ABORTING! Error was:\n%s\n", err)
	}
}

func initFlags() {
	// terminal always overrides config-file
	zboth.Debug().Msg("Start: initialize flags")
	// flag 1: instance, i.e. name of the instance to operate upon
	// default is empty string if no instance exists
	// default is then the last instance selected (by actively using the CLI)
	rootCmd.PersistentFlags().StringVarP(&currentInstance, "selected-instance", "i", "", toSprintf("select an existing instance of %s when starting %s", nameProject, nameCLI))
	// flag 2: quiet, i.e. should the CLI run in interactive mode
	// default is false
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, toSprintf("use %s in scripted mode i.e. without an interactive prompt", nameCLI))
	// flag 3: debug, i.e. should debug messages be logged
	// default is false
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "enable logging of debug messages")
	zboth.Debug().Msg("End: initialize flags")
}

// Viper is used to load values from config file. Cobra is the basis of our command line interface.
// This function uses Viper to set flags on Cobra.
// (See how cool this sounds, make sure you pick fun project names!)
func initConf() {
	zboth.Debug().Msg("Start: initialize configuration")
	zboth.Debug().Msg("Attempting to read configuration file")
	conf.SetConfigFile(defaultConfigFilepath)
	configFileFound := existingFile(conf.ConfigFileUsed())
	if configFileFound {
		// Try and read the configuration file, then unmarshal it
		if err := conf.ReadInConfig(); err == nil {
			switch conf.GetString("version") { // version of the YAML file
			case "1.0":
				zboth.Info().Msgf("Your current version of %s is not compatible with the existing configuration file.", nameCLI)
				if selectYesNo("Would you like to upgrade?", false) {
					if upgradeThisTool("0.1_to_0.2") {
						// successful upgrade exits the tool
					} else {
						zboth.Fatal().Err(toError("upgrade failed")).Msgf("Upgrade failed! Please contact Chemotion helpdesk.")
					}
				} else {
					zboth.Fatal().Err(toError("version", conf.GetString("version"), "of", conf.ConfigFileUsed(), "incompatible with", nameCLI, versionCLI)).Msgf("Please upgrade %s to continue. Or use %s version 0.1.", nameCLI, nameCLI)
				}
			case "2.0":
				if conf.IsSet(joinKey(stateWord, selectorWord)) {
					if currentInstance == "" { // i.e. the flag was not set
						if errUnmarshal := conf.UnmarshalKey(joinKey(stateWord, selectorWord), &currentInstance); errUnmarshal == nil {
							if ivErr := instanceValidate(currentInstance); ivErr != nil { // confirm that the specified current instance is described in the config file
								zboth.Fatal().Err(ivErr).Msgf("Failed to find the description for instance `%s` in the file: %s.", currentInstance, conf.ConfigFileUsed())
							}
						} else {
							zboth.Fatal().Err(errUnmarshal).Msgf("Failed to unmarshal the key %s in the file: %s.", joinKey(stateWord, selectorWord), conf.ConfigFileUsed())
						}
					}
				}
				if !conf.IsSet(joinKey(stateWord, "version")) {
					zboth.Fatal().Err(toError("unmarshal failed")).Msgf("Failed to find the mandatory key `%s` in the file: %s.", joinKey(stateWord, "version"), conf.ConfigFileUsed())
				}
			default:
				zboth.Fatal().Err(toError("config file version unsupported.")).Msgf("This version of the configuration file is unsupported. One likely fix is to update %s to the latest!", nameCLI)
			}
		} else {
			zboth.Fatal().Err(err).Msgf("Failed to read configuration file: %s. ABORT!", conf.ConfigFileUsed())
		}
	}
	zboth.Debug().Msgf("End: initialize configuration; Config found?: %t; is inside container?: %t", configFileFound, isInContainer)
}

// bind the command line flags to the configuration
func bindFlags() {
	zboth.Debug().Msg("Start: bind flags")
	if err := conf.BindPFlag(joinKey(stateWord, selectorWord), rootCmd.Flag("selected-instance")); err != nil {
		zboth.Warn().Err(err).Msgf("Failed to bind flag: %s. Will ignore command line input.", "selected-instance")
		if !existingFile(conf.ConfigFileUsed()) {
			conf.Set(joinKey(stateWord, selectorWord), currentInstance) // as a backup in case of failure
		}
	}
	for _, flag := range []string{"debug", "quiet"} {
		if err := conf.BindPFlag(joinKey(stateWord, flag), rootCmd.Flag(flag)); err != nil {
			zboth.Warn().Err(err).Msgf("Failed to bind flag: %s. Will ignore command line input.", flag)
			if !existingFile(conf.ConfigFileUsed()) {
				conf.Set(joinKey(stateWord, flag), false) // as a backup in case of failure
			}
			if flag == "debug" {
				zboth.Info().Msgf("Turning on Debug because flag binding failed")
				conf.Set(joinKey(stateWord, flag), true)
			}
		}
	}
	zboth.Debug().Msg("End: bind flags")
}
