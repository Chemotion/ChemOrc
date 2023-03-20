package cli

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"

	vercompare "github.com/hashicorp/go-version"
)

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
	gotoFolder("workdir")
	return
}

// confirm that virtualizer is the required minimum version
func confirmVirtualizer(minimum string) {
	if ver, err := execShell(toLower(virtualizer) + " --version"); err == nil {
		version, errConvert := vercompare.NewVersion(strings.TrimPrefix(strings.Split(string(ver), ",")[0], "Docker version "))
		if errConvert == nil {
			if errCompare := compareSoftwareVersion(minimum, version.String()); errCompare == nil {
				zboth.Debug().Msgf("Running version %s of %s", version.String(), virtualizer)
			} else {
				zboth.Fatal().Err(err).Msgf("%s is out of date. Please update it before proceeding.", virtualizer)
			}
		} else {
			zboth.Fatal().Err(toError("failed to convert version string")).Msgf("Failed to understand the following output upon executing `%s --version`: %s", toLower(virtualizer), ver)
		}
	} else {
		if err.Error() == "exit status 1" && runtime.GOOS == "linux" {
			zboth.Fatal().Err(toError("%s on WSL not running", virtualizer)).Msgf("%s is not running in your WSL environment. Hint: Turn on WSL integration setting in %s Desktop.", virtualizer, virtualizer)
		} else if err.Error() == "exit status 127" { // 127 is software not found
			zboth.Fatal().Err(toError("%s not found", virtualizer)).Msgf("%s is necessary to run %s", virtualizer, nameCLI)
		} else {
			zboth.Fatal().Err(err).Msgf(err.Error())
		}
	}
}

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

// call to virtualizer (this must not end in fatal error, i.e. must return `var success bool`)
func callVirtualizer(args string) (success bool) {
	if strings.Contains(args, "busybox") || strings.Contains(args, "mikefarah/yq") {
		zboth.Debug().Msgf("%s will now start the execution with command `%s %s` sent to shell.", nameCLI, toLower(virtualizer), args)
	} else {
		zboth.Info().Msgf("%s will now start the execution with command `%s %s` sent to shell.", nameCLI, toLower(virtualizer), args)
	}
	args = strings.TrimSpace(args)
	commandArgs := strings.Split(args, " ")
	commandExec := exec.Command(toLower(virtualizer), commandArgs...)
	// see https://blog.kowalczyk.info/article/wOYk/advanced-command-execution-in-go-with-osexec.html#:~:text=Capture%20output%20but%20also%20show%20progress%20%233
	var stdoutBuf, stderrBuf bytes.Buffer
	commandExec.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
	commandExec.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)
	if err := commandExec.Run(); err == nil {
		success = true
	} else {
		success = false
		zboth.Warn().Err(err).Msgf("%s command failed! Check log. ABORT!", virtualizer)
	}
	// TODO-v2: make this more elegent by using a virtual terminal, see https://github.com/creack/pty
	return
}
