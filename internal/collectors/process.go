package collectors

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"github.com/tiwariParth/whosay/internal/models"
	"github.com/tiwariParth/whosay/internal/ui"
)

// GetProcessInfo displays information about running processes
func GetProcessInfo(opts models.Options) {
	// Create default display options
	display := models.ProcessDisplay{
		SortBy:    "cpu",
		Ascending: false,
		Filter:    "",
		Limit:     10,
	}

	// Get processes
	processes, err := GetTopProcesses(display.Limit)
	if err != nil {
		fmt.Printf("Error getting process information: %v\n", err)
		return
	}

	// Apply sorting
	sortProcesses(processes, display.SortBy, display.Ascending)

	if opts.JSONOutput {
		jsonData, err := json.MarshalIndent(processes, "", "  ")
		if err != nil {
			fmt.Printf("Error serializing process data: %v\n", err)
			return
		}
		fmt.Println(string(jsonData))
		return
	}

	// Format and display the process info in compact view
	sections := GetProcessInfoSections(processes, opts)
	ui.CompactDisplay(sections)
}

// GetProcessInfoSections formats process information for the compact display
func GetProcessInfoSections(processes []models.ProcessInfo, opts models.Options) map[string][][]string {
    // Create the process section
    processData := [][]string{
        {"Count", fmt.Sprintf("%d", len(processes))},
    }

    // Format top processes for display with a simpler table layout
    topProcSection := [][]string{
        {"PID", "Name", "CPU%", "Memory%"},
    }

    // Add process data with proper column separation
    for i, proc := range processes {
        if i >= 10 { // Only show top 10 in compact view
            break
        }

        procName := proc.Name
        if len(procName) > 15 {
            procName = procName[:12] + "..." // Truncate long names
        }

        topProcSection = append(topProcSection, []string{
            fmt.Sprintf("%d", proc.PID),
            procName,
            fmt.Sprintf("%.1f", proc.CPU),
            fmt.Sprintf("%.1f", proc.Memory),
        })
    }

    // Build the final sections map
    result := map[string][][]string{
        "Processes": processData,
    }

    // Add top processes section
    if len(topProcSection) > 1 { // Only if we have actual processes
        result["Top Processes"] = topProcSection
    }

    return result
}

// GetTopProcesses returns the top processes by CPU or memory usage
func GetTopProcesses(limit int) ([]models.ProcessInfo, error) {
	switch runtime.GOOS {
	case "linux":
		return getLinuxProcesses(limit)
	case "darwin":
		return getDarwinProcesses(limit)
	case "windows":
		return getWindowsProcesses(limit)
	default:
		return []models.ProcessInfo{}, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// FindProcess returns processes matching the given name pattern
func FindProcess(name string) ([]models.ProcessInfo, error) {
	processes, err := GetTopProcesses(100) // Get a larger set to search in
	if err != nil {
		return nil, err
	}

	// Filter processes by name using case-insensitive matching
	nameRegex, err := regexp.Compile(fmt.Sprintf("(?i)%s", regexp.QuoteMeta(name)))
	if err != nil {
		return nil, fmt.Errorf("invalid search pattern: %v", err)
	}

	filtered := []models.ProcessInfo{}
	for _, proc := range processes {
		if nameRegex.MatchString(proc.Name) || nameRegex.MatchString(proc.CommandLine) {
			filtered = append(filtered, proc)
		}
	}

	return filtered, nil
}

// getLinuxProcesses gets process information on Linux
func getLinuxProcesses(limit int) ([]models.ProcessInfo, error) {
	result := []models.ProcessInfo{}

	// Run ps command to get process info
	cmd := exec.Command("ps", "-eo", "pid,ppid,user,%cpu,%mem,rss,stat,start,comm,cmd", "--sort=-%cpu")
	output, err := cmd.Output()
	if err != nil {
		return result, fmt.Errorf("failed to run ps command: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) <= 1 {
		return result, nil
	}

	// Skip the header line
	lines = lines[1:]

	// Parse each line
	for _, line := range lines {
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 9 {
			continue // Skip invalid lines
		}

		// Extract process info
		pid, err := strconv.Atoi(fields[0])
		if err != nil {
			continue
		}

		ppid, _ := strconv.Atoi(fields[1])
		user := fields[2]
		
		cpu, err := strconv.ParseFloat(fields[3], 64)
		if err != nil {
			cpu = 0.0
		}
		
		mem, err := strconv.ParseFloat(fields[4], 64)
		if err != nil {
			mem = 0.0
		}
		
		rss, err := strconv.ParseUint(fields[5], 10, 64)
		if err != nil {
			rss = 0
		}
		
		status := fields[6]
		name := fields[8]
		
		// Join remaining fields for command line
		cmdLine := ""
		if len(fields) >= 10 {
			cmdLine = strings.Join(fields[9:], " ")
		}

		proc := models.ProcessInfo{
			PID:         pid,
			PPID:        ppid,
			Name:        name,
			User:        user,
			CPU:         cpu,
			Memory:      mem,
			MemoryRSS:   rss,
			Status:      status,
			CommandLine: cmdLine,
		}

		result = append(result, proc)

		// Stop if we have enough processes
		if len(result) >= limit {
			break
		}
	}

	return result, nil
}

// getDarwinProcesses gets process information on macOS
func getDarwinProcesses(limit int) ([]models.ProcessInfo, error) {
	result := []models.ProcessInfo{}

	// Run ps command to get process info
	cmd := exec.Command("ps", "-eo", "pid,ppid,user,%cpu,%mem,rss,state,start,comm", "-r")
	output, err := cmd.Output()
	if err != nil {
		return result, fmt.Errorf("failed to run ps command: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) <= 1 {
		return result, nil
	}

	// Skip the header line
	lines = lines[1:]

	// Parse each line
	for _, line := range lines {
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 8 {
			continue // Skip invalid lines
		}

		// Extract process info
		pid, err := strconv.Atoi(fields[0])
		if err != nil {
			continue
		}

		ppid, _ := strconv.Atoi(fields[1])
		user := fields[2]
		
		cpu, err := strconv.ParseFloat(fields[3], 64)
		if err != nil {
			cpu = 0.0
		}
		
		mem, err := strconv.ParseFloat(fields[4], 64)
		if err != nil {
			mem = 0.0
		}
		
		rss, err := strconv.ParseUint(fields[5], 10, 64)
		if err != nil {
			rss = 0
		}
		
		status := fields[6]
		name := fields[8]

		proc := models.ProcessInfo{
			PID:        pid,
			PPID:       ppid,
			Name:       name,
			User:       user,
			CPU:        cpu,
			Memory:     mem,
			MemoryRSS:  rss,
			Status:     status,
		}

		result = append(result, proc)

		// Stop if we have enough processes
		if len(result) >= limit {
			break
		}
	}

	return result, nil
}

// getWindowsProcesses gets process information on Windows
func getWindowsProcesses(limit int) ([]models.ProcessInfo, error) {
	result := []models.ProcessInfo{}

	// On Windows we need to use the tasklist command
	cmd := exec.Command("tasklist", "/NH", "/FO", "CSV")
	output, err := cmd.Output()
	if err != nil {
		return result, fmt.Errorf("failed to run tasklist command: %w", err)
	}

	// Also get process details with wmic for more information
	cpuCmd := exec.Command("wmic", "path", "Win32_PerfFormattedData_PerfProc_Process", "get", "Name,PercentProcessorTime", "/format:csv")
	cpuOutput, _ := cpuCmd.Output()
	
	// Parse CPU usage
	cpuMap := make(map[string]float64)
	cpuLines := strings.Split(string(cpuOutput), "\n")
	for _, line := range cpuLines {
		if !strings.Contains(line, ",") {
			continue
		}
		
		parts := strings.Split(line, ",")
		if len(parts) < 3 {
			continue
		}
		
		name := parts[1]
		cpu, err := strconv.ParseFloat(parts[2], 64)
		if err == nil && name != "" {
			cpuMap[strings.ToLower(name)] = cpu
		}
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		// Parse CSV format - remove quotes
		line = strings.Trim(line, "\r\n")
		line = strings.ReplaceAll(line, "\"", "")
		fields := strings.Split(line, ",")
		
		if len(fields) < 5 {
			continue
		}

		name := fields[0]
		pid, err := strconv.Atoi(fields[1])
		if err != nil {
			continue
		}
		
		// Extract memory (convert from K to bytes)
		memStr := strings.ReplaceAll(fields[4], " K", "")
		memKB, err := strconv.ParseUint(memStr, 10, 64)
		if err != nil {
			memKB = 0
		}
		
		// Calculate memory percentage (rough estimate)
		memTotal := 8 * 1024 * 1024 // Assume 8GB if we can't detect
		memPerc := float64(memKB) / float64(memTotal) * 100
		
		// Look up CPU usage
		cpuUsage := cpuMap[strings.ToLower(name)]

		proc := models.ProcessInfo{
			PID:       pid,
			Name:      name,
			User:      "N/A", // Not easily available on Windows
			CPU:       cpuUsage,
			Memory:    memPerc,
			MemoryRSS: memKB,
			Status:    fields[2], // Session name as status
		}

		result = append(result, proc)

		// Stop if we have enough processes
		if len(result) >= limit {
			break
		}
	}

	return result, nil
}

// sortProcesses sorts the process list by the given field
func sortProcesses(processes []models.ProcessInfo, sortBy string, ascending bool) {
	switch strings.ToLower(sortBy) {
	case "cpu":
		sort.Slice(processes, func(i, j int) bool {
			if ascending {
				return processes[i].CPU < processes[j].CPU
			}
			return processes[i].CPU > processes[j].CPU
		})
	case "memory", "mem":
		sort.Slice(processes, func(i, j int) bool {
			if ascending {
				return processes[i].Memory < processes[j].Memory
			}
			return processes[i].Memory > processes[j].Memory
		})
	case "pid":
		sort.Slice(processes, func(i, j int) bool {
			if ascending {
				return processes[i].PID < processes[j].PID
			}
			return processes[i].PID > processes[j].PID
		})
	case "name":
		sort.Slice(processes, func(i, j int) bool {
			if ascending {
				return processes[i].Name < processes[j].Name
			}
			return processes[i].Name > processes[j].Name
		})
	}
}
