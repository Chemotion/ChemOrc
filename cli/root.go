/*
Copyright © 2022 Peter Krauß, Shashank S. Harivyasi
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice,
   this list of conditions and the following disclaimer.

2. Redistributions in binary form must reproduce the above copyright notice,
   this list of conditions and the following disclaimer in the documentation
   and/or other materials provided with the distribution.

3. Neither the name of the copyright holder nor the names of its contributors
   may be used to endorse or promote products derived from this software
   without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
POSSIBILITY OF SUCH DAMAGE.
*/

package cli

import (
	"fmt"
	"os"

	"github.com/chigopher/pathlib"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	nameCLI                       = "Chemotion"
	versionCLI                    = "0.2.0-alpha"
	versionYAML                   = "1.1"
	insideFilename                = "/.version"
	defaultConfigFilepath         = "chemotion-cli.yml"
	defaultLogFilename            = "chemotion-cli.log"
	defaultComposeFilename        = "docker-compose.yml"
	extenedComposeFilename        = "docker-compose.cli.yml"
	instanceDefault               = "initial"
	addressDefault                = "http://localhost"
	stateWord                     = "cli_state"
	selectorWord                  = "selected"  // key that is expected in the configFile to figure out the selected instance
	instancesWord                 = "instances" // the folder/key in which chemotion expects to find all the instances
	virtualizer                   = "Docker"
	shell                         = "bash"    // should work with linux (ubuntu, windows < WSL runs when running in powershell >, and macOS)
	minimumVirtualizer            = "20.10.2" // so as to support docker compose files version 3.5 and avoid this: https://github.com/docker/for-mac/issues/4975 by forcing Docker Desktop >= 3.0.4
	maxInstancesOfKind            = 63
	firstPort              uint64 = 4000
	composeURL                    = "https://raw.githubusercontent.com/harivyasi/chemotion/chemotion-cli/docker-compose.yml"
	releaseUnresolvedURL          = "https://github.com/harivyasi/chemotion/releases/latest"
	rollNum                       = 1 // the default index number assigned by virtualizer to every container
	primaryService                = "eln"
)

// data type that maps a string to corresponding cobra command
type cmdTable map[string]func(*cobra.Command, []string)

var (
	// configuration and logging
	currentInstance string
	configFile      string
	firstRun        bool        = true                         // switches to false when configFile is found/given
	isInContainer   bool        = existingFile(insideFilename) // switches to true when insideFilename is found/given
	conf            viper.Viper = *viper.New()
	zlog            zerolog.Logger
	zboth           zerolog.Logger
	// path of the working directory
	workDir pathlib.Path = *pathlib.NewPath(".") // it is expected that all files and folders are relative to this path, unless specified otherwise by the user
	// how the executable was called
	commandForCLI string = os.Args[0]
	// others
	reseveredWords = []string{"instance", "advanced", "back", "exit"}
	composeCall    = toSprintf("compose -f %s -f %s ", defaultComposeFilename, extenedComposeFilename) // extra space at end is on purpose
)

var rootCmdTable = make(cmdTable)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     toSprintf("%s [command]", commandForCLI),
	Short:   "CLI for Chemotion ELN",
	Long:    "Chemotion ELN is an Electronic Lab Notebook solution.\nDeveloped for researchers, the software aims to work for you.\nSee, https://www.chemotion.net.",
	Version: versionCLI,
	Args:    cobra.NoArgs,
	// The following lines are the action associated with a bare application run i.e. without any arguments
	PersistentPreRun: func(cmd *cobra.Command, _ []string) {
		if zerolog.SetGlobalLevel(zerolog.InfoLevel); conf.GetBool(joinKey(stateWord, "debug")) {
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
			logwhere()
		}
		confirmVirtualizer(minimumVirtualizer)
		if firstRun && cmd.CalledAs() != "install" {
			// Println output so that user is not discouraged by a FATAL error on-screen... especially when beginning with the tool.
			msg := toSprintf("Please install %s by running `%s install` before using it.", nameCLI, commandForCLI)
			fmt.Println(msg)
			zlog.Fatal().Err(toError("chemotion not installed")).Msgf(msg) // zlog i.e. don't print on screen.
		}
		zboth.Info().Msgf("Welcome to %s! You are on a host machine.", nameCLI)
		if !firstRun {
			if updateRequired() {
				zboth.Info().Msgf("The version of %s - the CLI tool - you are using is outdated. Please update it by using `%s advanced update` command.", nameCLI, commandForCLI)
			}
			if cmd.Flag("selected-instance").Changed {
				if err := instanceValidate(cmd.Flag("selected-instance").Value.String()); err != nil {
					zboth.Fatal().Err(err).Msgf(err.Error())
				}
			}
			zboth.Info().Msgf("The instance you are currently managing is %s%s%s%s.", string("\033[31m"), string("\033[1m"), currentInstance, string("\033[0m"))
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		isInteractive(true)
		var acceptedOpts []string
		status := instanceStatus(currentInstance)
		if status == "Up" {
			acceptedOpts = []string{"off", "restart"}
			rootCmdTable["off"] = offRootCmd.Run
			rootCmdTable["restart"] = restartRootCmd.Run
		} else if status == "Created" || status == "Exited" {
			acceptedOpts = []string{"on"}
			rootCmdTable["on"] = onRootCmd.Run
		} else {
			acceptedOpts = []string{"on", "off", "restart"}
			rootCmdTable["on"] = onRootCmd.Run
			rootCmdTable["off"] = offRootCmd.Run
			rootCmdTable["restart"] = restartRootCmd.Run
		}
		acceptedOpts = append(acceptedOpts, []string{"instance", "advanced", "exit"}...)
		rootCmdTable["instance"] = instanceRootCmd.Run
		rootCmdTable["advanced"] = advancedRootCmd.Run
		rootCmdTable[selectOpt(acceptedOpts, "")](cmd, args)
	},
}

// This is called by main.main(). It only needs to happen once.
func Execute() {
	if err := rootCmd.Execute(); err == nil {
		zlog.Debug().Msgf("%s exited gracefully", nameCLI)
	} else {
		zboth.Fatal().Err(toError("unexplained")).Msgf("%s exited abruptly, check log file if necessary. ABORT!", nameCLI)
	}
}

func init() {
	initLog()                               // initialize logging
	initFlags()                             // initialize flags
	cobra.OnInitialize(initConf, bindFlags) // intitialize configuration // bind the flag
	rootCmd.SetVersionTemplate(fmt.Sprintln("Chemotion CLI version", versionCLI))
}
