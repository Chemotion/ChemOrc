package cli

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"

	"github.com/chigopher/pathlib"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	dc "github.com/docker/docker/client"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func dropIntoConsole(givenName string, consoleName string) {
	commandExec := exec.Command(toLower(virtualizer), []string{"compose", "exec", "eln", "chemotion", consoleName}...)
	commandExec.Stdin, commandExec.Stdout, commandExec.Stderr = os.Stdin, os.Stdout, os.Stderr
	switch consoleName { // use proper name when printing to user
	case "shell":
		consoleName = "shell"
	case "railsc":
		consoleName = "Rails console"
	case "psql":
		consoleName = "postgreSQL console"
	case "resetAdminPW":
		consoleName = "reset password console"
	}
	status := instanceStatus(givenName)
	if status == "Up" {
		zboth.Info().Msgf("Starting %s for instance `%s`.", consoleName, givenName)
		if _, err, _ := gotoFolder(givenName), commandExec.Run(), gotoFolder("workdir"); err == nil {
			zboth.Debug().Msgf("Successfuly closed %s for `%s`.", consoleName, givenName)
		} else {
			zboth.Fatal().Err(err).Msgf("%s ended with exit message: %s.", consoleName, err.Error())
		}
	} else {
		zboth.Warn().Err(toError("instance is %s", status)).Msgf("Cannot start a %s for `%s`. Instance is not running.", consoleName, givenName)
	}
}

var shellConsoleInstanceRootCmd = &cobra.Command{
	Use:     "shell",
	Aliases: []string{"bash"},
	Short:   "Drop into a shell (bash) console",
	Args:    cobra.NoArgs,
	Run: func(_ *cobra.Command, _ []string) {
		dropIntoConsole(currentInstance, "shell")
	},
}

var railsConsoleInstanceRootCmd = &cobra.Command{
	Use:     "rails",
	Aliases: []string{"ruby", "railsc"},
	Short:   "Drop into a Ruby on Rails console",
	Args:    cobra.NoArgs,
	Run: func(_ *cobra.Command, _ []string) {
		dropIntoConsole(currentInstance, "railsc")
	},
}

var psqlConsoleInstanceRootCmd = &cobra.Command{
	Use:     "psql",
	Aliases: []string{"postgresql", "sql", "postgres", "PostgreSQL"},
	Short:   "Drop into a PostgreSQL console",
	Args:    cobra.NoArgs,
	Run: func(_ *cobra.Command, _ []string) {
		dropIntoConsole(currentInstance, "psql")
	},
}

func panicCheck(err error) {
	if err != nil {
		panic(err)
	}
}

func getContainerID_api(givenName string, service string) (containerId string) {
	ctx := context.Background()
	cli, err := dc.NewClientWithOpts(dc.FromEnv, dc.WithAPIVersionNegotiation())
	panicCheck(err)

	filters := filters.NewArgs()
	filters.Add("name", getInternalName(givenName)+"-"+service)

	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{All: true, Filters: filters})

	panicCheck(err)

	return containers[0].ID
}

func copyFilesInContainer(givenName string, service string, scrFilePath string, dstFilePath string) {
	ctx := context.Background()
	cli, err := dc.NewClientWithOpts(dc.FromEnv, dc.WithAPIVersionNegotiation())
	panicCheck(err)

	status := instanceStatus(givenName)
	if status == "Up" {
		containerId := getContainerID_api(givenName, "eln")

		zboth.Info().Msgf("copying file: %s inside container (service: %s) with container id: %s at location: %s", scrFilePath, service, containerId, dstFilePath)

		file, err := os.Open(scrFilePath)
		panicCheck(err)

		err = cli.CopyToContainer(ctx, containerId, dstFilePath, bufio.NewReader(file), types.CopyToContainerOptions{
			AllowOverwriteDirWithFile: true,
		})

		if err != nil {
			zboth.Fatal().Err(err).Msgf("cannot copy %s: to the container %s", file.Name(), containerId)
		}
	}
}

// execute a script on a running container.
func executeScript(containerID string, pathToScript string) {
	// Download a file and copy it inside running container
	script := downloadFile("https://raw.githubusercontent.com/mehmood86/chemotion/release-121/eln/embed/scripts/resetAdminPW.sh", "resetAdminPW.sh")
	script_tar := getNewUniqueID() + ".tar"

	cmd := exec.Command("tar", "-cf", script_tar, script.String())
	cmd.Stdout = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	log.Printf("path: %s", cmd)
	err := cmd.Run()
	if err != nil {
		log.Fatalf("failed to call cmd.Run(): %v", err)
	}

	script.Remove()

	copyFilesInContainer(
		currentInstance,
		"eln",      //target service
		script_tar, //source
		"/script")  // destination

	pathlib.NewPath(script_tar).Remove()

	fmt.Println("Pleae type in the Password or press Enter for default password")
	passwd, err := term.ReadPassword(0)

	if err != nil {
		fmt.Println("An error occured while reading input. Please try again", err)
		return
	}

	if string(passwd) == "" {
		fmt.Println("setting default password [chemotion]")
	}

	ctx := context.Background()
	cli, err := dc.NewClientWithOpts(dc.FromEnv, dc.WithAPIVersionNegotiation())
	panicCheck(err)

	cmdStatementExecuteScript := []string{"bash", pathToScript, string(passwd)}
	optionsCreateExecuteScript := types.ExecConfig{
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          cmdStatementExecuteScript,
	}

	rst_ExecuteScript, err := cli.ContainerExecCreate(ctx, containerID, optionsCreateExecuteScript)
	panicCheck(err)

	response_ExecuteScript, err := cli.ContainerExecAttach(ctx, rst_ExecuteScript.ID, types.ExecStartCheck{})
	panicCheck(err)

	defer response_ExecuteScript.Close()
	data1, _ := io.ReadAll(response_ExecuteScript.Reader)
	fmt.Println(string(data1))
	panicCheck(err)
}

var resetPasswordConsoleInstanceRootCmd = &cobra.Command{
	Use:   "reset admin user password",
	Short: "reset admin user password inside ELN service " + nameCLI,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, _ []string) {
		// password reset for ADM
		containerID := getContainerID_api(currentInstance, "eln")
		executeScript(containerID, "/script/resetAdminPW.sh")
	},
}

func init() {
	consoleInstanceRootCmd.AddCommand(shellConsoleInstanceRootCmd)
	consoleInstanceRootCmd.AddCommand(railsConsoleInstanceRootCmd)
	consoleInstanceRootCmd.AddCommand(psqlConsoleInstanceRootCmd)
	consoleInstanceRootCmd.AddCommand(resetPasswordConsoleInstanceRootCmd)
}
