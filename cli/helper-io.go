package cli

import (
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/cavaliergopher/grab/v3"
	"github.com/chigopher/pathlib"
	"github.com/spf13/viper"
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
		conf.Set(joinKey(stateWord, selectorWord), currentInstance)
		conf.Set(joinKey(stateWord, "quiet"), false)
		conf.Set(joinKey(stateWord, "debug"), false)
		conf.Set(joinKey(stateWord, "version"), versionCLI)
	} else { // backup values
		oldConf, err := readYAML(conf.ConfigFileUsed())
		if err != nil {
			zboth.Fatal().Err(err).Msgf("Failed to read in the existing config file, cannot overwrite it.")
		}
		for _, key := range keysToPreserve {
			preserve[key] = conf.GetBool(key)   // backup key into memory
			conf.Set(key, oldConf.GetBool(key)) // set conf's key to what is read from existing file
		}
	}
	conf.Set("version", versionConfig)
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
	var read []byte
	if read, err = pathlib.NewPath(source).ReadFile(); err == nil {
		err = pathlib.NewPath(destination).WriteFile(read)
	}
	return
}

// to read in a YAML file
func readYAML(filepath string) (yamlFile viper.Viper, err error) {
	// parse the YAML file
	yamlFile = *viper.New()
	yamlFile.SetConfigFile(filepath)
	err = yamlFile.ReadInConfig()
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
		zboth.Fatal().Err(err).Msgf("Failed to changed working directory as required.")
	}
	return
}

// determine shell to use
func determineShell() (shell string) {
	var err error
	if shell = os.Getenv("SHELL"); shell == "" {
		if runtime.GOOS == "windows" {
			if shell, err = exec.LookPath("pwsh.exe"); err == nil {
				return
			} else {
				if shell, err = exec.LookPath("powershell.exe"); err == nil {
					return
				}
			}
		} else {
			err = toError("$SHELL variable not set")
		}
		zboth.Fatal().Err(err).Msgf("Cannot run this tool. No compatible shell found.")
	}
	return
}

// execute a command in shell
func execShell(command string) (result []byte, err error) {
	if result, err = exec.Command(shell, "-c", command).CombinedOutput(); err == nil {
		zboth.Debug().Msgf("Sucessfully executed shell command: %s in shell: %s", command, shell)
	} else {
		zboth.Warn().Err(err).Msgf("Failed execution of command: %s in shell: %s", command, shell)
	}
	zlog.Debug().Msgf("Output of execution: %s", result) // output not shown on screen
	return
}

// to be called from the folder where file exists
func changeExposedPort(filename string, newPort string) (err error) {
	if existingFile(filename) {
		var result []byte
		if callVirtualizer("pull mikefarah/yq") { // get the latest version
			if result, err = execShell(toSprintf("cat %s | %s run -i --rm mikefarah/yq '.%s |= sub(\"%d:\", \"%s:\")'", filename, virtualizer, joinKey("services", "eln", "ports[0]"), firstPort, newPort)); err == nil {
				yamlFile := pathlib.NewPath(filename)
				err = yamlFile.WriteFile(result)
			}
		} else {
			zboth.Fatal().Err(toError("failed to pull `yq`")).Msgf("Failed to pull `yq` image.")
		}
	} else {
		err = toError("file %s not found", filename)
	}
	return
}
