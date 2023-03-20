package cli

import (
	"strings"

	"github.com/spf13/cobra"
)

func instanceStatus(givenName string) (status string) {
	out := getColumn(givenName, "Status", "")
	var statuses []string
	for _, line := range out { // determine what are the status messages for all associated containers
		l := strings.Split(line, " ") // use only the first word
		if len(l) > 0 {
			status := l[0]                 // use only the first word
			if l[len(l)-1] == "(Paused)" { // use this if the last word is Paused
				status = "Paused"
			}
			if elementInSlice(status, &statuses) == -1 && len(status) != 0 {
				statuses = append(statuses, status)
			}
		}
	}
	if len(statuses) == 0 {
		status = "Instance not found"
	} else if len(statuses) == 1 {
		status = statuses[0]
	} else if len(statuses) > 1 {
		status = strings.Join(statuses, " and ")
	}
	return
}

func instanceStat(givenName string) {
	name, services, out := getInternalName(givenName), getServices(givenName), []string{""}
	zboth.Info().Msgf("The status of %s is: %s.\n\nIts stats are:", givenName, instanceStatus(givenName))
	if res, err := execShell(toSprintf("%s stats --all --no-stream --no-trunc --format \"{{ .Name }} {{ .ID }} {{ .MemUsage }} {{ .MemPerc }} {{ .CPUPerc }}\"", toLower(virtualizer))); err == nil {
		out[0] = string(res)
		out = strings.Split(out[0], "\n")
		zboth.Info().Msgf("%10s %10s %10s %10s %10s", "Name", "ID", " Memory", "Mem %", "CPU %")
		zboth.Info().Msgf("---------- ---------- ---------- ---------- ----------")
		for _, service := range services {
			found := false
			for _, line := range out {
				l := strings.Split(line, " ")
				if l[0] == toSprintf("%s-%s-%d", name, service, rollNum) {
					zboth.Info().Msgf("%10s %10s %10s %10s %10s", service, l[1][:10], l[2], l[5], l[6])
					found = true
					break
				}
			}
			if !found {
				zboth.Warn().Err(toError("stats not found")).Msgf("Error while parsing stats for the instance-container called %s-%s-%d.", name, service, rollNum)
			}
		}
	} else {
		zboth.Fatal().Err(err).Msgf("Failed to get stats from %s.", virtualizer)
	}
}

var statInstanceRootCmd = &cobra.Command{
	Use:     "stat",
	Aliases: []string{"stats", "status"},
	Args:    cobra.NoArgs,
	Short:   "Get status and status of an instance of " + nameCLI,
	Run: func(cmd *cobra.Command, _ []string) {
		if ownCall(cmd) && toBool(cmd.Flag("all").Value.String()) {
			existingInstances := allInstances()
			if len(existingInstances) == 1 {
				zboth.Info().Msgf("You have only one instance of %s. Ignoring the `--all` flag.", nameCLI)
				instanceStat(currentInstance)
			} else {
				for _, instance := range existingInstances {
					status := instanceStatus(instance)
					zboth.Info().Msgf("The status of %s is: %s.", instance, status)
				}
			}
		} else {
			instanceStat(currentInstance)
		}
	},
}

func init() {
	instanceRootCmd.AddCommand(statInstanceRootCmd)
	statInstanceRootCmd.Flags().Bool("all", false, "show the status of all instances")
}
