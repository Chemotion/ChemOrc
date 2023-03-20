package cli

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

// helper function that is used by infoAdvancedRootCmd
func systemInfo() (info string) {
	// CPU
	info += fmt.Sprintln("- CPU Cores:", runtime.NumCPU())
	if runtime.GOOS == "linux" {
		info += getDiskSpace() // Disk Space
		info += getMemory()    // Memory
	}
	info += fmt.Sprintln("Used software versions:")
	printVersionOf := []string{"docker", "ruby", "passenger", "node", "npm"}
	for _, software := range printVersionOf {
		info += fmt.Sprintf("- %s: %s\n", strings.ToTitle(software), findVersion(software))
	}
	return
}

// Show host machine information to the user
// See also, chemotion instance info
var infoAdvancedRootCmd = &cobra.Command{
	Use:   "info",
	Args:  cobra.NoArgs,
	Short: "get information about the system",
	Run: func(cmd *cobra.Command, args []string) {
		logWhere()
		confirmInstalled()
		info := systemInfo()
		if currentState.quiet {
			if err := workDir.Join("system.info").WriteFile([]byte(info)); err != nil {
				zboth.Debug().Msgf(info)
			}
		} else {
			if currentState.debug {
				zboth.Debug().Msgf(info)
			} else {
				fmt.Println("This is what we know about the host machine:")
				fmt.Println(info)
			}
		}
	},
}

func init() {
	advancedRootCmd.AddCommand(infoAdvancedRootCmd)
}
