package cli

import (
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/chigopher/pathlib"
	vercompare "github.com/hashicorp/go-version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// helper to determine the required compose file
func getComposeAddressToUse(currentVersion, action string) (use string) {
	versions := make(map[string]string)
	latestForThisCLIRelease := "2.1.1"
	orderVersions := []string{latestForThisCLIRelease, "1.10.5", "1.9.3"} // descending order
	versions[latestForThisCLIRelease] = "https://raw.githubusercontent.com/Chemotion/ChemOrc/8b8765e1279b9b4fbd2748245a3819abd67bae93/payload/docker-compose.yml"
	versions["1.10.5"] = "https://raw.githubusercontent.com/Chemotion/ChemOrc/3a9339fe7156da32d786975482aa97c993a997b9/payload/docker-compose.yml"
	versions["1.9.3"] = "https://raw.githubusercontent.com/Chemotion/ChemCLI/b7ad83fba1e1db6c5525a11b06bf7eed59a769f6/payload/docker-compose.yml"
	validVersions := []string{}
	for _, version := range orderVersions {
		now, _ := vercompare.NewVersion(currentVersion)
		planned, _ := vercompare.NewVersion(version)
		if planned.GreaterThan(now) {
			validVersions = append(validVersions, planned.String())
		} else if planned.Equal(now) {
			validVersions = append(validVersions, now.String()+" - current version")
		}
	}
	if conf.IsSet(joinKey(stateWord, "latest_eln")) {
		latest := conf.GetString(joinKey(stateWord, "latest_eln"))
		if latest == latestForThisCLIRelease {
			for i := range validVersions {
				if validVersions[i] == latest {
					validVersions[i] = latest + " - latest stable version"
				}
			}
		} else {
			validVersions = append([]string{latest + " - latest stable version"}, validVersions...)
			versions[latest] = composeURL
		}
	}
	if len(validVersions) != 0 {
		selected := selectOpt(validVersions, toSprintf("Select the version of ELN you want to %s", action))
		selected, _, _ = strings.Cut(selected, " -")
		use = versions[selected]
	} else {
		use = composeURL
	}
	return
}

// helper to get a compose file
func parseAndPullCompose(use string, pull bool) (compose viper.Viper) {
	var (
		composeFilepath pathlib.Path
		isUrl           bool
		err             error
	)
	if existingFile(use) {
		composeFilepath = *pathlib.NewPath(use)
	} else if _, err = url.ParseRequestURI(use); err == nil {
		isUrl = true
		composeFilepath = downloadFile(use, pathlib.NewPath(".").Join(toSprintf("%s.%s", getNewUniqueID(), chemotionComposeFilename)).String()) // downloads to where-ever it is called from
	} else {
		if isUrl {
			zboth.Fatal().Err(err).Msgf("Failed to download the file from URL: %s.", use)
		} else {
			zboth.Fatal().Err(err).Msgf("Failed: %s file not found.", use)
		}
	}
	if pull {
		if success := callVirtualizer(toSprintf("compose -f %s pull", composeFilepath.String())); !success {
			zboth.Warn().Err(toError("pull failed")).Msgf("Failed to pull images for the services in the compose file %s", composeFilepath.Name())
		}
	}
	compose, err = readYAML(composeFilepath.String())
	if isUrl {
		composeFilepath.Remove()
	}
	if err != nil {
		zboth.Fatal().Err(err).Msgf("Failed to read the file: %s", compose.ConfigFileUsed())
	}
	return
}

// helper to get a fresh (unassigned port)
func getFreshPort() (port uint64) {
	existingPorts := allPorts()
	var firstPort uint64
	if assigned := conf.GetInt(joinKey(stateWord, "first_port")); assigned == 0 {
		firstPort = 4000
		conf.Set(joinKey(stateWord, "first_port"), firstPort)
	} else {
		firstPort = uint64(assigned)
	}
	if len(existingPorts) == 0 {
		port = firstPort
	} else {
		for i := firstPort + 101; i <= maxInstancesOfKind+(firstPort+101); i++ {
			if elementInSlice(i, &existingPorts) == -1 {
				port = i
				break
			}
		}

		if port == (firstPort+101)+maxInstancesOfKind {
			zboth.Fatal().Err(toError("max instances")).Msgf("A maximum of %d instances of %s are allowed. Please contact us if you hit this limit.", maxInstancesOfKind, nameProject)
		}
	}
	return
}

func createExtendedCompose(details map[string]string, use string) (extendedCompose viper.Viper) {
	name := details["name"]
	extendedCompose = *viper.New()
	compose := parseAndPullCompose(use, false)
	extendedCompose.Set("name", name) // set project name for the virtulizer
	// create an additional service to run commands
	extendedCompose.Set(joinKey("services", "executor", "image"), compose.GetString(joinKey("services", primaryService, "image")))
	extendedCompose.Set(joinKey("services", "executor", "volumes"), compose.GetStringSlice(joinKey("services", primaryService, "volumes")))
	extendedCompose.Set(joinKey("services", "executor", "environment"), []string{toSprintf("CONFIG_ROLE=%s", primaryService), "RUBYOPT=-W0"})
	extendedCompose.Set(joinKey("services", "executor", "depends_on"), []string{"db"})
	extendedCompose.Set(joinKey("services", "executor", "networks"), []string{"chemotion"})
	extendedCompose.Set(joinKey("services", "executor", "profiles"), []string{"execution"})
	// set labels on services and volumes for future identification
	sections := []string{"services", "volumes"}
	for _, section := range sections {
		subheadings := getSubHeadings(&compose, section) // subheadings are the names of the services and volumes
		for _, k := range subheadings {
			extendedCompose.Set(joinKey(section, k, "labels"), []string{toSprintf("net.chemotion.cli.project=%s", name)})
		}
	}
	// set unique name for volumes in the compose file
	volumes := getSubHeadings(&compose, "volumes")
	for _, volume := range volumes {
		n := compose.GetString(joinKey("volumes", volume, "name"))
		if n == "" && volume == "spectra" {
			n = "chemotion_spectra"
		} // because the spectra volume has no name
		if strings.HasPrefix(n, name) { // for compatibility with upgradeThisTool("0.1_to_0.2")
			extendedCompose.Set(joinKey("volumes", volume, "name"), n)
		} else {
			extendedCompose.Set(joinKey("volumes", volume, "name"), name+"_"+n)
		}
	}
	key := getNewUniqueID() + getNewUniqueID() + getNewUniqueID()
	for _, service := range []string{"worker", "eln", "executor"} {
		extendedCompose.Set(toSprintf("services.%s.environment", service), []string{"PUBLIC_URL=" + details["accessAddress"], "SECRET_KEY_BASE=" + key})
	}
	if extendedCompose.IsSet("services.converter") {
		extendedCompose.Set("services.converter.environment", []string{"SECRET_KEY=" + getNewUniqueID() + getNewUniqueID() + getNewUniqueID()})
	}
	return
}

func instanceCreate(details map[string]string) (success bool) {
	pro, add, port := splitAddress(details["accessAddress"])
	details["protocol"], details["address"] = pro, add
	if port == 0 {
		port = getFreshPort()
		if details["address"] == "localhost" {
			details["accessAddress"] += toSprintf(":%d", port)
		}
	} else {
		if details["address"] == "localhost" {
			zboth.Warn().Err(toError("localhost && port suggested")).Msgf("You suggested a port while running on localhost. We strongly recommend that you use the default schema i.e. do not assign a specific port.")
			if isInteractive(false) {
				if !selectYesNo("Continue still", false) {
					zboth.Info().Msgf("Operation cancelled")
					os.Exit(2)
				}
			}
		}
	}
	details["port"] = strconv.FormatUint(port, 10)
	// download and modify the compose file
	var composeFile pathlib.Path
	if existingFile(details["use"]) {
		dest := workDir.Join(toSprintf("%s.%s", getNewUniqueID(), chemotionComposeFilename))
		if err := copyfile(details["use"], dest.String()); err == nil {
			composeFile = *dest
		} else {
			zboth.Fatal().Err(err).Msgf("Failed to copy the suggested compose file: %s. This is necessary for future use.", details["use"])
		}
	} else {
		composeFile = downloadFile(details["use"], workDir.Join(toSprintf("%s.%s", getNewUniqueID(), chemotionComposeFilename)).String())
	}
	if err := changeExposedPort(composeFile.String(), details["port"]); err != nil {
		composeFile.Remove()
		zboth.Fatal().Err(err).Msgf("Failed to update the downloaded compose file. This is necessary for future use.")
	}
	extendedCompose := createExtendedCompose(details, composeFile.String())
	// store values in the conf, the conf file is modified only later
	conf.Set(joinKey(instancesWord, details["givenName"], "port"), port)
	conf.Set(joinKey(instancesWord, details["givenName"], "kind"), "production") // because we do not want to change the format of chem_cli.yml file
	for _, key := range []string{"name", "accessAddress"} {
		conf.Set(joinKey(instancesWord, details["givenName"], key), details[key])
	}
	// make folder and move the compose file into it
	zboth.Info().Msgf("Creating a new instance of %s called %s.", nameCLI, details["name"])
	if err := workDir.Join(instancesWord, details["name"]).MkdirAll(); err == nil {
		composeFile.Rename(workDir.Join(instancesWord, details["name"], chemotionComposeFilename))
	} else {
		zboth.Fatal().Err(err).Msgf("Unable to create folder to store instances of %s.", nameProject)
	}
	// write out the extended compose file
	if _, err, _ := gotoFolder(details["givenName"]), extendedCompose.WriteConfigAs(cliComposeFilename), gotoFolder("work.dir"); err == nil {
		zboth.Info().Msgf("Written compose files %s and %s in the above steps.", chemotionComposeFilename, cliComposeFilename)
	} else {
		zboth.Fatal().Err(err).Msgf("Failed to write the extended compose file to its repective folder. This is necessary for future use.")
	}
	if _, success, _ = gotoFolder(details["givenName"]), callVirtualizer(composeCall+"up --no-start"), gotoFolder("work.dir"); !success {
		zboth.Fatal().Err(toError("compose up failed")).Msgf("Failed to setup an instance of %s. Check log. ABORT!", nameProject)
	}
	var firstRun bool
	if !existingFile(conf.ConfigFileUsed()) && currentInstance == "" {
		firstRun = true
		currentInstance = details["givenName"]
	}
	compose := viper.New()
	compose.SetConfigFile(composeFile.String())
	compose.ReadInConfig()
	conf.Set(joinKey(instancesWord, details["givenName"], "image"), compose.GetString(joinKey("services", "eln", "image")))
	if err := writeConfig(firstRun); err != nil {
		zboth.Fatal().Err(err).Msg("Failed to write config file. Check log. ABORT!") // we want a fatal error in this case, `rewriteConfig()` does a Warn error
	}
	return success
}

// interaction when creating a new instance
func processInstanceCreateCmd(cmd *cobra.Command, details map[string]string) (create bool) {
	askName, askAddress, askUse := true, true, true
	create = true
	details["accessAddress"] = addressDefault
	details["use"] = composeURL
	if ownCall(cmd) {
		if cmd.Flag("development").Changed {
			zboth.Fatal().Err(toError("deprecated flag")).Msgf("This flag is now deprecated. Please use the devcontainer functionality from Chemotion ELN repository to work with development instance.")
		}
		if cmd.Flag("name").Changed {
			details["givenName"] = cmd.Flag("name").Value.String()
			if err := newInstanceValidate(details["givenName"]); err != nil {
				zboth.Fatal().Err(err).Msgf("Cannot create new instance with name %s: %s", details["givenName"], err.Error())
			}
			askName = false
		} else {
			if !isInteractive(false) {
				zboth.Fatal().Err(toError("specify instance name")).Msgf("Instance must be specified using `-n` flag when in quiet mode.")
			}
		}
		if cmd.Flag("address").Changed {
			details["accessAddress"] = cmd.Flag("address").Value.String()
			if err := addressValidate(details["accessAddress"]); err != nil {
				zboth.Fatal().Err(err).Msgf("Cannot accept the address %s: %s", details["accessAddress"], err.Error())
			}
			askAddress = false
		}
		if cmd.Flag("use").Changed {
			details["use"] = cmd.Flag("use").Value.String()
			askUse = false
		}
	}
	if isInteractive(false) {
		if !ownCall(cmd) { // don't ask if the command is run directly i.e. without the menu
			{
				create = selectYesNo("Installation process may download containers (of multiple GBs) and can take some time. Continue", true)
			}
		}
		if create {
			if askName {
				details["givenName"] = getString("Please enter the name of the instance you want to create", newInstanceValidate)
			}
			if askUse {
				details["use"] = getComposeAddressToUse("1.3.1", "install")
			}
			if askAddress {
				if selectYesNo("Will this instance have its own web-address?", true) {
					details["accessAddress"] = getString("Please enter the web-address (e.g. https://chemotion.uni.de or http://chemotion.uni.de:4100)", addressValidate)
				}
			}
		}
	}
	// create new unique name for the instance
	if cmd.Flags().Lookup("suffix") != nil && cmd.Flag("suffix").Changed { // suffix exists only for restore flag
		suffix := cmd.Flag("suffix").Value.String()
		rec_len := len(getNewUniqueID())
		if len(suffix) != rec_len {
			zboth.Warn().Msgf("It is recommended that the length of the suffix is %d.", rec_len)
		}
		details["name"] = toSprintf("%s-%s", details["givenName"], suffix)
	} else {
		details["name"] = toSprintf("%s-%s", details["givenName"], getNewUniqueID())
	}
	return
}

// command to install a new instance of Chemotion
var newInstanceRootCmd = &cobra.Command{
	Use:   "new",
	Args:  cobra.NoArgs,
	Short: "Create a new instance of " + nameProject,
	Run: func(cmd *cobra.Command, _ []string) {
		details := make(map[string]string)
		create := processInstanceCreateCmd(cmd, details)
		if create {
			if success := instanceCreate(details); success {
				zboth.Info().Msgf("Successfully created a new instance. Once switched on, it can be found at: %s", details["accessAddress"])
			}
		}
	},
}

func init() {
	instanceRootCmd.AddCommand(newInstanceRootCmd)
	newInstanceRootCmd.Flags().StringP("name", "n", "must.be.given", "Name for the new instance")
	newInstanceRootCmd.Flags().String("use", composeURL, "URL or filepath of the compose file to use for creating the instance")
	newInstanceRootCmd.Flags().String("address", addressDefault, "Web-address (or hostname) for accessing the instance")
	newInstanceRootCmd.Flags().Bool("development", false, "NOW DEPRECATED Create a development instance")
}
