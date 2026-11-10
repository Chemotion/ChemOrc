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

// oldest ELN version offered for new installations; older versions remain available as upgrade sources only
const oldestInstallableELN = "3.0.0"

// the versions to offer, from `orderVersions` (newest first): those newer than `currentVersion`, the current one
// marked as such, and none older than `minimumVersion` (no lower limit if it is empty). A `currentVersion` that
// cannot be read (e.g. an image tag like `eln-latest`) is treated as unknown: every version is offered.
func offeredVersions(currentVersion, minimumVersion string, orderVersions []string) (offered []string) {
	now, errNow := vercompare.NewVersion(currentVersion)
	if errNow != nil {
		zlog.Debug().Msgf("Current ELN version %q not understood, offering every version.", currentVersion)
		now = nil
	}
	var minimum *vercompare.Version
	if minimumVersion != "" {
		minimum, _ = vercompare.NewVersion(minimumVersion)
	}
	for _, version := range orderVersions {
		planned, err := vercompare.NewVersion(version)
		if err != nil {
			continue
		}
		if minimum != nil && planned.LessThan(minimum) {
			continue
		}
		if now == nil || planned.GreaterThan(now) {
			offered = append(offered, planned.String())
		} else if planned.Equal(now) {
			offered = append(offered, now.String()+" - current version")
		}
	}
	return
}

// Every ELN version this CLI offers, newest first, with the commit of this repository whose
// payload/docker-compose.yml shipped it. The first entry is the latest version known to this release.
// This is also the record of which compose file belonged to which ELN release: entries are only ever added,
// pinned to the commit that was released (see the git history of this file for when each was added).
var elnComposeFiles = []struct{ version, commit string }{
	{"4.0.1", "feat/eln-4.0.1"}, // TODO: pin to the version-0.2 commit with the final 4.0.1 payload, before the release
	{"3.1.2", "df94de05d9011322604c2273666d72461793cea2"},
	{"3.0.0", "666a9f9bf1c0a985942c653972adfeaccbad1b52"},
	{"2.2.0", "ece196c567e6505afee853f4acf7baedfdd0770e"},
	{"2.1.1", "8b8765e1279b9b4fbd2748245a3819abd67bae93"},
	{"1.10.5", "3a9339fe7156da32d786975482aa97c993a997b9"},
	{"1.9.3", "b7ad83fba1e1db6c5525a11b06bf7eed59a769f6"},
	{"1.9.2", "f30d8ae2dfdfb0160c3ceef30d5540e404c4f23a"},
	{"1.8.2", "7a62248a40416586e5d5e7d1a77adb6fe4f360fe"},
	{"1.8.1", "681cab48409d6c9a6be51ac5903e04890467d998"},
	{"1.8.0", "0664a5667d06da7898e83a11e31e4b471235daf1"},
	{"1.7.3", "b8bb1280a6e042b96b8d3e71d030709b113686bc"},
	{"1.7.2", "b07becc0f421a76aaee6f9e32ed6bf49f942258b"},
	{"1.6.2", "e577832edaba14fa21ee9aa9288e4b00052729c8"},
	{"1.5.4", "548ead617a552307f30d5051e72c01d95e99b30f"},
}

// URL of the compose file shipped with an ELN version, at the given commit (or ref) of this repository
func elnComposeURL(commit string) string {
	return "https://raw.githubusercontent.com/Chemotion/ChemOrc/" + commit + "/payload/docker-compose.yml"
}

// Mark or add the latest stable ELN version, as last read from the latest release's compose file
// (`cli_state.latest_eln`, which can be stale, e.g. after downloading a new chemCLI by hand):
//   - `newest` itself is only marked, where it is offered, and keeps its pinned compose file;
//   - a version newer than `newest` (this CLI's newest), i.e. released after this CLI, is offered first with the
//     latest release's compose file, unless the instance already runs it or something newer;
//   - anything else (older, or not a version) is ignored.
func withLatestELN(offered []string, versions map[string]string, newest, currentVersion, latest string) []string {
	const mark = " - latest stable version"
	if _, known := versions[latest]; known {
		if latest != newest {
			return offered // a known but older version: a stale value, not the latest
		}
		for i := range offered {
			if offered[i] == latest {
				offered[i] = latest + mark
			}
		}
		return offered
	}
	latestV, err := vercompare.NewVersion(latest)
	if err != nil {
		return offered
	}
	if newestV, err := vercompare.NewVersion(newest); err != nil || !latestV.GreaterThan(newestV) {
		return offered // older than what this CLI knows: a stale value
	}
	if now, err := vercompare.NewVersion(currentVersion); err == nil && !latestV.GreaterThan(now) {
		return offered // the instance already runs it (or something newer)
	}
	versions[latest] = composeURL
	return append([]string{latest + mark}, offered...)
}

// helper to determine the required compose file; `minimumVersion` limits the versions offered (empty: no limit)
func getComposeAddressToUse(currentVersion, minimumVersion, action string) (use string) {
	versions := make(map[string]string)
	orderVersions := make([]string, 0, len(elnComposeFiles))
	for _, e := range elnComposeFiles {
		versions[e.version] = elnComposeURL(e.commit)
		orderVersions = append(orderVersions, e.version)
	}
	validVersions := offeredVersions(currentVersion, minimumVersion, orderVersions)
	if conf.IsSet(joinKey(stateWord, "latest_eln")) {
		validVersions = withLatestELN(validVersions, versions, orderVersions[0], currentVersion, conf.GetString(joinKey(stateWord, "latest_eln")))
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
	compose, err = readYAML(composeFilepath.String())
	if err != nil {
		zboth.Fatal().Err(err).Msgf("Failed to read the file: %s", compose.ConfigFileUsed())
	}
	var imagesNotFound []string
	if pull {
		zboth.Info().Msgf("Pulling images for the services in the compose file %s", composeFilepath.Name())
		if success := callVirtualizer(toSprintf("compose -f %s pull", composeFilepath.String())); !success {
			zboth.Warn().Err(toError("pull failed")).Msgf("Failed to pull images for the services in the compose file %s", composeFilepath.Name())
			// check if we have all the images locally
			services := getSubHeadings(&compose, "services")
			for _, service := range services {
				if image := compose.GetString(joinKey("services", service, "image")); image != "" {
					// checks if the image exists locally
					if !callVirtualizer(toSprintf("image inspect %s", image)) {
						imagesNotFound = append(imagesNotFound, image)
					}
				}
			}

		}
	}
	if isUrl {
		if errRemove := composeFilepath.Remove(); errRemove != nil {
			zboth.Warn().Err(errRemove).Msgf("Failed to clean up temporary download file: %s", composeFilepath.String())
		}
	}
	if pull && len(imagesNotFound) > 0 {
		for _, image := range imagesNotFound {
			zboth.Fatal().Err(toError(toSprintf("image %s not found locally", image))).Msgf("Image %s not found locally. Please pull the image manually.", image)
		}
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
		// ports firstPort+101 up to (and excluding) firstPort+101+maxInstancesOfKind are available for allocation
		found := false
		for i := firstPort + 101; i < (firstPort+101)+maxInstancesOfKind; i++ {
			if elementInSlice(i, &existingPorts) == -1 {
				port = i
				found = true
				break
			}
		}
		if !found {
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
	otp := getNewOTP()
	envVars := []string{
		"PUBLIC_URL=" + details["accessAddress"],
		"SECRET_KEY_BASE=" + key,
		"OTP_SECRET_KEY=" + otp,
	}
	for _, service := range []string{primaryService, "worker", "executor"} {
		extendedCompose.Set(joinKey("services", service, "environment"), envVars)
	}
	if extendedCompose.IsSet(joinKey("services", "converter")) {
		extendedCompose.Set(joinKey("services", "converter", "environment"), []string{"SECRET_KEY=" + getNewUniqueID() + getNewUniqueID() + getNewUniqueID()})
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
		zboth.Fatal().Err(err).Msgf("Failed to write the extended compose file to its respective folder. This is necessary for future use.")
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
				details["givenName"] = getString("Please enter the name of the instance you want to create", []string{}, newInstanceValidate)
			}
			if askUse {
				details["use"] = getComposeAddressToUse("1.3.1", oldestInstallableELN, "install")
			}
			if askAddress {
				if selectYesNo("Will this instance have its own web-address?", true) {
					details["accessAddress"] = getString("Please enter the web-address (e.g. https://chemotion.uni.de or http://chemotion.uni.de:4100)", []string{}, addressValidate)
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
}
