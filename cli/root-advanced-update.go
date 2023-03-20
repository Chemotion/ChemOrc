package cli

import (
	"net/http"
	"strings"
	"time"

	"github.com/chigopher/pathlib"
	vercompare "github.com/hashicorp/go-version"
	"github.com/spf13/cobra"
)

func getLatestReleaseURL() (url string, err error) {
	client := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	if resp, errGet := client.Get(releaseUnresolvedURL); errGet == nil {
		if loc, errLoc := resp.Location(); errLoc == nil {
			url = loc.String()
		} else {
			err = errLoc
		}
	} else {
		err = errGet
	}
	return strings.Replace(url, "tag", "download", -1), err
}

func getLatestVersion() (version string) {
	if url, err := getLatestReleaseURL(); err == nil {
		urlInParts := strings.Split(url, "/")
		version = urlInParts[len(urlInParts)-1]
		zboth.Debug().Msgf("Latest version of CLI is %s, installed version is %s.", version, versionCLI)
	} else {
		zboth.Fatal().Err(err).Msgf("Could not resolve the version of latest release.")
	}
	return
}

func updateRequired() (required bool) {
	verKey, timeKey := joinKey(stateWord, "version"), joinKey(stateWord, "version_checked_on")
	// update version in conf if required
	confVersion := conf.GetString(verKey)
	if confVersion == "" || confVersion != versionCLI {
		conf.Set(verKey, versionCLI)
		rewriteConfig()
	}
	checkedOn := conf.GetTime(timeKey)
	if checkedOn.IsZero() || (time.Since(checkedOn).Hours() > 24) { // check every 24 hours
		existingVer, _ := vercompare.NewVersion(versionCLI)
		newVer, _ := vercompare.NewVersion(getLatestVersion())
		required = newVer.GreaterThan(existingVer)
		conf.Set(timeKey, time.Now())
		rewriteConfig()
	}
	return
}

func selfUpdate() {
	var url string
	if u, err := getLatestReleaseURL(); err == nil {
		url = u
	} else {
		zboth.Fatal().Err(err).Msgf("Could not determine address of latest executable.")
	}
	oldVersion := pathlib.NewPath(commandForCLI)
	stat, _ := oldVersion.Stat()
	cliFileName := oldVersion.Name()
	url = toSprintf("%s/%s", url, cliFileName)
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

var updateSelfAdvancedRootCmd = &cobra.Command{
	Use:   "update",
	Short: "Update this tool itself",
	Run: func(cmd *cobra.Command, _ []string) {
		update := updateRequired()
		if !update && ownCall(cmd) {
			update = toBool(cmd.Flag("force").Value.String())
		}
		if update {
			selfUpdate()
		} else {
			zboth.Info().Msgf("You are already on the latest version of %s CLI tool.", nameCLI)
		}
	},
}

func init() {
	advancedRootCmd.AddCommand(updateSelfAdvancedRootCmd)
	updateSelfAdvancedRootCmd.Flags().Bool("force", false, toSprintf("Force update the %s CLI.", nameCLI))
}
