package cli

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// modify file system in container
func modifyContainer(givenName string, args []string) (success bool) {
	commandStr := fmt.Sprintf("run --rm -v %s:/mountedFolder --name chemotion-helper-safe-to-remove busybox %s", gotoFolder(givenName), args[0])
	zboth.Debug().Msgf("Executing `%s` in the container of %s", strings.Join(args, " "), givenName)
	for i := 1; i < len(args); i++ {
		commandStr += fmt.Sprintf(" /mountedFolder/%s", args[i])
	}
	success = callVirtualizer(commandStr)
	gotoFolder("workdir")
	return
}

// confirm that virtualizer is the required minimum version
func confirmVirtualizer(minimum string) {
	if version := findVersion(toLower(virtualizer)); version == "docker on WSL not running!" {
		zboth.Fatal().Err(fmt.Errorf(version)).Msgf("Docker is not running in your WSL environment. Hint: Turn on WSL integration setting in Docker Desktop.")
	} else if version == "Unknown / not installed or found!" {
		zboth.Fatal().Err(fmt.Errorf(version)).Msgf("%s is necessary to run %s", virtualizer, nameCLI)
	} else {
		zboth.Debug().Msgf("%s version %s is installed", virtualizer, version)
	}
	if err := compareSoftwareVersion(minimum, toLower(virtualizer)); err != nil {
		zboth.Fatal().Err(err).Msgf(err.Error())
	}
}

// call to virtualizer (this must not end in fatal error)
func callVirtualizer(args string) (success bool) {
	if strings.Contains(args, "busybox") {
		zboth.Debug().Msgf("%s will now fork the execution with command `%s %s` sent to shell.", nameCLI, toLower(virtualizer), args)
	} else {
		zboth.Info().Msgf("%s will now fork the execution with command `%s %s` sent to shell.", nameCLI, toLower(virtualizer), args)
	}
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
