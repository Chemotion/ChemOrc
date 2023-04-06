package cli

import (
	"fmt"
	"runtime"
	"time"

	"github.com/spf13/cobra"
)

// get system information
func getSystemInfo() (info string) {
	// CPU
	info += toSprintf("\n- CPU Cores: %d", runtime.NumCPU())
	info += getDiskSpace() // Disk Space
	info += getMemory()    // Memory
	return
}

// print system info depending on the debug tag
func systemInfo() {
	info := getSystemInfo()
	if isInteractive(false) {
		if conf.GetBool(joinKey(stateWord, "debug")) {
			zboth.Info().Msgf("Also writing all information in the log file.")
			zlog.Debug().Msgf(info)
		}
		fmt.Println("This is what we know about the host machine:")
		fmt.Println(info)
	} else {
		sysFilename := toSprintf("%s.system.info", time.Now().Format("060102150405"))
		if err := workDir.Join(sysFilename).WriteFile([]byte(info + "\n")); err == nil {
			zboth.Info().Msgf("Written %s containing system information.", sysFilename)
		} else {
			zboth.Warn().Err(err).Msgf("Failed to write system.info. Writing all information in the log file instead.")
			zlog.Info().Msgf(info)
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
