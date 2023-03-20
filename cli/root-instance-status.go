package cli

import (
	"strings"

	"github.com/spf13/cobra"
)

var _root_instance_status_all_ bool

func instanceStatus(givenName string) (status string) {
	out := getColumn(givenName, "Status")
	var statuses []string
	for _, line := range out { // determine what are the status messages for all associated containers
		l := strings.Split(line, " ") // use only the first word
		if len(l) > 0 {
			status := l[0] // use only the first word
			if l[len(l)-1] == "(Paused)" {
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

var statusInstanceRootCmd = &cobra.Command{
	Use:   "status",
	Args:  cobra.NoArgs,
	Short: "Get status of an instance of " + nameCLI,
	Run: func(cmd *cobra.Command, args []string) {
		logWhere()
		confirmInstalled()
		if !_root_instance_status_all_ {
			status := instanceStatus(currentState.name)
			zboth.Info().Msgf("The status of %s is: %s.", currentState.name, status)
		} else {
			for _, instance := range allInstances() {
				status := instanceStatus(instance)
				zboth.Info().Msgf("The status of %s is: %s.", instance, status)
			}
		}
	},
}

func init() {
	instanceRootCmd.AddCommand(statusInstanceRootCmd)
	statusInstanceRootCmd.Flags().BoolVar(&_root_instance_status_all_, "all", false, "show status of all instances")
}
