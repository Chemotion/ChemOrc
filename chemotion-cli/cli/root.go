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

	"github.com/chigopher/pathlib"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	versionCLI            = "0.1.2-alpha"
	versionYAML           = "1.0"
	nameCLI               = "Chemotion"
	defaultConfigFilepath = "chemotion-cli.yml"
	logFilename           = "chemotion-cli.log"
	instanceDefault       = "initial"
	addressDefault        = "http://localhost"
	selector_key          = "selected" // key that is expected in the configFile to figure out the selected instance
	stateFile             = "/.version"
	instancesFolder       = "instances" // the folder in which chemotion expects to find all the instances
	virtualizer           = "Docker"
	shell                 = "bash"    // should work with linux (ubuntu, windows < WSL runs when running in powershell >, and macOS)
	minimumVirtualizer    = "20.10.2" // so as to support docker compose files version 3.5 and avoid this: https://github.com/docker/for-mac/issues/4975 by forcing Docker Desktop >= 3.0.4
	composeFilename       = "docker-compose.yml"
	maxInstancesOfKind    = 64
	firstPort             = 4000
	composeURL            = "https://raw.githubusercontent.com/harivyasi/chemotion/chemotion-cli/docker-compose.yml"
	rollNum               = 1 // the default index number assigned by virtualizer to every container
)

var (
	// configuration
	currentState state
	configFile   string
	firstRun     bool        = true // switches to false when configFile is found/given
	conf         viper.Viper = *viper.New()
	compose      viper.Viper = *viper.New()
	// logging
	zlog  zerolog.Logger
	zboth zerolog.Logger
	// path of the working directory
	workDir pathlib.Path = *pathlib.NewPath(".") // it is expected that all files and folders are relative to this path, unless specified otherwise by the user
)

// struct to store information about the currently selected instance, which has implications for the current state of this tool
type state struct {
	debug    bool
	quiet    bool
	name     string
	isInside bool
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "chemotion {on|off|instance|advanced}",
	Short:   "CLI for Chemotion ELN",
	Long:    "Chemotion ELN is an Electronic Lab Notebook solution.\nDeveloped for researchers, the software aims to work for you.\nSee, https://www.chemotion.net.",
	Version: versionCLI,
	// The following lines are the action associated with a bare application run i.e. without any arguments
	Run: func(cmd *cobra.Command, args []string) {
		logWhere()
		confirmInstalled()
		confirmInteractive()
		fmt.Printf("Welcome to %s! You are on a host machine. The instance you are currently managing is %s%s%s%s.\n", nameCLI, string("\033[31m"), string("\033[1m"), currentState.name, string("\033[0m"))
		acceptedOpts := []string{"on", "off", "instance", "advanced", "exit"}
		selected := selectOpt(acceptedOpts)
		switch selected {
		case "on":
			onRootCmd.Run(cmd, args)
		case "off":
			offRootCmd.Run(cmd, args)
		case "instance":
			instanceRootCmd.Run(cmd, args)
		case "advanced":
			advancedRootCmd.Run(cmd, args)
		case "exit":
			zlog.Debug().Msg("Chose to exit")
		}
	},
}

// This is called by main.main(). It only needs to happen once.
func Execute() {
	if err := rootCmd.Execute(); err == nil {
		zlog.Debug().Msgf("%s exited gracefully", nameCLI)
	} else {
		zboth.Fatal().Err(fmt.Errorf("unexplained")).Msgf("%s exited abruptly, check log file if necessary. ABORT!", nameCLI)
	}
}

func init() {
	// flag 0: isInside, determined automatically whenever CLI runs
	currentState.isInside = existingFile(stateFile)
	// begin by setting up logging
	initLog() // in logger.go
	// initialize flags
	zlog.Debug().Msg("Start: init(): initialize flags")
	// flag 1: instance, i.e. name of the instance to operate upon
	// terminal overrides config-file, default is `default`
	rootCmd.PersistentFlags().StringVarP(&currentState.name, "selected-instance", "i", "", fmt.Sprintf("select an existing instance of %s when starting", nameCLI))
	// flag 2: config, the configuration file
	// config as a flag cannot be read from the configuration file because that creates a circular dependency, default name is hard-coded
	rootCmd.PersistentFlags().StringVarP(&configFile, "config-file", "f", defaultConfigFilepath, "path to the configuration file")
	// flag 3: quiet, i.e. should the CLI run in interactive mode
	// terminal overrides config-file, default is false
	rootCmd.PersistentFlags().BoolVarP(&currentState.quiet, "quiet", "q", false, fmt.Sprintf("use %s in scripted mode i.e. without an interactive prompt", nameCLI))
	// flag 4: debug, i.e. should debug messages be logged
	// terminal overrides config-file, default is false
	rootCmd.PersistentFlags().BoolVarP(&currentState.debug, "debug", "d", false, "enable logging of debug messages")
	zlog.Debug().Msg("End: init(): initialize flags")
	// viper bindings, one for each value in the struct called currentState
	zlog.Debug().Msg("Start: init(): bind flags")
	if err := conf.BindPFlag(selector_key, rootCmd.PersistentFlags().Lookup("selected-instance")); err != nil {
		zboth.Warn().Err(err).Msgf("Failed to bind flag: %s. Will ignore command line input.", "selected-instance")
	}
	if currentState.name != "" { // i.e. create these entries on "instance" only once an instance has been selected
		if err := conf.BindPFlag(joinKey("instances", currentState.name, "quiet"), rootCmd.PersistentFlags().Lookup("quiet")); err != nil {
			zboth.Warn().Err(err).Msgf("Failed to bind flag: %s. Will ignore command line input.", "quiet")
		}
		if err := conf.BindPFlag(joinKey("instances", currentState.name, "debug"), rootCmd.PersistentFlags().Lookup("debug")); err != nil {
			zboth.Warn().Err(err).Msgf("Failed to bind flag: %s. Will ignore command line input.", "debug")
		}
	}
	zlog.Debug().Msg("End: init(): bind flags")
	// initialize viper (runs last, i.e. when rootCmd.Execute runs)
	cobra.OnInitialize(initConf) // in configure.go
}
