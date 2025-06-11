package cli

import (
	"fmt"
	"math"
	"runtime"
	"time"

	"github.com/chigopher/pathlib"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/spf13/cobra"
)

func getDiskSpace() (msg string) {
	if diskStat, err := disk.Usage(pathlib.NewPath("/").String()); err == nil {
		total := float64(diskStat.Total) / math.Pow(2, 30) // Convert from bytes to GiB
		free := float64(diskStat.Free) / math.Pow(2, 30)   // Convert from bytes to GiB
		msg = fmt.Sprintf("\n- Disk space:\n  - %7.1fGi (total) %7.1fGi (free)", total, free)
	} else {
		zboth.Warn().Err(err).Msgf("Failed to retrieve information about disk space.")
	}
	return
}

func getMemory() (msg string) {
	if memStat, err := mem.VirtualMemory(); err == nil {
		total := float64(memStat.Total) / math.Pow(2, 30)    // Convert from bytes to GiB
		free := float64(memStat.Available) / math.Pow(2, 30) // Convert from bytes to GiB
		msg = fmt.Sprintf("\n- Memory:\n  - %7.1fGi (total) %7.1fGi (free)", total, free)
	} else {
		zboth.Warn().Err(err).Msgf("Failed to retrieve information about memory.")
	}
	return
}

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
			zboth.Info().Msg("Also writing all information in the log file.")
			zlog.Debug().Msg(info)
		}
		fmt.Println("This is what we know about the host machine:")
		fmt.Println(info)
	} else {
		sysFilename := toSprintf("%s.system.info", time.Now().Format("060102150405"))
		if err := workDir.Join(sysFilename).WriteFile([]byte(info + "\n")); err == nil {
			zboth.Info().Msgf("Written %s containing system information.", sysFilename)
		} else {
			zboth.Warn().Err(err).Msgf("Failed to write system.info. Writing all information in the log file instead.")
			zlog.Info().Msg(info)
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
