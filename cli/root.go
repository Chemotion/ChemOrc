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
	"os"
	"strings"

	"github.com/chigopher/pathlib"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/exp/maps"
)

const (
	nameProject                     = "Chemotion ELN"
	nameCLI                         = "ChemCLI"
	versionCLI                      = "0.2.0-pre1"
	versionConfig                   = "2.0"
	logFilename                     = "chem_cli.log"
	defaultConfigFilepath           = "chem_cli.yml"
	chemotionComposeFilename        = "docker-compose.yml"
	cliComposeFilename              = "docker-compose.cli.yml"
	stateWord                       = "cli_state"
	selectorWord                    = "selected"  // key that is expected in the configFile to figure out the selected instance
	instancesWord                   = "instances" // the folder/key in which chemotion expects to find all the instances
	virtualizer                     = "docker"
	addressDefault                  = "http://localhost"
	shell                           = "bash"     // should work with linux (ubuntu, windows < WSL runs when running in powershell >, and macOS)
	minimumVirtualizer              = "20.10.10" // so as to support docker compose files version 3.9 and forcing Docker Desktop >= 4
	maxInstancesOfKind              = 63
	firstPort                uint64 = 4000
	repositoryGH                    = "https://github.com/harivyasi/ChemCLI"
	releaseURL                      = repositoryGH + "/releases/latest"
	composeURL                      = releaseURL + "/download/docker-compose.yml"
	backupshURL                     = releaseURL + "/download/backup.sh"
	rollNum                         = 1 // the default index number assigned by virtualizer to every container
	primaryService                  = "eln"
)

// configuration and logging
var (
	// currently selected instance
	currentInstance string
	// switches to true when this file is found in root of a computer
	isInContainer bool = existingFile("/.version")
	// stores the configuration of the CLI
	conf viper.Viper = *viper.New()
	// off-screen logger, initialized in initLog()
	zlog zerolog.Logger
	// off-screen + on-screen logger, initialized in initLog()
	zboth zerolog.Logger
	// path of the working directory: it is expected that all files and folders are relative to this path.
	// at the moment this cannot be changed; in future, we might make it customizable, so that the user can specify this.
	workDir pathlib.Path = *pathlib.NewPath(".")
	// how the executable was called
	commandForCLI string = os.Args[0]
	// call for the compose file -- it calls two file together
	composeCall = toSprintf("compose -f %s -f %s ", chemotionComposeFilename, cliComposeFilename) // extra space at end is on purpose
)

// data type that maps a string to corresponding cobra command
type cmdTable map[string]func(*cobra.Command, []string)

var rootCmdTable = make(cmdTable)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:       toSprintf("%s", commandForCLI),
	Short:     "CLI for Chemotion ELN",
	Long:      "Chemotion ELN is an Electronic Lab Notebook solution.\nDeveloped for researchers, the software aims to work for you.\nSee, https://www.chemotion.net.",
	Version:   versionCLI,
	Args:      cobra.NoArgs,
	ValidArgs: maps.Keys(rootCmdTable),
	// The following lines are the action associated with a bare application run i.e. without any arguments
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if zerolog.SetGlobalLevel(zerolog.InfoLevel); conf.GetBool(joinKey(stateWord, "debug")) {
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
			zboth.Debug().Msgf("Backing up current configuration to disk as chem_cli.debug.yml if possible")
			_ = conf.WriteConfigAs(workDir.Join("chem_cli.debug.yml").String())
			logwhere()
		}
		confirmVirtualizer(minimumVirtualizer)
		zboth.Info().Msgf("Welcome to %s! You are on a host machine.", nameCLI)
		if currentInstance != "" {
			if err := instanceValidate(currentInstance); err == nil {
				zboth.Info().Msgf("The instance you are currently managing is %s%s%s%s.", string("\033[31m"), string("\033[1m"), currentInstance, string("\033[0m"))
			} else {
				zboth.Fatal().Err(err).Msgf(err.Error())
			}
		}
		if strings.Contains(versionCLI, "pre") {
			zboth.Warn().Msgf("This is a pre-release version. Do not use this in production.")
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		isInteractive(true)
		var acceptedOpts []string
		if currentInstance == "" {
			acceptedOpts = append(acceptedOpts, "install - "+nameProject)
			rootCmdTable["install - "+nameProject] = newInstanceRootCmd.Run
		} else {
			status := instanceStatus(currentInstance)
			if status == "Up" {
				acceptedOpts = []string{"off", "restart"}
				rootCmdTable["off"] = offRootCmd.Run
				rootCmdTable["restart"] = restartRootCmd.Run
			} else if status == "Exited" || status == "Created" {
				acceptedOpts = []string{"on"}
				rootCmdTable["on"] = onRootCmd.Run
			} else {
				acceptedOpts = []string{"on", "off", "restart"}
				rootCmdTable["on"] = onRootCmd.Run
				rootCmdTable["off"] = offRootCmd.Run
				rootCmdTable["restart"] = restartRootCmd.Run
			}
			rootCmdTable["instance"] = instanceRootCmd.Run
			acceptedOpts = append(acceptedOpts, "instance")
		}
		acceptedOpts = append(acceptedOpts, []string{"advanced", "exit"}...)
		rootCmdTable["advanced"] = advancedRootCmd.Run
		rootCmdTable[selectOpt(acceptedOpts, "")](cmd, args)
	},
}

// This is called by main.main(). It only needs to happen once.
func Execute() {
	if err := rootCmd.Execute(); err == nil {
		zlog.Debug().Msgf("%s exited gracefully", nameCLI)
	} else {
		zboth.Fatal().Err(err).Msgf("%s exited abruptly, check log file if necessary. ABORT!", nameCLI)
	}
}

func init() {
	initLog()                               // initialize logging
	initFlags()                             // initialize flags
	cobra.OnInitialize(initConf, bindFlags) // intitialize configuration // bind the flag
	rootCmd.SetVersionTemplate(toSprintf("%s version %s\n", nameCLI, versionCLI))
}
