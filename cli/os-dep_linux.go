//go:build linux
// +build linux

package cli

import (
	"math"
	"runtime"
	"syscall"
)

func getDiskSpace() (line string) {
	if runtime.GOOS == "linux" {
		var disk syscall.Statfs_t
		if err := syscall.Statfs(workDir.String(), &disk); err == nil {
			line = toSprintf("\n- Disk space:\n  - %7.1fGi (total) %7.1fGi (free)", float64(disk.Blocks*uint64(disk.Bsize))/math.Pow(2, 30), float64(disk.Bavail*uint64(disk.Bsize))/math.Pow(2, 30))
		} else {
			zboth.Warn().Err(err).Msgf("Failed to retrieve information about disk space.")
		}
	} else {
		zboth.Warn().Err(toError("running on %s", runtime.GOOS)).Msgf("Cannot retrieve disk space information for this operating system.")
	}
	return
}

func getMemory() (line string) {
	if runtime.GOOS == "linux" {
		var mem syscall.Sysinfo_t
		if err := syscall.Sysinfo(&mem); err == nil {
			line = toSprintf("\n- Memory:\n  - %7.1fGi (total) %7.1fGi (free)", float64(mem.Totalram)/math.Pow(2, 30), float64(mem.Freeram)/math.Pow(2, 30))
		} else {
			zboth.Warn().Err(err).Msgf("Failed to retrieve information about memory.")
		}
	} else {
		zboth.Warn().Err(toError("running on %s", runtime.GOOS)).Msgf("Cannot retrieve memory information for this operating system.")
	}
	return
}
