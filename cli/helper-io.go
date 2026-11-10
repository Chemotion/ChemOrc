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
		err = toError("port mapping %q does not publish a fixed host port (expected e.g. `4100:4000`)", mapping)
		return
	}
	hostSide, containerPort := base[:lastColon], base[lastColon+1:]
	if ip := strings.LastIndex(hostSide, ":"); ip != -1 {
		prefix, hostPort = hostSide[:ip+1], hostSide[ip+1:]
	} else {
		hostPort = hostSide
	}
	if hostPort == "" {
		err = toError("port mapping %q does not publish a fixed host port (expected e.g. `127.0.0.1:4100:4000`)", mapping)
		return
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

// get how the primary service of a compose file is published: the bind address prefix (e.g. "127.0.0.1:", or "")
// and the host port of its first port mapping. `set` is false when the compose file has no ports entry at all.
// Long-syntax entries (`- target: 4000`), which viper cannot read as strings, are an error rather than "not set".
func publishedExposure(compose *viper.Viper) (bindPrefix, hostPort string, set bool, err error) {
	key := joinKey("services", primaryService, "ports")
	if set = compose.IsSet(key); !set {
		return
	}
	mappings := compose.GetStringSlice(key)
	if len(mappings) == 0 || strings.TrimSpace(mappings[0]) == "" {
		err = toError("%s is not in the short form `[address:]host-port:container-port`", key)
		return
	}
	bindPrefix, hostPort, _, err = splitPortMapping(mappings[0])
	return
}

// to be called from the folder where file exists
// replaces the host port of the first port mapping of the primary service, keeping the file's own bind address
func changeExposedPort(filename string, newPort string) (err error) {
	return rewriteFirstPortMapping(filename, func(prefix string) (string, string) { return prefix, newPort })
}

// replaces both the bind address and the host port of the first port mapping of the primary service,
// so that an upgrade keeps the instance published exactly as before (e.g. only on 127.0.0.1)
func carryOverExposure(filename, bindPrefix, hostPort string) (err error) {
	return rewriteFirstPortMapping(filename, func(string) (string, string) { return bindPrefix, hostPort })
}

// rewrite the host side of the first port mapping of the primary service; `hostSide` gets the file's current bind
// address prefix and returns the prefix and host port to write; the container port and protocol are kept
func rewriteFirstPortMapping(filename string, hostSide func(prefix string) (newPrefix, newPort string)) (err error) {
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
		newPrefix, newPort := hostSide(prefix)
		expression := toSprintf("%s = %q", key, newPrefix+newPort+suffix)
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

// Evaluate a yq expression on a YAML document. Uses yq's standard YAML decoder (the one the yq command uses):
// the goccy decoder turned a scalar with a trailing comment (`cpus: 1 # …`) into `!!int '1 # …'`, which Docker
// Compose cannot parse, and dropped full-line comments.
func evaluateYamlExpression(inputYaml string, expression string) (string, error) {
	prefs := yqlib.NewDefaultYamlPreferences()
	encoder := yqlib.NewYamlEncoder(prefs)
	decoder := yqlib.NewYamlDecoder(prefs)
	se := yqlib.NewStringEvaluator()
	return se.EvaluateAll(expression, inputYaml, encoder, decoder)
}
