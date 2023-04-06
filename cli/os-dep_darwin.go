//go:build darwin
// +build darwin

package cli

import (
	"math"
	"runtime"
	"syscall"
)

// stub function
func getDiskSpace() (line string) {
	if runtime.GOOS == "darwin" {
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

// stub function
func getMemory() (line string) {
	return
}
