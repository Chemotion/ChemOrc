package cli

import (
	"net/http"
	"strings"
	"time"

	"github.com/chigopher/pathlib"
	vercompare "github.com/hashicorp/go-version"
	"github.com/spf13/cobra"
)

// update the tool itself
func selfUpdate() {
	oldVersion := pathlib.NewPath(commandForCLI)
	stat, _ := oldVersion.Stat()
	cliFileName := oldVersion.Name()
	url := releaseURL + "/download/" + cliFileName
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
}

// get the version string of the latest release
func getLatestVersion() (version string) {
	client := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	var url string
	if resp, errGet := client.Get(releaseURL); errGet == nil {
		if loc, errLoc := resp.Location(); errLoc == nil {
			url = loc.String()
			urlInParts := strings.Split(url, "/")
			version = urlInParts[len(urlInParts)-1]
			zboth.Debug().Msgf("Latest version of CLI is %s, installed version is %s.", version, versionCLI)
		} else {
			zboth.Fatal().Err(errLoc).Msgf("Could not resolve the version of latest release.")
		}
	} else {
		zboth.Fatal().Err(errGet).Msgf("Could not resolve the version of latest release.")
	}
	return
}

// check if an update is required; store time of check in the config file if it exists
func updateRequired() (required bool) {
	verKey, timeKey := joinKey(stateWord, "version"), joinKey(stateWord, "version_checked_on")
	if conf.IsSet(verKey) {
		checkedOn := conf.GetTime(timeKey)
		// check against version in file
		if versionCLI == conf.GetString(verKey) {
			if (time.Since(checkedOn).Hours() > 24) || checkedOn.IsZero() { // check every 24 hours
				return false
			}
		} else {
			zboth.Fatal().Msgf("%s version wrong in %s, stopping as a safety measure!", nameCLI, conf.ConfigFileUsed())
		}
	}
	existingVer, _ := vercompare.NewVersion(versionCLI)
	newVer, _ := vercompare.NewVersion(getLatestVersion())
	required = newVer.GreaterThan(existingVer)
	conf.Set(timeKey, time.Now())
	if existingFile(conf.ConfigFileUsed()) {
		writeConfig(false)
	}
	return
}

var updateSelfAdvancedRootCmd = &cobra.Command{
	Use:   "update",
	Short: "Update " + nameCLI + " itself",
	Run: func(cmd *cobra.Command, _ []string) {
		if selectYesNo("This process establishes contact with GitHub and gets data from them. Continue?", true) && (updateRequired() || (ownCall(cmd) && toBool(cmd.Flag("force").Value.String()))) {
			selfUpdate()
		} else {
			zboth.Info().Msgf("You are on the latest version of %s.", nameCLI)
		}
	},
}

func init() {
	advancedRootCmd.AddCommand(updateSelfAdvancedRootCmd)
	updateSelfAdvancedRootCmd.Flags().Bool("force", false, toSprintf("Force update the %s.", nameCLI))
}
