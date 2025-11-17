package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"
)

// SystemStats holds all system statistics
type SystemStats struct {
	CPU       CPUStats       `json:"cpu"`
	Memory    MemoryStats    `json:"memory"`
	Processes []ProcessInfo  `json:"processes"`
	Uptime    float64        `json:"uptime"`
}

// CPUStats holds CPU information
type CPUStats struct {
	Usage      float64  `json:"usage"`
	Cores      int      `json:"cores"`
	ModelName  string   `json:"modelName"`
	User       uint64   `json:"user"`
	System     uint64   `json:"system"`
	Idle       uint64   `json:"idle"`
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

var lastCPUStats CPUStats
var lastCPUTime time.Time

// GetSystemStats returns current system statistics
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

// getCPUStats reads CPU statistics from /proc/stat and /proc/cpuinfo
func getCPUStats() (CPUStats, error) {
	stats := CPUStats{}

	// Read /proc/stat for CPU usage
	file, err := os.Open("/proc/stat")
	if err != nil {
		return stats, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) > 0 && fields[0] == "cpu" {
			if len(fields) >= 5 {
				stats.User, _ = strconv.ParseUint(fields[1], 10, 64)
				stats.System, _ = strconv.ParseUint(fields[3], 10, 64)
				stats.Idle, _ = strconv.ParseUint(fields[4], 10, 64)
			}
		}
	}

	// Calculate CPU usage percentage
	if !lastCPUTime.IsZero() {
		totalDelta := float64((stats.User + stats.System + stats.Idle) - (lastCPUStats.User + lastCPUStats.System + lastCPUStats.Idle))
		idleDelta := float64(stats.Idle - lastCPUStats.Idle)
		if totalDelta > 0 {
			stats.Usage = ((totalDelta - idleDelta) / totalDelta) * 100
		}
	}

	lastCPUStats = stats
	lastCPUTime = time.Now()

	// Read /proc/cpuinfo for CPU model and core count
	cpuInfo, err := ioutil.ReadFile("/proc/cpuinfo")
	if err == nil {
		lines := strings.Split(string(cpuInfo), "\n")
		coreCount := 0
		for _, line := range lines {
			if strings.HasPrefix(line, "model name") {
				parts := strings.Split(line, ":")
				if len(parts) > 1 && stats.ModelName == "" {
					stats.ModelName = strings.TrimSpace(parts[1])
				}
			}
			if strings.HasPrefix(line, "processor") {
				coreCount++
			}
		}
		stats.Cores = coreCount
	}

	return stats, nil
}

// getMemoryStats reads memory statistics from /proc/meminfo
func getMemoryStats() (MemoryStats, error) {
	stats := MemoryStats{}

	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return stats, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		value, _ := strconv.ParseUint(fields[1], 10, 64)
		// Convert from KB to bytes
		value *= 1024

		switch fields[0] {
		case "MemTotal:":
			stats.Total = value
		case "MemAvailable:":
			stats.Available = value
		case "MemFree:":
			stats.Free = value
		case "Buffers:":
			stats.Buffers = value
		case "Cached:":
			stats.Cached = value
		case "SwapTotal:":
			stats.SwapTotal = value
		case "SwapFree:":
			stats.SwapFree = value
		}
	}

	stats.Used = stats.Total - stats.Available
	if stats.Total > 0 {
		stats.UsedPercent = float64(stats.Used) / float64(stats.Total) * 100
	}
	stats.SwapUsed = stats.SwapTotal - stats.SwapFree

	return stats, nil
}

// getProcessList reads process information from /proc
func getProcessList() ([]ProcessInfo, error) {
	processes := []ProcessInfo{}

	files, err := ioutil.ReadDir("/proc")
	if err != nil {
		return processes, err
	}

	for _, f := range files {
		if !f.IsDir() {
			continue
		}

		// Check if directory name is a number (PID)
		pid, err := strconv.Atoi(f.Name())
		if err != nil {
			continue
		}

		process, err := getProcessInfo(pid)
		if err != nil {
			continue // Skip processes we can't read
		}

		processes = append(processes, process)

		// Limit to top 50 processes to avoid overwhelming the UI
		if len(processes) >= 50 {
			break
		}
	}

	return processes, nil
}

// getProcessInfo reads information for a specific process
func getProcessInfo(pid int) (ProcessInfo, error) {
	info := ProcessInfo{PID: pid}

	// Read process name from /proc/[pid]/comm
	commPath := fmt.Sprintf("/proc/%d/comm", pid)
	commData, err := ioutil.ReadFile(commPath)
	if err != nil {
		return info, err
	}
	info.Name = strings.TrimSpace(string(commData))

	// Read process status from /proc/[pid]/status
	statusPath := fmt.Sprintf("/proc/%d/status", pid)
	statusFile, err := os.Open(statusPath)
	if err != nil {
		return info, err
	}
	defer statusFile.Close()

	scanner := bufio.NewScanner(statusFile)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		switch fields[0] {
		case "State:":
			info.State = fields[1]
		case "Threads:":
			info.Threads, _ = strconv.Atoi(fields[1])
		case "VmRSS:":
			// Memory in KB, convert to bytes
			mem, _ := strconv.ParseUint(fields[1], 10, 64)
			info.Memory = mem * 1024
		}
	}

	return info, nil
}

// getUptime reads system uptime from /proc/uptime
func getUptime() (float64, error) {
	data, err := ioutil.ReadFile("/proc/uptime")
	if err != nil {
		return 0, err
	}

	fields := strings.Fields(string(data))
	if len(fields) > 0 {
		uptime, err := strconv.ParseFloat(fields[0], 64)
		if err != nil {
			return 0, err
		}
		return uptime, nil
	}

	return 0, fmt.Errorf("invalid uptime format")
}
