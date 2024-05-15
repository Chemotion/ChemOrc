package cli

import (
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

func instancePing(givenName string) (response string) {
	url := conf.GetString(joinKey(instancesWord, givenName, "accessAddress"))
	client := http.Client{Timeout: 2 * time.Second}
	if req, err := http.NewRequest("HEAD", url, nil); err == nil {
		if resp, err := client.Do(req); err == nil {
			resp.Body.Close()
			response = resp.Status
		} else {
			response = err.Error()
		}
	}
	return
}

// PingCmd represents the ping command
var pingInstanceRootCmd = &cobra.Command{
	Use:   "ping",
	Args:  cobra.NoArgs,
	Short: "Ping an instance of " + nameProject,
	Run: func(cmd *cobra.Command, _ []string) {
		if response := instancePing(currentInstance); response == "200 OK" {
			zboth.Info().Msgf("Success, received: %s.", response)
		} else {
			zboth.Warn().Msgf("Failed with response: %s.", response)
		}
	},
}

func init() {
	instanceRootCmd.AddCommand(pingInstanceRootCmd)
}
