package cli

import (
	"fmt"

	"github.com/rs/zerolog"
)

// Viper is used to load values from config file. Cobra is the basis of our command line interface.
// This function uses Viper to set flags on Cobra.
// (See how cool this sounds, make sure you pick fun project names!)
func initConf() {
	zlog.Debug().Msg("Start initConf()")
	// check status of the config flag
	// if changed and specified file is not found then exit
	// otherwise set the value of configFile on viper
	// then use the path as determined by viper and set as the value of configFile
	if rootCmd.Flag("config-file").Changed && !existingFile(configFile) {
		// here configFile should be same as rootCmd.Flag("config-file").Value.String()
		zboth.Fatal().Err(fmt.Errorf("specified config file not found")).Msgf("Please ensure that the file you specify using --config/-f flag does exist.")
	}
	conf.SetConfigFile(configFile)
	configFile = conf.ConfigFileUsed()
	zlog.Debug().Msg("Attempting to read configuration file")
	// if the flag is not changed, check for the posibility of first run
	if existingFile(configFile) {
		firstRun = false
		// Try and read the configuration file, then unmarshal it
		if err := conf.ReadInConfig(); err == nil {
			if errUnmarshal := conf.UnmarshalKey(selector_key, &currentState.name); errUnmarshal != nil {
				zboth.Fatal().Err(fmt.Errorf("unmarshal failed")).Msgf("Failed to find the mandatory key %s in the file: %s.", selector_key, configFile)
			}
			if !conf.IsSet(joinKey("instances", currentState.name)) {
				zboth.Fatal().Err(fmt.Errorf("unmarshal failed")).Msgf("Failed to find the description for instance `%s` in the file: %s.", currentState.name, configFile)
			}
			if errUnmarshal := conf.UnmarshalKey(joinKey("instances", currentState.name), &currentState); err == nil {
				zboth.Info().Msgf("Read configuration file: %s.", configFile)
				if currentState.debug {
					zerolog.SetGlobalLevel(zerolog.DebugLevel) // escalate the debug level if said so by the config file
				} // don't do else because flags have the final say!
			} else {
				zboth.Fatal().Err(errUnmarshal).Msg("Failed to map values from configuration file. ABORT!")
			}
		} else {
			zboth.Fatal().Err(err).Msgf("Failed to read configuration file: %s. ABORT!", configFile)
		}
	}
	zlog.Debug().Msgf("End: initConf(), Config found?: %t, is Inside?: %t", !firstRun, currentState.isInside)
}
