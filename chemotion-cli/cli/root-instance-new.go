package cli

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/chigopher/pathlib"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	_root_instance_new_name_        string
	_root_instance_new_use_         string
	_root_instance_new_development_ bool
	_root_instance_new_address_     string
	_root_instance_new_env_         string
)

// helper function to get a fresh (unassigned port)
func getFreshPort(kind string) (port int) {
	if firstRun {
		port = firstPort
	} else {
		existingPorts := allPorts()
		if kind == "Production" {
			for i := firstPort + 100; i <= maxInstancesOfKind+(firstPort+100); i++ {
				if elementInSlice(i, &existingPorts) == -1 {
					port = i
					break
				}
			}
		} else if kind == "Development" {
			for i := firstPort + 200; i <= maxInstancesOfKind+(firstPort+200); i++ {
				if elementInSlice(i, &existingPorts) == -1 {
					port = i
					break
				}
			}
		}
		if port == (firstPort+100)+maxInstancesOfKind || port == (firstPort+200)+maxInstancesOfKind {
			zboth.Fatal().Err(fmt.Errorf("max instances")).Msgf("A maximum of %d instances of %s are allowed. Please contact us if you hit this limit.", maxInstancesOfKind, nameCLI)
		}
	}
	return
}

func instanceCreate(givenName string, use string, kind string, givenAddress string) (success bool) {
	if err := newInstanceValidate(givenName); err != nil {
		zboth.Fatal().Err(err).Msgf("Given instance name is invalid because %s.", err.Error())
	}
	var (
		port     int
		protocol string
		address  string
		env      *viper.Viper
	)
	env = viper.New()
	env.SetConfigType("env")
	if _root_instance_new_env_ != "" {
		env.SetConfigFile(_root_instance_new_env_)
		if err := env.ReadInConfig(); err != nil {
			zboth.Fatal().Err(err).Msgf("Failed to parse the supplied .env file.")
		}
		if env.InConfig("URL_PROTOCOL") && env.InConfig("URL_HOST") {
			if givenAddress == addressDefault {
				givenAddress = env.GetString("URL_PROTOCOL") + "://" + env.GetString("URL_HOST")
			} else {
				zboth.Warn().Msgf("It seems you have `address` set in .env file as well as via the --address flag. The value in the given .env file will be overwritten.")
			}
		}
	}
	protocol, address, port = splitAddress(givenAddress)
	if address != "localhost" && (protocol == "http" && port == 443) || (protocol == "https" && port == 80) {
		zboth.Warn().Err(fmt.Errorf("port mismatch")).Msgf("You have chosen port %d for protocol %s. This is generally a very bad idea.", port, protocol)
		if !currentState.quiet {
			if !selectYesNo("Continue still", false) {
				zboth.Info().Msgf("Operation cancelled")
				os.Exit(2)
			}
		}
	}
	if port == -1 { // i.e. a port was not suggested by the user
		if address == "localhost" {
			port = getFreshPort(kind)
			givenAddress += ":" + strconv.Itoa(port)
		} else {
			if protocol == "http" {
				port = 80
			} else {
				port = 443
			}
		}
	} else {
		if address == "localhost" {
			zboth.Warn().Err(fmt.Errorf("localhost && port suggested")).Msgf("You suggested a port while running on localhost. We strongly recommend that you use the default schema i.e. do not assign a specific port.")
			if !currentState.quiet {
				if !selectYesNo("Continue still", false) {
					zboth.Info().Msgf("Operation cancelled")
					os.Exit(2)
				}
			}
		}
	}
	// create new unique name for the instance
	name := fmt.Sprintf("%s-%s", givenName, getNewUniqueID())
	// store values in the conf, the conf file is modified only later
	if firstRun {
		conf.SetConfigFile(workDir.Join(defaultConfigFilepath).String())
		conf.Set("version", versionYAML)
		conf.Set(selector_key, givenName)
	}
	conf.Set(joinKey("instances", givenName, "name"), name)
	conf.Set(joinKey("instances", givenName, "kind"), kind)
	conf.Set(joinKey("instances", givenName, "quiet"), false)
	conf.Set(joinKey("instances", givenName, "debug"), kind == "Development")
	conf.Set(joinKey("instances", givenName, "protocol"), protocol)
	conf.Set(joinKey("instances", givenName, "address"), address)
	conf.Set(joinKey("instances", givenName, "port"), port)
	// get the compose file for the instance
	var composeFilepath pathlib.Path // TODO: check on the version of the compose file
	var isUrl bool = false
	if existingFile(use) {
		composeFilepath = *pathlib.NewPath(use)
	} else if _, err := url.ParseRequestURI(use); err == nil {
		isUrl = true
		composeFilepath = downloadFile(use, workDir.String()) // downloads to the working directory
	} else {
		zboth.Fatal().Err(err).Msgf("Failed to parse the URL/file: %s.", use)
	}
	// parse the compose file
	compose.SetConfigFile(composeFilepath.String())
	if err := compose.ReadInConfig(); err == nil {
		if isUrl {
			composeFilepath.Remove()
		}
	} else {
		if isUrl {
			composeFilepath.Remove()
		}
		zboth.Fatal().Err(err).Msgf("Invalid formatting for a compose file.")
	}
	// set labels in the compose file
	sections := []string{"services", "volumes", "networks"}
	for _, section := range sections {
		subheadings, _ := getKeysValues(&compose, section) // subheadings are the names of the services, volumes and networks
		for _, k := range subheadings {
			compose.Set(joinKey(section, k, "labels"), map[string]string{"net.chemotion.cli.project": name})
		}
	}
	// set unique name for volumes in the compose file
	volumes, _ := getKeysValues(&compose, "volumes")
	for _, volume := range volumes {
		n := compose.GetString(joinKey("volumes", volume, "name"))
		compose.Set(joinKey("volumes", volume, "name"), name+"_"+n)
	}
	// set the port in the compose file
	compose.Set(joinKey("services", "eln", "ports"), []string{fmt.Sprintf("%d:4000", port)})
	zboth.Info().Msgf("Creating a new instance of %s called %s.", nameCLI, name)
	// make folder
	if err := workDir.Join(instancesFolder, name).MkdirAll(); err != nil {
		zboth.Fatal().Err(err).Msgf("Unable to create folder to store instances of %s.", nameCLI)
	}
	if _, err, _ := gotoFolder(givenName), compose.WriteConfigAs(composeFilename), gotoFolder("workdir"); err == nil {
		zboth.Info().Msgf("Written compose file %s in the above step.", compose.ConfigFileUsed())
		commandStr := fmt.Sprintf("compose -f %s up --no-start", composeFilename)
		zboth.Info().Msgf("Starting %s with command: %s", virtualizer, commandStr)
		if _, worked, _ := gotoFolder(givenName), callVirtualizer(commandStr), gotoFolder("workdir"); !worked {
			success = worked
			zboth.Fatal().Err(fmt.Errorf("%s failed", commandStr)).Msgf("Failed to setup %s. Check log. ABORT!", nameCLI)
		}
	} else {
		success = false
		zboth.Fatal().Err(err).Msgf("Failed to write the compose file to its repective folder. This is necessary for future use.")
	}
	// write env file into the container
	envFile := workDir.Join(instancesFolder, name, ".env")
	env.SetConfigFile(envFile.String())
	env.Set("URL_HOST", strings.TrimPrefix(givenAddress, protocol+"://"))
	env.Set("URL_PROTOCOL", protocol)
	if err := env.WriteConfig(); err == nil {
		modifyContainer(givenName, []string{"mkdir -p", "shared/pullin"})
		if worked := modifyContainer(givenName, []string{"cp", ".env", "shared/pullin"}); !worked {
			success = worked
			zboth.Warn().Msgf("Failed to write .env file in `%s/shared/pullin`", name)
		}
	} else {
		zboth.Warn().Err(err).Msgf("Failed to write .env file")
	}
	envFile.Remove()
	zboth.Info().Msgf("Successfully created container the container. New %s port available at %d.", nameCLI, port)
	// now modify the config file
	if err := conf.WriteConfig(); err == nil {
		zboth.Info().Msgf("Written config file: %s.", conf.ConfigFileUsed())
	} else {
		zboth.Fatal().Err(err).Msg("Failed to write config file. Check log. ABORT!")
	}
	return success
}

func newInstanceInteraction() (create bool) {
	create = selectYesNo("Installation process may download containers (of multiple GBs) and can take some time. Continue", true)
	if create {
		if _root_instance_new_name_ == instanceDefault { // i.e user has not changed it by passing an argument
			_root_instance_new_name_ = getString("Please enter the name of the instance you want to create", newInstanceValidate)
		}
		if _root_instance_new_env_ == "" && _root_instance_new_address_ == addressDefault && selectYesNo("Is this instance running on a web-server?", false) { // i.e user has not changed it by passing an argument
			_root_instance_new_address_ = getString("Please enter the web-address e.g. https://chemotion.uni.de:125", addressValidate)
		}
	} else {
		zboth.Info().Msgf("Installation cancelled.")
	}
	return
}

// command to install a new container of Chemotion
var newInstanceRootCmd = &cobra.Command{
	Use:   "new",
	Args:  cobra.NoArgs,
	Short: "Create a new instance of " + nameCLI,
	Run: func(cmd *cobra.Command, args []string) {
		logWhere()
		confirmInstalled()
		create := true
		kind := "Production"
		if _root_instance_new_development_ {
			kind = "Development"
		}
		if !currentState.quiet {
			confirmInteractive()
			create = newInstanceInteraction()
			if create && !_root_instance_new_development_ { // i.e. the flag was not set
				fmt.Println("What kind of instance do you want?")
				kind = selectOpt([]string{"Production", "Development"})
			}
		}
		if create {
			if success := instanceCreate(_root_instance_new_name_, _root_instance_new_use_, kind, _root_instance_new_address_); success {
				zboth.Info().Msg("Successfully created the new instance")
			}
		}
	},
}

func init() {
	instanceRootCmd.AddCommand(newInstanceRootCmd)
	newInstanceRootCmd.Flags().StringVar(&_root_instance_new_name_, "name", instanceDefault, "Name for the new instance")
	newInstanceRootCmd.Flags().StringVar(&_root_instance_new_use_, "use", composeURL, "URL or filepath to use for creating the instance")
	newInstanceRootCmd.Flags().StringVar(&_root_instance_new_address_, "address", addressDefault, "Web-address (or hostname) for accessing the instance")
	newInstanceRootCmd.Flags().StringVar(&_root_instance_new_env_, "env", "", ".env file for the new instance")
	newInstanceRootCmd.Flags().BoolVar(&_root_instance_new_development_, "development", false, "Create a development instance")
}
