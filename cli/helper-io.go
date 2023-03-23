package cli

import (
	"os"
	"os/exec"
	"strings"

	"github.com/cavaliergopher/grab/v3"
	"github.com/chigopher/pathlib"
)

// debug level logging of where we are running at the moment
func logwhere() {
	if isInContainer {
		if currentInstance == "" {
			zboth.Debug().Msgf("Running inside an unknown container") // TODO: read .version file or get from environment
		} else {
			zboth.Debug().Msgf("Running inside `%s`", currentInstance)
		}
	} else {
		if currentInstance == "" {
			zboth.Debug().Msgf("Running on host machine; no instance selected yet")
		} else {
			zboth.Debug().Msgf("Running on host machine; selected instance: %s", currentInstance)
		}
	}
	zboth.Debug().Msgf("Called as: %s", strings.Join(os.Args, " "))
}

// to write the configuration file
func writeConfig(firstRun bool) (err error) {
	var (
		keysToPreserve = []string{joinKey(stateWord, "quiet"), joinKey(stateWord, "debug")}
		preserve       = make(map[string]any)
	)
	if firstRun {
		conf.Set("version", versionConfig)
		conf.Set(joinKey(stateWord, selectorWord), currentInstance)
		conf.Set(joinKey(stateWord, "quiet"), false)
		conf.Set(joinKey(stateWord, "debug"), false)
		conf.Set(joinKey(stateWord, "version"), versionCLI)
	} else { // backup values
		oldConf := parseCompose(conf.ConfigFileUsed())
		for _, key := range keysToPreserve {
			preserve[key] = conf.GetBool(key)   // backup key into memory
			conf.Set(key, oldConf.GetBool(key)) // set conf's key to what is read from existing file
		}
	}
	// write to file
	if err = conf.WriteConfig(); err == nil {
		zboth.Debug().Msgf("Modified configuration file `%s`.", conf.ConfigFileUsed())
	} else {
		zboth.Warn().Err(err).Msgf("Failed to update the configuration file.")
	}
	if !firstRun { // restore values in conf from memory
		for _, key := range keysToPreserve {
			conf.Set(key, preserve[key])
		}
	}
	return
}

// check if file exists, and is a file (keep it simple, runs before logging starts!
func existingFile(filePath string) (exists bool) {
	exists, _ = pathlib.NewPath(filePath).IsFile()
	return
}

// download a file, filepath is respective to current working directory
func downloadFile(fileURL string, downloadLocation string) (filepath pathlib.Path) {
	zboth.Debug().Msgf("Trying to download %s to %s", fileURL, downloadLocation)
	if resp, err := grab.Get(downloadLocation, fileURL); err == nil {
		zboth.Debug().Msgf("Downloaded file saved as: %s", resp.Filename)
		filepath = *pathlib.NewPath(resp.Filename)
	} else {
		zboth.Fatal().Err(err).Msgf("Failed to download file from: %s. Check log. ABORT!", fileURL)
	}
	return
}

// to copy a file
func copyfile(source, destination string) (err error) {
	file := *pathlib.NewPath(source)
	var read []byte
	read, err = file.ReadFile()
	if err == nil {
		err = pathlib.NewPath(destination).WriteFile(read)
	}
	return
}

// change directory with logging
func gotoFolder(givenName string) (pwd string) {
	var folder string
	if givenName == "workdir" {
		folder = "../.."
	} else {
		folder = workDir.Join(instancesWord, getInternalName(givenName)).String()
	}
	if err := os.Chdir(folder); err == nil {
		pwd, _ = os.Getwd()
		zboth.Debug().Msgf("Changed working directory to: %s", pwd)
	} else {
		zboth.Fatal().Msgf("Failed to changed working directory as required.")
	}
	return
}

// execute a command in shell
func execShell(command string) (result []byte, err error) {
	if result, err = exec.Command(shell, "-c", command).CombinedOutput(); err == nil {
		zboth.Debug().Msgf("Sucessfully executed shell command: %s", command)
	} else {
		zboth.Warn().Err(err).Msgf("Failed execution of command: %s", command)
	}
	return
}

// to be called from the folder where file exists
func changeKey(filename string, key string, value string) (err error) {
	var where string
	where, err = os.Getwd()
	if err == nil {
		if existingFile(filename) {
			if success := callVirtualizer(toSprintf("run --rm -v %s:/workdir mikefarah/yq eval -i .%s=\"%s\" %s", where, key, value, filename)); !success {
				err = toError("failed to update %s in %s", key, filename)
				return
			}
		} else {
			err = toError("file %s not found", filename)
		}
	}
	return
}
