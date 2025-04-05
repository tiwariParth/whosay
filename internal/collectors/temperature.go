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
	"time"

	"github.com/tiwariParth/whosay/internal/models"
	"github.com/tiwariParth/whosay/internal/ui"
)

var temperatureHistory []models.TemperatureHistoryRecord

// GetTemperatureInfo displays system temperature information
func GetTemperatureInfo(opts models.Options) {
	info := collectTemperatureInfo()

	if opts.JSONOutput {
		jsonData, _ := json.MarshalIndent(info, "", "  ")
		fmt.Println(string(jsonData))
		return
	}

	sections := GetTemperatureInfoSections(opts)
	ui.CompactDisplay(sections)
}

// GetTemperatureInfoSections formats temperature information
func GetTemperatureInfoSections(opts models.Options) map[string][][]string {
	info := collectTemperatureInfo()
	
	tempData := [][]string{}
	
	if info.CPU > 0 {
		tempData = append(tempData, []string{
			"CPU", fmt.Sprintf("%.1f°%s", info.CPU, info.Units),
		})
	}
	
	if info.GPU > 0 {
		tempData = append(tempData, []string{
			"GPU", fmt.Sprintf("%.1f°%s", info.GPU, info.Units),
		})
	}
	
	for component, temp := range info.Components {
		if temp > 0 {
			tempData = append(tempData, []string{
				component, fmt.Sprintf("%.1f°%s", temp, info.Units),
			})
		}
	}
	
	if len(tempData) == 0 {
		tempData = append(tempData, []string{
			"Status", "No temperature sensors detected",
		})
	}
	
	return map[string][][]string{
		"Temperature": tempData,
	}
}

// collectTemperatureInfo gathers temperature information
func collectTemperatureInfo() models.TemperatureInfo {
	info := models.TemperatureInfo{
		Units:      "C",
		Components: make(map[string]float64),
	}
	
	switch runtime.GOOS {
	case "linux":
		return getLinuxTemperatures()
	case "darwin":
		return getDarwinTemperatures()
	case "windows":
		return getWindowsTemperatures()
	}
	
	return info
}

// Platform-specific temperature collection
func getLinuxTemperatures() models.TemperatureInfo {
	info := models.TemperatureInfo{
		Units:      "C",
		Components: make(map[string]float64),
	}
	
	// Method 1: Read from sysfs thermal zones
	thermalZonesPath := "/sys/class/thermal"
	if _, err := os.Stat(thermalZonesPath); err == nil {
		items, err := os.ReadDir(thermalZonesPath)
		if err == nil {
			for _, item := range items {
				if strings.HasPrefix(item.Name(), "thermal_zone") {
					zonePath := filepath.Join(thermalZonesPath, item.Name())
					
					typeBytes, err := os.ReadFile(filepath.Join(zonePath, "type"))
					if err != nil {
						continue
					}
					zoneType := strings.TrimSpace(string(typeBytes))
					
					tempBytes, err := os.ReadFile(filepath.Join(zonePath, "temp"))
					if err != nil {
						continue
					}
					tempValue, err := strconv.ParseFloat(strings.TrimSpace(string(tempBytes)), 64)
					if err != nil {
						continue
					}
					
					tempValue /= 1000.0
					
					switch {
					case strings.Contains(zoneType, "cpu"):
						info.CPU = tempValue
					case strings.Contains(zoneType, "gpu"):
						info.GPU = tempValue
					default:
						info.Components[zoneType] = tempValue
					}
				}
			}
		}
	}
	
	// Method 2: Use lm-sensors if available
	if info.CPU == 0 {
		cmd := exec.Command("sensors", "-j")
		output, err := cmd.Output()
		if err == nil {
			var sensorsData map[string]interface{}
			if err := json.Unmarshal(output, &sensorsData); err == nil {
				for _, data := range sensorsData {
					if adapterData, ok := data.(map[string]interface{}); ok {
						for chip, chipData := range adapterData {
							if chipMap, ok := chipData.(map[string]interface{}); ok {
								for key, value := range chipMap {
									if strings.Contains(strings.ToLower(key), "temp") || strings.Contains(strings.ToLower(key), "temperature") {
										if tempMap, ok := value.(map[string]interface{}); ok {
											if tempValue, ok := tempMap["temp1_input"].(float64); ok {
												chipLower := strings.ToLower(chip)
												if strings.Contains(chipLower, "cpu") || strings.Contains(chipLower, "core") {
													info.CPU = tempValue
												} else if strings.Contains(chipLower, "gpu") || strings.Contains(chipLower, "graphics") {
													info.GPU = tempValue
												} else {
													info.Components[chip] = tempValue
												}
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
	
	return info
}

func getDarwinTemperatures() models.TemperatureInfo {
	info := models.TemperatureInfo{
		Units:      "C",
		Components: make(map[string]float64),
	}
	
	cmd := exec.Command("osx-cpu-temp")
	output, err := cmd.Output()
	if err == nil {
		tempRegex := regexp.MustCompile(`(\d+\.\d+)°C`)
		if match := tempRegex.FindStringSubmatch(string(output)); len(match) > 1 {
			if temp, err := strconv.ParseFloat(match[1], 64); err == nil {
				info.CPU = temp
			}
		}
	}
	
	return info
}

func getWindowsTemperatures() models.TemperatureInfo {
	info := models.TemperatureInfo{
		Units:      "C",
		Components: make(map[string]float64),
	}
	
	cmd := exec.Command("wmic", "/namespace:\\\\root\\wmi", "PATH", "MSAcpi_ThermalZoneTemperature", "GET", "CurrentTemperature")
	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		if len(lines) > 1 {
			for i := 1; i < len(lines); i++ {
				line := strings.TrimSpace(lines[i])
				if line == "" {
					continue
				}
				
				tempKelvin, err := strconv.ParseFloat(line, 64)
				if err == nil {
					tempCelsius := (tempKelvin / 10.0) - 273.15
					info.CPU = tempCelsius
					break
				}
			}
		}
	}
	
	return info
}

// Alert and history functionality
func GetTemperatureAlerts(opts models.Options) []models.Alert {
	info := collectTemperatureInfo()
	var alerts []models.Alert
	
	// CPU temperature thresholds
	if info.CPU > 0 {
		const cpuWarnThreshold = 70.0
		const cpuCritThreshold = 85.0
		
		if info.CPU >= cpuCritThreshold {
			alerts = append(alerts, models.Alert{
				Level:     models.Critical,
				Title:     "Critical CPU Temperature",
				Message:   fmt.Sprintf("CPU temperature is at %.1f°%s, exceeding critical threshold of %.1f°%s", info.CPU, info.Units, cpuCritThreshold, info.Units),
				Resource:  "CPU",
				Value:     info.CPU,
				Threshold: cpuCritThreshold,
				Time:      time.Now(),
			})
		} else if info.CPU >= cpuWarnThreshold {
			alerts = append(alerts, models.Alert{
				Level:     models.Warning,
				Title:     "High CPU Temperature",
				Message:   fmt.Sprintf("CPU temperature is at %.1f°%s, exceeding warning threshold of %.1f°%s", info.CPU, info.Units, cpuWarnThreshold, info.Units),
				Resource:  "CPU",
				Value:     info.CPU,
				Threshold: cpuWarnThreshold,
				Time:      time.Now(),
			})
		}
	}
	
	// GPU temperature thresholds
	if info.GPU > 0 {
		const gpuWarnThreshold = 80.0
		const gpuCritThreshold = 95.0
		
		if info.GPU >= gpuCritThreshold {
			alerts = append(alerts, models.Alert{
				Level:     models.Critical,
				Title:     "Critical GPU Temperature",
				Message:   fmt.Sprintf("GPU temperature is at %.1f°%s, exceeding critical threshold of %.1f°%s", info.GPU, info.Units, gpuCritThreshold, info.Units),
				Resource:  "GPU",
				Value:     info.GPU,
				Threshold: gpuCritThreshold,
				Time:      time.Now(),
			})
		} else if info.GPU >= gpuWarnThreshold {
			alerts = append(alerts, models.Alert{
				Level:     models.Warning,
				Title:     "High GPU Temperature",
				Message:   fmt.Sprintf("GPU temperature is at %.1f°%s, exceeding warning threshold of %.1f°%s", info.GPU, info.Units, gpuWarnThreshold, info.Units),
				Resource:  "GPU",
				Value:     info.GPU,
				Threshold: gpuWarnThreshold,
				Time:      time.Now(),
			})
		}
	}
	
	return alerts
}

// StoreTemperatureHistory stores temperature data for historical tracking
func StoreTemperatureHistory(info models.TemperatureInfo) {
	record := models.TemperatureHistoryRecord{
		Timestamp:  time.Now(),
		CPU:        info.CPU,
		GPU:        info.GPU,
		Components: info.Components,
	}
	
	temperatureHistory = append(temperatureHistory, record)
	
	// Keep only last 24 hours of data (1440 minutes)
	if len(temperatureHistory) > 1440 {
		temperatureHistory = temperatureHistory[len(temperatureHistory)-1440:]
	}
}

// GetTemperatureGraph generates a graph of temperature history
func GetTemperatureGraph(width, height int) string {
	if len(temperatureHistory) < 2 {
		return "Not enough temperature data available for graphing"
	}
	
	cpuData := make([]float64, len(temperatureHistory))
	for i, record := range temperatureHistory {
		cpuData[i] = record.CPU
	}
	
	return ui.RenderLineGraph(cpuData, width, height, "CPU Temperature (°C)")
}

// GetTemperatureHistory returns the stored temperature history
func GetTemperatureHistory() []models.TemperatureHistoryRecord {
	return temperatureHistory
}
