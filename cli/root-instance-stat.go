package cli

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var (
	_root_instance_stat_all_     bool
	_root_instance_stat_service_ string
)

func instanceStat(givenName, service string) {
	name, services, out, statOf := getInternalName(givenName), getServices(givenName), []string{""}, []string{}
	if _root_instance_stat_all_ {
		statOf = services
		zboth.Info().Msgf("Printing stats for the instance called %s.", givenName)
	} else {
		if elementInSlice(service, &services) > -1 {
			statOf = []string{service}
			zboth.Info().Msgf("Printing stats for the instance-service called %s-%s.", givenName, service)
		} else {
			zboth.Fatal().Err(fmt.Errorf("named service not found")).Msgf("No service called %s found associated with the instance called %s.", service, givenName)
		}
	}
	zboth.Info().Msgf("The status of %s is: %s.", givenName, instanceStatus(givenName))
	if res, err := execShell(fmt.Sprintf("%s stats --all --no-stream --no-trunc --format \"{{ .Name }} {{ .MemUsage }} {{ .MemPerc }} {{ .CPUPerc }}\"", toLower(virtualizer))); err == nil {
		out[0] = string(res)
		out = strings.Split(out[0], "\n")
		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 12, 8, 0, '\t', 0)
		fmt.Fprintf(w, "\n %s\t%s\t%s\t%s", "Name", " Memory ", "Mem %", "CPU %")
		fmt.Fprintf(w, "\n %s\t%s\t%s\t%s", "----", "--------", "-----", "-----")
		for _, service := range statOf {
			found := false
			for _, line := range out {
				zlog.Log().Msgf("stats for %s-%s: %s", name, service, line)
				l := strings.Split(line, " ")
				if l[0] == fmt.Sprintf("%s-%s-%d", name, service, rollNum) {
					fmt.Fprintf(w, "\n %s\t%s\t%s\t%s", service, l[1], l[4], l[5])
					found = true
					break
				}
			}
			if !found {
				zboth.Warn().Err(fmt.Errorf("stats not found")).Msgf("Error while parsing stats for the instance-container called %s-%s-%d.", name, service, rollNum)
			}
		}
		fmt.Fprintf(w, "\n")
		w.Flush()
	} else {
		zboth.Fatal().Err(err).Msgf("Failed to get stats from %s.", virtualizer)
	}
}

var statInstanceRootCmd = &cobra.Command{
	Use:     "stat",
	Aliases: []string{"stats"},
	Args:    cobra.NoArgs,
	Short:   "Get stats of an instance of " + nameCLI,
	Run: func(cmd *cobra.Command, args []string) {
		logWhere()
		confirmInstalled()
		if currentState.quiet {
			zboth.Warn().Err(fmt.Errorf("illegal operation")).Msgf("Stats can't be printed in quiet mode.")
		} else {
			if _root_instance_stat_service_ == "" {
				zboth.Debug().Msgf("No service specified, printing stats for all services.")
				_root_instance_stat_all_ = true
			}
			_root_instance_stat_service_ = toLower(_root_instance_stat_service_)
			instanceStat(currentState.name, _root_instance_stat_service_)
		}
	},
}

func init() {
	instanceRootCmd.AddCommand(statInstanceRootCmd)
	statInstanceRootCmd.Flags().StringVar(&_root_instance_stat_service_, "service", "", "show the stats of a given service")
	statInstanceRootCmd.Flags().BoolVar(&_root_instance_stat_all_, "all", false, "show the stats for all services of an instance")
}
