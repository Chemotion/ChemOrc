package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	_root_instance_log_all_            bool
	_root_instance_log_service_        string
	_root_instance_log_defaultService_ = "eln" // the service that is used as the default for printing log"
	_root_instance_log_v_details_      bool
	_root_instance_log_v_follow_       bool
	_root_instance_log_v_timestamps_   bool
	_root_instance_log_v_since_        string
	_root_instance_log_v_until_        string
	_root_instance_log_v_tail_         string
)

func instanceLog(givenName, service string) {
	name, services := getInternalName(givenName), getServices(givenName)
	var logOf []string
	if _root_instance_log_all_ {
		logOf = services
	} else {
		if elementInSlice(service, &services) > -1 {
			logOf = []string{service}
		} else {
			zboth.Fatal().Err(fmt.Errorf("named service not found")).Msgf("No service called %s found associated with the instance called %s.", service, givenName)
		}
	}
	for _, service := range logOf {
		zboth.Info().Msgf("Printing logs for the instance-service called %s-%s.", givenName, service)
		args := fmt.Sprintf("--tail %s", _root_instance_log_v_tail_)
		if _root_instance_log_v_details_ {
			args += " --details"
		}
		if _root_instance_log_v_timestamps_ {
			args += " --timestamps"
		}
		if _root_instance_log_v_since_ != "" {
			args += fmt.Sprintf(" --since %s", _root_instance_log_v_until_)
		}
		if _root_instance_log_v_until_ != "" {
			args += fmt.Sprintf(" --until %s", _root_instance_log_v_until_)
		}
		if _root_instance_log_v_follow_ {
			if _root_instance_log_all_ {
				zboth.Fatal().Err(fmt.Errorf("illegal operation")).Msgf("Cannot `follow` all the services. Use only one of the `--all` and `--follow` flags.")
			}
			args += " --follow"
			callVirtualizer(fmt.Sprintf("logs %s %s-%s-%d", args, name, service, rollNum))
		} else {
			if res, err := execShell(fmt.Sprintf("%s logs %s %s-%s-%d", toLower(virtualizer), args, name, service, rollNum)); err == nil {
				if n, errPrint := fmt.Println(string(res)); errPrint == nil {
					zboth.Debug().Msgf("Printed logs to screen that were %d lines long", n)
				} else {
					zboth.Warn().Err(errPrint).Msgf("Error while printing logs for the instance-container called %s-%s-%d.", name, service, rollNum)
				}
			} else {
				zboth.Fatal().Err(err).Msgf("Failed to get logs for the instance-container called %s-%s-%d.", name, service, rollNum)
			}
		}
	}
}

var logInstanceRootCmd = &cobra.Command{
	Use:     "log",
	Aliases: []string{"logs"},
	Args:    cobra.NoArgs,
	Short:   "Get logs of an instance of " + nameCLI,
	Run: func(cmd *cobra.Command, args []string) {
		logWhere()
		confirmInstalled()
		if currentState.quiet {
			zboth.Warn().Err(fmt.Errorf("illegal operation")).Msgf("Logs can't be printed in quiet mode.")
		} else {
			if _root_instance_log_service_ == "" {
				zboth.Info().Msgf("No service specified, printing logs for all services.")
				_root_instance_log_all_ = true
			}
			_root_instance_log_service_ = toLower(_root_instance_log_service_)
			instanceLog(currentState.name, _root_instance_log_service_)
		}
	},
}

func init() {
	instanceRootCmd.AddCommand(logInstanceRootCmd)
	logInstanceRootCmd.Flags().StringVar(&_root_instance_log_service_, "service", _root_instance_log_defaultService_, "show the log of a given service")
	logInstanceRootCmd.Flags().BoolVar(&_root_instance_log_all_, "all", false, "show the logs of all services")
	logInstanceRootCmd.Flags().BoolVar(&_root_instance_log_v_details_, "details", false, fmt.Sprintf("--details flag as received by %s logs command", virtualizer))
	logInstanceRootCmd.Flags().BoolVar(&_root_instance_log_v_follow_, "follow", false, fmt.Sprintf("--follow flag as received by %s logs command", virtualizer))
	logInstanceRootCmd.Flags().BoolVarP(&_root_instance_log_v_timestamps_, "timestamps", "t", false, fmt.Sprintf("--timestamps flag as received by %s logs command", virtualizer))
	logInstanceRootCmd.Flags().StringVar(&_root_instance_log_v_since_, "since", "", fmt.Sprintf("--since flag as received by %s logs command", virtualizer))
	logInstanceRootCmd.Flags().StringVar(&_root_instance_log_v_until_, "until", "", fmt.Sprintf("--until flag as received by %s logs command", virtualizer))
	logInstanceRootCmd.Flags().StringVarP(&_root_instance_log_v_tail_, "tail", "n", "all", fmt.Sprintf("--tail flag as received by %s logs command", virtualizer))
}
