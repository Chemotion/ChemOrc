package cli

import (
	"os"
	"os/exec"
	"runtime"
	"strings"

	vercompare "github.com/hashicorp/go-version"
)

// confirm a minimum version for a given software
func compareSoftwareVersion(required, current string) (err error) {
	var req, curr *vercompare.Version
	if req, err = vercompare.NewVersion(required); err == nil {
		if curr, err = vercompare.NewVersion(current); err == nil {
			if curr.LessThan(req) {
				return toError("current version: %s is less than the minimum required: %s", curr.String(), req.String())
			}
		}
	}
	return
}

// confirm that virtualizer is the required minimum version
func confirmVirtualizer(minimum string) {
	if ver, err := execShell(virtualizer + " --version"); err == nil {
		version, errConvert := vercompare.NewVersion(strings.TrimPrefix(strings.Split(string(ver), ",")[0], "Docker version "))
		if errConvert == nil {
			if errCompare := compareSoftwareVersion(minimum, version.String()); errCompare == nil {
				zboth.Debug().Msgf("Running version %s of %s", version.String(), virtualizer)
			} else {
				zboth.Fatal().Err(err).Msgf("%s is out of date. Please update it before proceeding.", virtualizer)
			}
		} else {
			zboth.Fatal().Err(toError("failed to convert version string")).Msgf("Failed to understand the following output upon executing `%s --version`: %s", virtualizer, ver)
		}
	} else {
		if err.Error() == "exit status 1" && runtime.GOOS == "linux" {
			zboth.Fatal().Err(toError("%s on WSL not running", virtualizer)).Msgf("%s is not running in your WSL environment. Hint: Turn on WSL integration setting in %s Desktop.", virtualizer, virtualizer)
		} else if err.Error() == "exit status 127" { // 127 is software not found
			zboth.Fatal().Err(toError("%s not found", virtualizer)).Msgf("%s is necessary to run %s using %s", virtualizer, nameProject, nameCLI)
		} else {
			zboth.Fatal().Err(err).Msg(err.Error())
		}
	}
}

// modify file system in container
func modifyContainer(givenName, command, source, target string) (success bool) {
	if target != "" {
		command = toSprintf("%s /mountedFolder/%s /mountedFolder/%s", command, source, target)
	} else {
		command = toSprintf("%s /mountedFolder/%s", command, source)
	}
	zboth.Debug().Msgf("Executing `%s` in the container of %s", command, givenName)
	commandStr := toSprintf("run --rm -v %s:/mountedFolder busybox %s", gotoFolder(givenName), command)
	success = callVirtualizer(commandStr)
	gotoFolder("work.dir")
	return
}

// call to virtualizer (this must not end in fatal error, i.e. must return `var success bool`)
func callVirtualizer(args string) (success bool) {
	if strings.Contains(args, "busybox") || strings.Contains(args, "mikefarah/yq") {
		zboth.Debug().Msgf("%s will now start the execution with command `%s %s` sent to shell.", nameProject, virtualizer, args)
	} else {
		zboth.Info().Msgf("%s will now start the execution with command `%s %s` sent to shell.", nameProject, virtualizer, args)
	}
	args = strings.TrimSpace(args)
	commandArgs := strings.Split(args, " ")
	commandExec := exec.Command(virtualizer, commandArgs...)
	commandExec.Stdin, commandExec.Stdout, commandExec.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := commandExec.Run(); err == nil {
		success = true
	} else {
		success = false
		zboth.Warn().Err(err).Msgf("%s command failed with error %s! ABORT!", virtualizer, err.Error())
	}
	return
}
