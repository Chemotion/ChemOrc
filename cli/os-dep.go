package cli

import (
	"fmt"
	"math"

	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
)

func getDiskSpace() string {
	diskStat, err := disk.Usage("/")
	if err != nil {
		zboth.Warn().Err(err).Msgf("Failed to retrieve information about disk space.")
		return ""
	}

	total := float64(diskStat.Total) / math.Pow(2, 30) // Convert from bytes to GiB
	free := float64(diskStat.Free) / math.Pow(2, 30)   // Convert from bytes to GiB

	return fmt.Sprintf("\n- Disk space:\n  - %7.1fGi (total) %7.1fGi (free)", total, free)
}

func getMemory() string {
	memStat, err := mem.VirtualMemory()
	if err != nil {
		zboth.Warn().Err(err).Msgf("Failed to retrieve information about memory.")
		return ""
	}

	total := float64(memStat.Total) / math.Pow(2, 30) // Convert from bytes to GiB
	free := float64(memStat.Available) / math.Pow(2, 30) // Convert from bytes to GiB

	return fmt.Sprintf("\n- Memory:\n  - %7.1fGi (total) %7.1fGi (free)", total, free)
}
