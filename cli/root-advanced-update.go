package cli

import (
	"context"
	"strings"
	"time"

	"github.com/chigopher/pathlib"
	api_gh "github.com/google/go-github/v72/github"
	vercompare "github.com/hashicorp/go-version"
	"github.com/spf13/cobra"
)

// update the tool itself
func selfUpdate(version string) {
	oldVersion := pathlib.NewPath(commandForCLI)
	stat, _ := oldVersion.Stat()
	cliFileName := oldVersion.Name()
	url := toSprintf("%s/releases/download/%s/%s", repositoryGH, version, nameCLI)
	newVersion := downloadFile(url, workDir.Join(toSprintf("%s.new", cliFileName)).String())
	if err := newVersion.Chmod(stat.Mode() | 100); err != nil { // make sure that it remains executable for the ErrUseLastResponse
		zboth.Warn().Err(err).Msgf("Could not grant executable permission to the downloaded file. Please do it yourself.")
	}
	if errOld := oldVersion.RenameStr(toSprintf("%s.old", cliFileName)); errOld == nil {
		if errNew := newVersion.RenameStr(cliFileName); errNew == nil {
			zboth.Info().Msgf("Successfully downloaded the new version. Old version is available as %s and is safe to remove.", oldVersion.Name())
		} else {
			zboth.Warn().Err(errNew).Msgf("Successfully downloaded the new version. Please rename it to %s for further use. The old version is available as %s and is safe to remove.", cliFileName, oldVersion.Name())
		}
	} else {
		zboth.Warn().Err(errOld).Msgf("Successfully downloaded the new version but failed to rename the old one. The new version is called %s, please rename it %s. The old version is safe to remove.", newVersion.Name(), cliFileName)
	}
	conf.Set(joinKey(stateWord, "version"), version)
	if existingFile(conf.ConfigFileUsed()) {
		latestCompose := parseCompose(composeURL)
		_, latestVersion, _ := strings.Cut(latestCompose.GetString(joinKey("services", primaryService, "image")), "-")
		conf.Set(joinKey(stateWord, "latest_eln"), latestVersion)
		if err := writeConfig(false); err != nil {
			zboth.Warn().Err(err).Msgf("Failed to rewrite config file. You will also need to update the %s.version to %s in the %s file manually.", stateWord, version, conf.ConfigFileUsed())
		}
	}
}

// get the version string of the latest release
func getLatestVersion() (version string) {
	client := api_gh.NewClient(nil)
	location := strings.Split(repositoryGH, "/")
	if release, _, err := client.Repositories.GetLatestRelease(context.Background(), location[len(location)-2], location[len(location)-1]); err == nil {
		return *release.TagName
	} else {
		return versionCLI
	}
}

// check if an update is required; store time of check in the config file if it exists
func updateRequired(check bool) (required bool) {
	verKey, timeKey := joinKey(stateWord, "version"), joinKey(stateWord, "version_checked_on")
	if !check {
		if conf.IsSet(verKey) {
			if versionCLI != conf.GetString(verKey) {
				zboth.Warn().Msgf("%s version wrong in %s, be sure of what you are doing!", nameCLI, conf.ConfigFileUsed())
			}
			checkedOn := conf.GetTime(timeKey)
			if checkedOn.IsZero() { // in case this not set
				check = true
			} else {
				if time.Since(checkedOn).Hours() > 24 { // check every 24 hours
					check = true
				}
			}
		}
	}
	if check {
		existingVer, _ := vercompare.NewVersion(versionCLI)
		newVer, _ := vercompare.NewVersion(getLatestVersion())
		required = newVer.GreaterThan(existingVer)
		conf.Set(timeKey, time.Now())
		if required {
			latestCompose := parseCompose(composeURL)
			_, latestVersion, _ := strings.Cut(latestCompose.GetString(joinKey("services", primaryService, "image")), "-")
			conf.Set(joinKey(stateWord, "latest_eln"), latestVersion)
		}
		if existingFile(conf.ConfigFileUsed()) {
			writeConfig(false)
		}
	}
	return
}

var updateSelfAdvancedRootCmd = &cobra.Command{
	Use:   "update",
	Short: "Update " + nameCLI + " itself",
	Run: func(cmd *cobra.Command, _ []string) {
		if ownCall(cmd) && toBool(cmd.Flag("disable-autocheck").Value.String()) {
			if existingFile(conf.ConfigFileUsed()) {
				conf.Set(joinKey(stateWord, "version_checked_on"), time.Now().Add(time.Duration(876576)*time.Hour))
				writeConfig(false)

			} else {
				if currentInstance == "" {
					zboth.Info().Err(toError("no config file")).Msgf("This flag stores the settings in config file which is created only on installation. Please re-run the command after installation.")
				} else {
					zboth.Fatal().Err(toError("config file missing")).Msgf("Could not find config file %s", conf.ConfigFileUsed())
				}
			}
		} else {
			if selectYesNo("This process establishes contact with GitHub and gets data from them. Continue?", true) && (updateRequired(true) || (ownCall(cmd) && toBool(cmd.Flag("force").Value.String()))) {
				latestVersion := getLatestVersion()
				if selectYesNo(toSprintf("Update %s from version %s to version %s", nameCLI, versionCLI, latestVersion), true) {
					selfUpdate(latestVersion)
				}
			} else {
				zboth.Info().Msgf("You are on the latest version of %s.", nameCLI)
			}
		}
	},
}

func init() {
	advancedRootCmd.AddCommand(updateSelfAdvancedRootCmd)
	updateSelfAdvancedRootCmd.Flags().Bool("force", false, toSprintf("Force update the %s.", nameCLI))
	updateSelfAdvancedRootCmd.Flags().Bool("disable-autocheck", false, toSprintf("Disable auto-check for latest version of %s.", nameCLI))
}
