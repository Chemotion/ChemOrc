package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func instanceLog(givenName, args string, logOf *[]string, follow bool) {
	name := getInternalName(givenName)
	for _, service := range *logOf {
		zboth.Info().Msgf("Printing logs for the instance-service called %s-%s.", givenName, service)
		if follow {
			args += " --follow"
			callVirtualizer(toSprintf("logs %s %s-%s-%d", args, name, service, rollNum))
		} else {
			if res, err := execShell(toSprintf("%s logs %s %s-%s-%d", toLower(virtualizer), args, name, service, rollNum)); err == nil {
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
	Run: func(cmd *cobra.Command, _ []string) {
		if isInteractive(false) {
			var services, logOf []string = getServices(currentInstance), []string{}
			var (
				args   string
				follow bool
			)
			if ownCall(cmd) {
				logOf = []string{toLower(cmd.Flag("service").Value.String())}
				if cmd.Flag("all").Changed && toBool(cmd.Flag("all").Value.String()) {
					logOf = services
				} else {
					if cmd.Flag("service").Changed {
						if cmd.Flag("all").Changed {
							zboth.Warn().Msgf("You used the `--service` and `--all` flags. Ignoring the `--service` flag.")
						}
						if elementInSlice(logOf[0], &services) == -1 {
							zboth.Fatal().Err(toError("service not found")).Msgf("No service called %s found associated with the instance called %s.", logOf[0], currentInstance)
						}
					} else {
						logOf = []string{selectOpt(services, "Please select the service whose logs you want")}
					}
				}
				args := toSprintf("--tail %s", cmd.Flag("tail").Value.String())
				if toBool(cmd.Flag("details").Value.String()) {
					args += " --details"
				}
				if toBool(cmd.Flag("timestamps").Value.String()) {
					args += " --timestamps"
				}
				if cmd.Flag("since").Changed {
					args += toSprintf(" --since %s", cmd.Flag("since").Value.String())
				}
				if cmd.Flag("until").Changed {
					args += toSprintf(" --until %s", cmd.Flag("until").Value.String())
				}
				follow = toBool(cmd.Flag("follow").Value.String())
				if !cmd.Flag("all").Changed && args == "--tail all" && !cmd.Flag("service").Changed {
					zboth.Info().Msgf("This command can be used with a number of flags. Check them with the `--help` option.")
				}
			} else {
				zboth.Info().Msgf("This command can be used as with a number of flags. Check `%s %s --help` for more.", commandForCLI, "instance logs")
				logOf = []string{selectOpt(services, "Please select the service whose logs you want")}
			}
			instanceLog(currentInstance, args, &logOf, follow)
		} else {
			zboth.Fatal().Err(toError("illegal operation")).Msgf("Logs can't be printed in quiet mode.")
		}
	},
}

func init() {
	instanceRootCmd.AddCommand(logInstanceRootCmd)
	logInstanceRootCmd.Flags().String("service", primaryService, "show the log of a given service")
	logInstanceRootCmd.Flags().Bool("all", false, "show the logs of all services")
	logInstanceRootCmd.Flags().Bool("details", false, toSprintf("--details flag as received by %s logs command", virtualizer))
	logInstanceRootCmd.Flags().Bool("follow", false, toSprintf("--follow flag as received by %s logs command", virtualizer))
	logInstanceRootCmd.Flags().BoolP("timestamps", "t", false, toSprintf("--timestamps flag as received by %s logs command", virtualizer))
	logInstanceRootCmd.Flags().String("since", "", toSprintf("--since flag as received by %s logs command", virtualizer))
	logInstanceRootCmd.Flags().String("until", "", toSprintf("--until flag as received by %s logs command", virtualizer))
	logInstanceRootCmd.Flags().StringP("tail", "n", "all", toSprintf("--tail flag as received by %s logs command", virtualizer))
}
