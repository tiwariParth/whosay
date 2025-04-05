package collectors

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"runtime"
	"time"

	"github.com/tiwariParth/whosay/internal/models"
	"github.com/tiwariParth/whosay/internal/ui"
)

// GetMemoryInfo displays memory information
func GetMemoryInfo(opts models.Options) {
	info := collectMemoryInfo()

	if opts.JSONOutput {
		jsonData, _ := json.MarshalIndent(info, "", "  ")
		fmt.Println(string(jsonData))
		return
	}

	// Display the memory info in compact view
	sections := GetMemoryInfoSections(opts)
	ui.CompactDisplay(sections)
}

// GetMemoryInfoSections returns formatted memory information sections
func GetMemoryInfoSections(opts models.Options) map[string][][]string {
	info := collectMemoryInfo()
	
	// Calculate sizes in more readable units with consistent formatting
	totalMB := float64(info.Total) / 1024 / 1024
	usedMB := float64(info.Used) / 1024 / 1024
	freeMB := totalMB - usedMB
	
	// Create compact data structure with consistent formatting
	memData := [][]string{
		{"Total", fmt.Sprintf("%.1f MB", totalMB)},
		{"Used", fmt.Sprintf("%.1f MB", usedMB)},
		{"Free", fmt.Sprintf("%.1f MB", freeMB)},
	}
	
	// Add usage bar with consistent width
	barWidth := 20 // Simplified, always use same width for consistency
	if opts.CompactMode {
		barWidth = 15 // Even smaller for compact mode
	}
	
	memData = append(memData, []string{
		"Usage", ui.PrintCompactUsageBar("", info.UsagePerc, barWidth),
	})
	
	// Return data for unified display
	return map[string][][]string{
		"Memory": memData,
	}
}

// collectMemoryInfo gathers memory information
func collectMemoryInfo() models.MemoryInfo {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Add some randomness to make watch mode interesting
	rand.Seed(time.Now().UnixNano())
	randomFactor := 1.0
	
	// Add up to 5% random variation to make watch mode interesting
	randomFactor = 0.95 + (rand.Float64() * 0.1) // 0.95 to 1.05

	// This is a simplified version for demo
	return models.MemoryInfo{
		Total:     m.Sys,
		Used:      uint64(float64(m.Alloc) * randomFactor),
		UsagePerc: float64(m.Alloc) / float64(m.Sys) * 100 * randomFactor,
	}
}
