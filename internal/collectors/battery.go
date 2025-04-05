package collectors

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/tiwariParth/whosay/internal/models"
	"github.com/tiwariParth/whosay/internal/ui"
)

// GetBatteryInfo displays battery information
func GetBatteryInfo(opts models.Options) {
	info := collectBatteryInfo()

	if opts.JSONOutput {
		jsonData, _ := json.MarshalIndent(info, "", "  ")
		fmt.Println(string(jsonData))
		return
	}

	// Display battery information using the compact layout
	sections := GetBatteryInfoSections(opts)
	ui.CompactDisplay(sections)
}

// GetBatteryInfoSections returns formatted battery information sections
func GetBatteryInfoSections(opts models.Options) map[string][][]string {
	info := collectBatteryInfo()
	
	// If no battery is present, return a simple message
	if !info.IsPresent {
		return map[string][][]string{
			"Battery": {
				{"Status", "No battery detected"},
			},
		}
	}
	
	// Create compact data structure with consistent formatting
	batteryData := [][]string{
		{"Status", info.Status},
		{"Charge", fmt.Sprintf("%.1f%%", info.Percentage)},
	}
	
	// Add time remaining if available
	if info.TimeRemaining != "" {
		batteryData = append(batteryData, []string{"Time Remaining", info.TimeRemaining})
	}
	
	// Add health info if available
	if info.Health != "" {
		batteryData = append(batteryData, []string{"Health", info.Health})
	}
	
	// Add power draw if available
	if info.PowerDraw > 0 {
		batteryData = append(batteryData, []string{"Power Draw", fmt.Sprintf("%.1f W", info.PowerDraw)})
	}
	
	// Add cycle count if available
	if info.CycleCount > 0 {
		batteryData = append(batteryData, []string{"Cycle Count", fmt.Sprintf("%d", info.CycleCount)})
	}
	
	// Add technology if available
	if info.Technology != "" {
		batteryData = append(batteryData, []string{"Technology", info.Technology})
	}
	
	// Add capacity info if available
	if info.DesignCapacity > 0 && info.FullCapacity > 0 {
		healthPercent := float64(info.FullCapacity) / float64(info.DesignCapacity) * 100
		batteryData = append(batteryData, []string{
			"Capacity", fmt.Sprintf("%.1f Wh / %.1f Wh (%.1f%%)", 
				float64(info.FullCapacity)/1000.0, 
				float64(info.DesignCapacity)/1000.0,
				healthPercent),
		})
	}
	
	// Add usage bar with consistent width
	barWidth := 20 // Simplified, always use same width for consistency
	if opts.CompactMode {
		barWidth = 15 // Even smaller for compact mode
	}
	
	batteryData = append(batteryData, []string{
		"", ui.PrintCompactUsageBar("", info.Percentage, barWidth),
	})
	
	// Return data for unified display
	return map[string][][]string{
		"Battery": batteryData,
	}
}

// collectBatteryInfo gathers battery information
func collectBatteryInfo() models.BatteryInfo {
	info := models.BatteryInfo{
		IsPresent: false,
		Status:    "Unknown",
	}
	
	switch runtime.GOOS {
	case "linux":
		return getLinuxBatteryInfo()
	case "darwin":
		return getDarwinBatteryInfo()
	case "windows":
		return getWindowsBatteryInfo()
	}
	
	return info
}

// getLinuxBatteryInfo collects battery info on Linux
func getLinuxBatteryInfo() models.BatteryInfo {
	info := models.BatteryInfo{
		IsPresent: false,
		Status:    "Not Present",
	}
	
	// Check if battery exists in sysfs
	batteryPath := "/sys/class/power_supply/BAT0"
	if _, err := os.Stat(batteryPath); os.IsNotExist(err) {
		// Try BAT1 as fallback
		batteryPath = "/sys/class/power_supply/BAT1"
		if _, err := os.Stat(batteryPath); os.IsNotExist(err) {
			return info
		}
	}
	
	// Battery exists
	info.IsPresent = true
	
	// Read battery status
	statusFile, err := os.ReadFile(filepath.Join(batteryPath, "status"))
	if err == nil {
		info.Status = strings.TrimSpace(string(statusFile))
	}
	
	// Read battery capacity (percentage)
	capacityFile, err := os.ReadFile(filepath.Join(batteryPath, "capacity"))
	if err == nil {
		capacity, err := strconv.ParseFloat(strings.TrimSpace(string(capacityFile)), 64)
		if err == nil {
			info.Percentage = capacity
		}
	}
	
	// Try to get energy values
	energyFullFile, err := os.ReadFile(filepath.Join(batteryPath, "energy_full"))
	if err == nil {
		energyFull, err := strconv.ParseUint(strings.TrimSpace(string(energyFullFile)), 10, 64)
		if err == nil {
			info.FullCapacity = energyFull * 1000 // Convert to mWh
		}
	}
	
	energyFullDesignFile, err := os.ReadFile(filepath.Join(batteryPath, "energy_full_design"))
	if err == nil {
		energyFullDesign, err := strconv.ParseUint(strings.TrimSpace(string(energyFullDesignFile)), 10, 64)
		if err == nil {
			info.DesignCapacity = energyFullDesign * 1000 // Convert to mWh
		}
	}
	
	// Try to get technology
	technologyFile, err := os.ReadFile(filepath.Join(batteryPath, "technology"))
	if err == nil {
		info.Technology = strings.TrimSpace(string(technologyFile))
	}
	
	// Try to get cycle count
	cycleCountFile, err := os.ReadFile(filepath.Join(batteryPath, "cycle_count"))
	if err == nil {
		cycleCount, err := strconv.Atoi(strings.TrimSpace(string(cycleCountFile)))
		if err == nil {
			info.CycleCount = cycleCount
		}
	}
	
	// Calculate time remaining if discharging
	if info.Status == "Discharging" {
		powerNowFile, err := os.ReadFile(filepath.Join(batteryPath, "power_now"))
		if err == nil {
			powerNow, err := strconv.ParseFloat(strings.TrimSpace(string(powerNowFile)), 64)
			if err == nil && powerNow > 0 {
				info.PowerDraw = powerNow / 1000000.0 // Convert to W
				
				// Try to get current energy level
				energyNowFile, err := os.ReadFile(filepath.Join(batteryPath, "energy_now"))
				if err == nil {
					energyNow, err := strconv.ParseFloat(strings.TrimSpace(string(energyNowFile)), 64)
					if err == nil && energyNow > 0 {
						// Calculate time remaining in hours
						hoursRemaining := energyNow / powerNow
						hours := int(hoursRemaining)
						minutes := int((hoursRemaining - float64(hours)) * 60)
						
						info.TimeRemaining = fmt.Sprintf("%dh %dm", hours, minutes)
					}
				}
			}
		}
	}
	
	// Set health based on capacity degradation
	if info.DesignCapacity > 0 && info.FullCapacity > 0 {
		healthPercent := float64(info.FullCapacity) / float64(info.DesignCapacity) * 100
		
		if healthPercent >= 80 {
			info.Health = "Good"
		} else if healthPercent >= 60 {
			info.Health = "Fair"
		} else if healthPercent >= 40 {
			info.Health = "Poor"
		} else {
			info.Health = "Bad"
		}
	}
	
	return info
}

// getDarwinBatteryInfo collects battery info on macOS
func getDarwinBatteryInfo() models.BatteryInfo {
	info := models.BatteryInfo{
		IsPresent: false,
		Status:    "Not Present",
	}
	
	// Use system_profiler to get battery info
	cmd := exec.Command("system_profiler", "SPPowerDataType")
	output, err := cmd.Output()
	if err != nil {
		return info
	}
	
	// Check if battery info is present
	batteryInfo := string(output)
	if !strings.Contains(batteryInfo, "Battery Information") {
		return info
	}
	
	// Battery exists
	info.IsPresent = true
	
	// Parse percentage
	percentRegex := regexp.MustCompile(`Charge: (\d+)%`)
	if match := percentRegex.FindStringSubmatch(batteryInfo); len(match) > 1 {
		if percent, err := strconv.ParseFloat(match[1], 64); err == nil {
			info.Percentage = percent
		}
	}
	
	// Parse status
	if strings.Contains(batteryInfo, "Charging") {
		info.Status = "Charging"
	} else if strings.Contains(batteryInfo, "Discharging") {
		info.Status = "Discharging"
	} else if strings.Contains(batteryInfo, "AC Power") {
		info.Status = "Full"
	}
	
	// Parse time remaining
	timeRegex := regexp.MustCompile(`Time Remaining: (\d+):(\d+)`)
	if match := timeRegex.FindStringSubmatch(batteryInfo); len(match) > 2 {
		hours, _ := strconv.Atoi(match[1])
		minutes, _ := strconv.Atoi(match[2])
		info.TimeRemaining = fmt.Sprintf("%dh %dm", hours, minutes)
	}
	
	// Parse cycle count
	cycleRegex := regexp.MustCompile(`Cycle Count: (\d+)`)
	if match := cycleRegex.FindStringSubmatch(batteryInfo); len(match) > 1 {
		if cycles, err := strconv.Atoi(match[1]); err == nil {
			info.CycleCount = cycles
		}
	}
	
	// Parse health
	if strings.Contains(batteryInfo, "Condition: Normal") {
		info.Health = "Good"
	} else if strings.Contains(batteryInfo, "Condition: Replace Soon") {
		info.Health = "Fair"
	} else if strings.Contains(batteryInfo, "Condition: Replace Now") {
		info.Health = "Poor"
	} else if strings.Contains(batteryInfo, "Condition: Service Battery") {
		info.Health = "Bad"
	}
	
	return info
}

// getWindowsBatteryInfo collects battery info on Windows
func getWindowsBatteryInfo() models.BatteryInfo {
	info := models.BatteryInfo{
		IsPresent: false,
		Status:    "Not Present",
	}
	
	// Use WMIC to get battery info
	cmd := exec.Command("wmic", "Path", "Win32_Battery", "Get", "EstimatedChargeRemaining,BatteryStatus,EstimatedRunTime,Status")
	output, err := cmd.Output()
	if err != nil {
		return info
	}
	
	// If the output is empty or just headers, no battery is present
	batteryInfo := string(output)
	if strings.Count(batteryInfo, "\n") <= 1 {
		return info
	}
	
	// Battery exists
	info.IsPresent = true
	
	// Parse the output - split into lines and extract the data line
	lines := strings.Split(batteryInfo, "\n")
	if len(lines) < 2 {
		return info
	}
	
	// Get the data line (typically the second line)
	dataLine := ""
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) != "" {
			dataLine = strings.TrimSpace(lines[i])
			break
		}
	}
	
	if dataLine == "" {
		return info
	}
	
	// Split the data into fields - this is tricky because WMIC output
	// can have inconsistent spacing
	fields := strings.Fields(dataLine)
	
	// If we have at least one field, assume it's the charge percentage
	if len(fields) >= 1 {
		if percent, err := strconv.ParseFloat(fields[0], 64); err == nil {
			info.Percentage = percent
		}
	}
	
	// If we have at least two fields, use the second for status
	if len(fields) >= 2 {
		switch fields[1] {
		case "1":
			info.Status = "Discharging"
		case "2":
			info.Status = "Charging"
		case "3", "4", "5":
			info.Status = "Full"
		default:
			info.Status = "Unknown"
		}
	}
	
	// If we have at least three fields, use the third for estimated runtime
	if len(fields) >= 3 {
		if runtime, err := strconv.Atoi(fields[2]); err == nil && runtime > 0 && runtime < 71582 {
			// WMIC returns minutes, convert to hours and minutes
			hours := runtime / 60
			minutes := runtime % 60
			info.TimeRemaining = fmt.Sprintf("%dh %dm", hours, minutes)
		}
	}
	
	return info
}
