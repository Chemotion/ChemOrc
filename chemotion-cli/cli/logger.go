package cli

import (
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
	if elementInSlice("--debug", &os.Args) > 0 {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		currentState.debug = true
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		currentState.debug = false
	}
	if elementInSlice("-q", &os.Args) > 0 || elementInSlice("--quiet", &os.Args) > 0 {
		currentState.quiet = true
	} else {
		currentState.quiet = false
	}
	// start logging
	if logFile, err := workDir.Join(logFilename).OpenFile(os.O_APPEND | os.O_CREATE | os.O_WRONLY); err == nil {
		zlog = zerolog.New(logFile).With().Timestamp().Logger()
		if currentState.quiet {
			zboth = zlog // in this case, both the loggers point to the same file and there should be no console output
		} else {
			console := zerolog.ConsoleWriter{Out: os.Stdout}
			console.FormatErrFieldName = func(i interface{}) string { return "" }  // we don't want error to be shown in the console
			console.FormatErrFieldValue = func(i interface{}) string { return "" } // PartsExclude doesn't seem to work!
			multi := zerolog.MultiLevelWriter(logFile, console)
			zboth = zerolog.New(multi).With().Timestamp().Logger()
		}
		zlog.Debug().Msgf("%s started. Successfully initialized logging", nameCLI)
		logPlatform()
	} else {
		minimalConsoleWriter := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()
		minimalConsoleWriter.Fatal().Err(err).Msg("Can't write log file. ABORT!") // minimal console writer
	}
}

// helpers: logging
func logPlatform() {
	if currentState.isInside {
		if currentState.name == "" {
			zlog.Debug().Msgf("Running inside an unknown container") // TODO: read .version file or get from environment
		} else {
			zlog.Debug().Msgf("Running inside `%s`", currentState.name)
		}
	} else {
		if currentState.name == "" {
			zlog.Debug().Msgf("Running on host machine; no instance selected yet")
		} else {
			zlog.Debug().Msgf("Running on host machine; selected instance: %s", currentState.name)
		}
	}
}

// debug level logging of where we are running at the moment
func logWhere() {
	logPlatform()
	zlog.Debug().Msgf("Called as: %s", strings.Join(os.Args, " "))
}
