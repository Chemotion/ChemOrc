package cli

import (
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/cavaliergopher/grab/v3"
	"github.com/chigopher/pathlib"
	"github.com/mikefarah/yq/v4/pkg/yqlib"
	"github.com/spf13/viper"
)

// debug level logging of where we are running at the moment
func logwhere() {
	if currentInstance == "" {
		zboth.Debug().Msgf("Running; no instance selected yet")
	} else {
		zboth.Debug().Msgf("Running; selected instance: %s", currentInstance)
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
		conf.Set(joinKey(stateWord, "first_port"), 4000)
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
	grab.DefaultClient.UserAgent = toSprintf("%s/%s", nameCLI, versionCLI)
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
	if givenName == "work.dir" {
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

// split a compose port mapping such as "4000:4000", "0.0.0.0:4100:4000" or "4100:4000/tcp"
// into the part before the host port, the host port, and the part after it
func splitPortMapping(mapping string) (prefix, hostPort, suffix string, err error) {
	mapping = strings.TrimSpace(mapping)
	base, protocol, hasProtocol := strings.Cut(mapping, "/")
	lastColon := strings.LastIndex(base, ":")
	if lastColon == -1 {
		err = toError("port mapping %q does not publish a host port", mapping)
		return
	}
	hostSide, containerPort := base[:lastColon], base[lastColon+1:]
	if ip := strings.LastIndex(hostSide, ":"); ip != -1 {
		prefix, hostPort = hostSide[:ip+1], hostSide[ip+1:]
	} else {
		hostPort = hostSide
	}
	if _, errConv := strconv.ParseUint(hostPort, 10, 16); errConv != nil {
		err = toError("port mapping %q has an invalid host port %q", mapping, hostPort)
		return
	}
	suffix = ":" + containerPort
	if hasProtocol {
		suffix += "/" + protocol
	}
	return
}

// get the host port from a compose port mapping
func hostPortOf(mapping string) (port string, err error) {
	_, port, _, err = splitPortMapping(mapping)
	return
}

// to be called from the folder where file exists
// replaces the host port of the first port mapping of the primary service, whatever it currently is
func changeExposedPort(filename string, newPort string) (err error) {
	if existingFile(filename) {
		var content []byte
		content, err = pathlib.NewPath(filename).ReadFile()
		if err != nil {
			return err
		}
		key := "." + joinKey("services", primaryService, "ports[0]")
		var current string
		if current, err = evaluateYamlExpression(string(content), key); err != nil {
			return err
		}
		if current = strings.TrimSpace(current); current == "" || current == "null" {
			return toError("no port mapping found at %s in %s", key, filename)
		}
		var prefix, suffix string
		if prefix, _, suffix, err = splitPortMapping(current); err != nil {
			return err
		}
		expression := toSprintf("%s = %q", key, prefix+newPort+suffix)
		var output string
		output, err = evaluateYamlExpression(string(content), expression)
		if err == nil {
			yamlFile := pathlib.NewPath(filename)
			err = yamlFile.WriteFile([]byte(output))
		}
	} else {
		err = toError("file %s not found", filename)
	}
	return
}

func evaluateYamlExpression(inputYaml string, expression string) (string, error) {
	prefs := yqlib.NewDefaultYamlPreferences()
	encoder := yqlib.NewYamlEncoder(prefs)
	decoder := yqlib.NewGoccyYAMLDecoder()
	se := yqlib.NewStringEvaluator()
	return se.EvaluateAll(expression, inputYaml, encoder, decoder)
}
