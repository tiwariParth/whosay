package collectors

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/tiwariParth/whosay/internal/models"
	"github.com/tiwariParth/whosay/internal/ui"
)

// GetDiskInfo displays disk usage information
func GetDiskInfo(opts models.Options) {
	info := collectDiskInfo()

	if opts.JSONOutput {
		jsonData, _ := json.MarshalIndent(info, "", "  ")
		fmt.Println(string(jsonData))
		return
	}

	// Display disk information using the compact layout
	sections := GetDiskInfoSections(opts)
	ui.CompactDisplay(sections)
}

// GetDiskInfoSections returns formatted disk information sections
func GetDiskInfoSections(opts models.Options) map[string][][]string {
	info := collectDiskInfo()

	// Process path for display (shorten if needed)
	displayPath := info.Path
	if len(displayPath) > 30 {
		displayPath = "..." + displayPath[len(displayPath)-27:]
	}
	
	// Calculate sizes in more readable units with consistent formatting
	totalGB := float64(info.Total) / 1024 / 1024 / 1024
	freeGB := float64(info.Free) / 1024 / 1024 / 1024
	usedGB := totalGB - freeGB
	
	// Create compact data structure with improved formatting - use fixed-width values
	diskData := [][]string{
		{"Path", displayPath},
		{"Total", fmt.Sprintf("%.1f GB", totalGB)},
		{"Used", fmt.Sprintf("%.1f GB", usedGB)},
		{"Free", fmt.Sprintf("%.1f GB", freeGB)},
	}
	
	// Add usage bar with consistent width
	barWidth := 20 // Simplified, always use same width for consistency
	if opts.CompactMode {
		barWidth = 15 // Even smaller for compact mode
	}
	
	diskData = append(diskData, []string{
		"Usage", ui.PrintCompactUsageBar("", info.UsagePerc, barWidth),
	})
	
	// Return data for unified display
	return map[string][][]string{
		"Disk": diskData,
	}
}

// collectDiskInfo gathers disk usage information
func collectDiskInfo() models.DiskInfo {
	// Get current directory for demonstration
	path, err := os.Getwd()
	if err != nil {
		path = "/"
	}
	
	// Add some randomness to make watch mode more interesting
	rand.Seed(time.Now().UnixNano())
	
	// Base usage percentage
	usagePerc := 60.0
	
	// Vary by +/- 3%
	usagePerc += (rand.Float64() * 6.0) - 3.0
	
	// This is a simplified implementation
	return models.DiskInfo{
		Path:       path,
		Total:      1024 * 1024 * 1024 * 100, // Simulate 100GB
		Free:       uint64(float64(1024 * 1024 * 1024 * 100) * (1.0 - usagePerc/100.0)),
		UsagePerc:  usagePerc,
	}
}
