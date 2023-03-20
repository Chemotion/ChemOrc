package cli

import (
	"fmt"
	"runtime"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

// get system information
func getSystemInfo() (info string) {
	// CPU
	info += toSprintf("\n- CPU Cores: %d", runtime.NumCPU())
	info += getDiskSpace() // Disk Space
	info += getMemory()    // Memory
	// info += fmt.Sprintln("Used software versions:") // TODO: fix this
	// printVersionOf := []string{"docker", "ruby", "passenger", "node", "npm"}
	// for _, software := range printVersionOf {
	// 	info += toSprintf("- %s: %s\n", strings.ToTitle(software), findVersion(software))
	// }
	return
}

// print system info depending on the debug tag
func systemInfo() {
	info := getSystemInfo()
	if isInteractive(false) {
		if conf.GetBool(joinKey(stateWord, "debug")) {
			zboth.Info().Msgf("Also writing all information in the log file.")
			zboth.Debug().Msgf(info)
		}
		fmt.Println("This is what we know about the host machine:")
		fmt.Println(info)
	} else {
		if err := workDir.Join("system.info").WriteFile([]byte(info + "\n")); err == nil {
			zboth.Info().Msgf("Written system.info containing system information.")
		} else {
			zboth.Warn().Err(err).Msgf("Failed to write system.info. Writing all information in the log file.")
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
			zboth.Debug().Msgf(info)
			if !conf.GetBool(joinKey(stateWord, "debug")) {
				zerolog.SetGlobalLevel(zerolog.InfoLevel)
			}
		}
	}
}

var infoAdvancedRootCmd = &cobra.Command{
	Use:   "info",
	Args:  cobra.NoArgs,
	Short: "get information about the system",
	Run: func(_ *cobra.Command, _ []string) {
		systemInfo()
	},
}

func init() {
	advancedRootCmd.AddCommand(infoAdvancedRootCmd)
}
