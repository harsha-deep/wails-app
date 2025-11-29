package main

import (
	"context"
	"fmt"
	"sort"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
)

// SystemStats holds all system statistics
type SystemStats struct {
	CPU       CPUStats      `json:"cpu"`
	Memory    MemoryStats   `json:"memory"`
	Processes []ProcessInfo `json:"processes"`
	Uptime    float64       `json:"uptime"`
}

// CPUStats holds CPU information
type CPUStats struct {
	Usage     float64 `json:"usage"`
	Cores     int     `json:"cores"`
	ModelName string  `json:"modelName"`
	User      uint64  `json:"user"`
	System    uint64  `json:"system"`
	Idle      uint64  `json:"idle"`
}

// MemoryStats holds memory information
type MemoryStats struct {
	Total       uint64  `json:"total"`
	Available   uint64  `json:"available"`
	Used        uint64  `json:"used"`
	UsedPercent float64 `json:"usedPercent"`
	Free        uint64  `json:"free"`
	Buffers     uint64  `json:"buffers"`
	Cached      uint64  `json:"cached"`
	SwapTotal   uint64  `json:"swapTotal"`
	SwapFree    uint64  `json:"swapFree"`
	SwapUsed    uint64  `json:"swapUsed"`
}

// ProcessInfo holds process information
type ProcessInfo struct {
	PID     int     `json:"pid"`
	Name    string  `json:"name"`
	State   string  `json:"state"`
	CPU     float64 `json:"cpu"`
	Memory  uint64  `json:"memory"`
	Threads int     `json:"threads"`
}

// GetSystemStats returns current system statistics (keeps same signature)
func (a *App) GetSystemStats() (*SystemStats, error) {
	stats := &SystemStats{}

	var err error
	stats.CPU, err = getCPUStats()
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU stats: %v", err)
	}

	stats.Memory, err = getMemoryStats()
	if err != nil {
		return nil, fmt.Errorf("failed to get memory stats: %v", err)
	}

	stats.Processes, err = getProcessList()
	if err != nil {
		return nil, fmt.Errorf("failed to get process list: %v", err)
	}

	stats.Uptime, err = getUptime()
	if err != nil {
		return nil, fmt.Errorf("failed to get uptime: %v", err)
	}

	return stats, nil
}

// getCPUStats uses gopsutil to populate CPUStats
func getCPUStats() (CPUStats, error) {
	var stats CPUStats

	// CPU times for aggregates (user, system, idle)
	times, err := cpu.Times(false)
	if err == nil && len(times) > 0 {
		stats.User = uint64(times[0].User)
		stats.System = uint64(times[0].System)
		stats.Idle = uint64(times[0].Idle)
	}

	// CPU percentage (overall)
	// cpu.Percent with interval 0 returns instantaneous since last call on some systems
	percentages, err := cpu.Percent(0, false)
	if err == nil && len(percentages) > 0 {
		stats.Usage = percentages[0]
	}

	// CPU core count (logical)
	cores, err := cpu.Counts(true)
	if err == nil {
		stats.Cores = cores
	}

	// CPU model name from cpu.Info()
	info, err := cpu.Info()
	if err == nil && len(info) > 0 {
		stats.ModelName = info[0].ModelName
	}

	return stats, nil
}

// getMemoryStats uses gopsutil to populate MemoryStats
func getMemoryStats() (MemoryStats, error) {
	var stats MemoryStats

	vm, err := mem.VirtualMemory()
	if err != nil {
		return stats, err
	}

	stats.Total = vm.Total
	stats.Available = vm.Available
	stats.Free = vm.Free
	stats.Buffers = vm.Buffers
	stats.Cached = vm.Cached
	stats.Used = vm.Used
	if vm.Total > 0 {
		stats.UsedPercent = vm.UsedPercent
	}
	// gopsutil returns swap info separately
	s := vm.SwapTotal // vm doesn't include swap in all builds, use swap.VirtualMemory() below as fallback

	// Use swap.VirtualMemory for swap fields
	swap, serr := mem.SwapMemory()
	if serr == nil {
		stats.SwapTotal = swap.Total
		stats.SwapFree = swap.Free
		stats.SwapUsed = swap.Used
	} else {
		// fallback to zeroes if unavailable
		stats.SwapTotal = 0
		stats.SwapFree = 0
		stats.SwapUsed = 0
		_ = s
	}

	return stats, nil
}

// getProcessList enumerates processes, collects data, sorts by memory, returns top 50
func getProcessList() ([]ProcessInfo, error) {
	ctx := context.Background()

	procs, err := process.Processes()
	if err != nil {
		return nil, err
	}

	infos := make([]ProcessInfo, 0, len(procs))

	// Collect info (best-effort; skip processes that error)
	for _, p := range procs {
		pi, err := getProcessInfoWithContext(ctx, p)
		if err != nil {
			continue
		}
		infos = append(infos, pi)
	}

	// Sort by memory descending (largest first)
	sort.Slice(infos, func(i, j int) bool {
		return infos[i].Memory > infos[j].Memory
	})

	// Limit to top 50
	if len(infos) > 50 {
		infos = infos[:50]
	}

	return infos, nil
}

// getProcessInfoWithContext reads process details using gopsutil's Process object
// getProcessInfoWithContext reads process details using gopsutil's Process object
func getProcessInfoWithContext(ctx context.Context, p *process.Process) (ProcessInfo, error) {
	info := ProcessInfo{}
	pid := int(p.Pid)
	info.PID = pid

	// Name
	if name, err := p.NameWithContext(ctx); err == nil {
		info.Name = name
	} else {
		if exe, err2 := p.ExeWithContext(ctx); err2 == nil {
			info.Name = exe
		}
	}

	// Status / State: StatusWithContext can return []string
	if statusSlice, err := p.StatusWithContext(ctx); err == nil {
		if len(statusSlice) > 0 {
			// Use the first status token (e.g., "R", "S"). If you prefer a joined string:
			// info.State = strings.Join(statusSlice, ",")
			info.State = statusSlice[0]
		}
	}

	// CPU percent (best-effort)
	if cpuPct, err := p.CPUPercentWithContext(ctx); err == nil {
		info.CPU = cpuPct
	}

	// Memory RSS
	if mi, err := p.MemoryInfoWithContext(ctx); err == nil {
		info.Memory = mi.RSS
	}

	// Threads
	if th, err := p.NumThreadsWithContext(ctx); err == nil {
		info.Threads = int(th)
	}

	return info, nil
}

// getUptime returns system uptime in seconds
func getUptime() (float64, error) {
	u, err := host.Uptime()
	if err != nil {
		return 0, err
	}
	return float64(u), nil
}
