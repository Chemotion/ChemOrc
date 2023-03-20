package cli

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/cavaliergopher/grab/v3"
	"github.com/chigopher/pathlib"
	"github.com/google/uuid"
	vercompare "github.com/hashicorp/go-version"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/viper"
)

var versionSuffix string = " --version"

// check if file exists, and is a file (keep it simple, runs before logging starts!
func existingFile(filePath string) (exists bool) {
	exists, _ = pathlib.NewPath(filePath).IsFile()
	return
}

// check if the CLI is running interactively; if not, then exit. Wrapper around currentState.quiet.
func confirmInteractive() {
	if currentState.quiet {
		zboth.Fatal().Err(fmt.Errorf("incomplete in quiet mode")).Msgf("%s is in quiet mode. Give all arguments to specify the desired action; use '--help' flag for more. ABORT!", nameCLI)
	}
	if currentState.isInside {
		zboth.Fatal().Err(fmt.Errorf("inside container in interactive mode")).Msgf("%s CLI is not meant to executed interactively from within a container. Use the `-q` flag. ABORT!", nameCLI)
	}
}

func confirmInstalled() {
	confirmVirtualizer(minimumVirtualizer)
	if firstRun {
		// Println output so that user is not discouraged by a FATAL error on-screen... especially when beginning with the tool.
		msg := fmt.Sprintf("Please install %s by running `%s` before using it.", nameCLI, "chemotion install")
		fmt.Println(msg)
		zlog.Fatal().Err(fmt.Errorf("chemotion not installed")).Msgf(msg)
	}
}

// check if an element is in an array of type(element), if yes, return the 1st index, else -1.
func elementInSlice[T int | float64 | string](elem T, slice *[]T) int {
	for index, element := range *slice {
		if element == elem {
			return index
		}
	}
	return -1
}

// execute a command in shell
func execShell(command string) (result []byte, err error) {
	if result, err = exec.Command(shell, "-c", command).CombinedOutput(); err == nil {
		zlog.Debug().Msgf("Sucessfully executed shell command: %s", command)
	} else if !strings.HasSuffix(command, versionSuffix) {
		zboth.Warn().Err(err).Msgf("Failed execution of command: %s", command)
	}
	return
}

// find version of a given software (using its command)
func findVersion(software string) (version string) {
	ver, err := execShell(software + versionSuffix)
	version = strings.TrimSpace(strings.Split(strings.TrimPrefix(strings.TrimPrefix(string(ver), "v"), "Docker version "), ",")[0]) // TODO: Regexify!
	if err != nil {
		zboth.Warn().Err(err).Msgf("Version determination of %s failed", software)
		if virtualizer == "Docker" && err.Error() == "exit status 1" && runtime.GOOS == "linux" {
			version = "Docker on WSL not running!"
		} else if err.Error() == "exit status 127" {
			version = "Unknown / not installed or found!" // 127 is software not found
		} else {
			version = err.Error()
		}
	}
	return
}

// confirm a minimum version for a given software
func compareSoftwareVersion(min string, software string) error {
	var minimum, current *vercompare.Version
	if ver, err := vercompare.NewVersion(findVersion(software)); err == nil {
		current = ver
	} else {
		return err
	}
	if ver, err := vercompare.NewVersion(min); err == nil {
		minimum = ver
	} else {
		return err
	}
	if current.LessThan(minimum) {
		return fmt.Errorf("current version of %s: %s is less than the minimum required: %s", software, current.String(), minimum.String())
	}
	return nil
}

// generate a new UID (of the form xxxxxxxx) as a string
func getNewUniqueID() string {
	id, _ := uuid.NewRandom()
	return strings.Split(id.String(), "-")[0]
}

// download a file, filepath is respective to current working directory
func downloadFile(fileURL string, downloadLocation string) (filepath pathlib.Path) {
	if resp, err := grab.Get(downloadLocation, fileURL); err == nil {
		zboth.Info().Msgf("Downloaded file saved as: %s", resp.Filename)
		filepath = *pathlib.NewPath(resp.Filename)
	} else {
		zboth.Fatal().Err(err).Msgf("Failed to download file from: %s. Check log. ABORT!", fileURL)
	}
	return
}

// show (and then remove) a progress bar that waits on screen for given seconds to lapse
func waitProgressBar(seconds int, message []string) {
	bar := progressbar.NewOptions(seconds,
		progressbar.OptionSetDescription(strings.Join(message, " ")+"..."),
		progressbar.OptionSetPredictTime(false),
		progressbar.OptionClearOnFinish(),
	)
	for i := 0; i < seconds; i++ {
		time.Sleep(1 * time.Second)
		bar.Add(1)
	}
	bar.Finish()
}

// copy a text file
// func copyTextFile(source *pathlib.Path, target *pathlib.Path) (err error) {
// 	fmt.Println(source.String(), target.String())
// 	if reader, errRead := source.ReadFile(); err == nil {
// 		if errWrite := target.WriteFile(reader); err != nil {
// 			err = errWrite
// 		}
// 	} else {
// 		err = errRead
// 	}
// 	return
// }

// to manage config files as loaded into Viper
func getSubHeadings(configuration *viper.Viper, key string) (subheadings []string) {
	for k := range configuration.GetStringMapString(key) {
		subheadings = append(subheadings, k)
	}
	return
}

// join keys so as to access them in a viper configuration
func joinKey(s ...string) (result string) {
	result = strings.Join(s, ".")
	return
}

// to lower case, same as strings.ToLower
var toLower = strings.ToLower

// to get all existing instances as determined by the configuration file
func allInstances() (instances []string) {
	instances = getSubHeadings(&conf, "instances")
	return
}

// to get all existing used ports
func allPorts() (ports []int) {
	existingInstances := allInstances()
	for _, instance := range existingInstances {
		ports = append(ports, int(conf.GetUint32(joinKey("instances", instance, "port"))))
	}
	return
}

// get internal name for an instance
func getInternalName(givenName string) (name string) {
	name = conf.GetString(joinKey("instances", givenName, "name"))
	if name == "" {
		zboth.Fatal().Err(fmt.Errorf("instance not found")).Msgf("No such instance: %s", givenName)
	}
	return
}

// get column associated with `ps` output for a given instance of chemotion
func getColumn(givenName, column string) (values []string) {
	name := getInternalName(givenName)
	if res, err := execShell(fmt.Sprintf("%s ps -a --filter \"label=net.chemotion.cli.project=%s\" --format \"{{.%s}}\"", toLower(virtualizer), name, column)); err == nil {
		values = strings.Split(string(res), "\n")
	} else {
		values = []string{}
	}
	return
}

// get services associated with a given `instance` of Chemotion
func getServices(givenName string) (services []string) {
	name, out := getInternalName(givenName), getColumn(givenName, "Names")
	for _, line := range out { // determine what are the status messages for all associated containers
		l := strings.TrimSpace(line) // use only the first word
		if len(l) > 0 {
			l = strings.TrimPrefix(l, fmt.Sprintf("%s-", name))
			l = strings.TrimSuffix(l, fmt.Sprintf("-%d", rollNum))
			services = append(services, l)
		}
	}
	return
}

// change directory with logging
func gotoFolder(givenName string) (pwd string) {
	var folder string
	if givenName == "workdir" {
		folder = "../.."
	} else {
		folder = workDir.Join(instancesFolder, getInternalName(givenName)).String()
	}
	if err := os.Chdir(folder); err == nil {
		pwd, _ = os.Getwd()
		zboth.Debug().Msgf("Changed working directory to: %s", pwd)
	} else {
		zboth.Fatal().Msgf("Failed to changed working directory as required.")
	}
	return
}

// split address into subcomponents
func splitAddress(full string) (protocol string, address string, port int) {
	if err := addressValidate(full); err != nil {
		zboth.Fatal().Err(err).Msgf("Given address %s is invalid.", full)
	}
	protocol, address, _ = strings.Cut(full, ":")
	address = strings.TrimPrefix(address, "//")
	address, portStr, _ := strings.Cut(address, ":")
	if port = -1; portStr != "" {
		port, _ = strconv.Atoi(portStr)
	}
	return
}

// TODO

// Start shell for user
// var shellSystemRootCmd = &cobra.Command{
// 	Use:        "shell",
// 	SuggestFor: []string{"she"},
// 	Args:       cobra.NoArgs,
// 	Run: func(cmd *cobra.Command, args []string) {
// 		logWhere()
// 		confirmInstalled()
// 		fmt.Println("We are now going to start shell")
// 	},
// }

// Start a rails shell for user
// var railsSystemRootCmd = &cobra.Command{
// 	Use:        "rails",
// 	SuggestFor: []string{"rai"},
// 	Args:       cobra.NoArgs,
// 	Run: func(cmd *cobra.Command, args []string) {
// 		logWhere()
// 		confirmInstalled()
// 		fmt.Println("We are now going to start Rails shell")
// 	},
// }

// example starter

// var uninstallAdvancedRootCmd = &cobra.Command{
// 	Use:   "uninstall",
// 	Args:  cobra.NoArgs,
// 	Short: fmt.Sprintf("Uninstall %s completely.", nameCLI),
// 	Run: func(cmd *cobra.Command, args []string) {
// 		logWhere()
// 		confirmInstalled()
// 		confirmInteractive()
// 	},
// }
