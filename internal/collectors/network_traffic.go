package collectors

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/tiwariParth/whosay/internal/models"
	"github.com/tiwariParth/whosay/internal/ui"
)

// Traffic rate history length (for historical graph)
const historyLength = 60 // Store last 60 seconds

// Network traffic data store with cached calculations
var (
	lastReadings     map[string]models.NetworkUsageInfo
	lastReadTime     time.Time
	trafficHistory   map[string][]models.NetworkUsageInfo
	networkTrafficMu sync.Mutex
)

func init() {
	lastReadings = make(map[string]models.NetworkUsageInfo)
	trafficHistory = make(map[string][]models.NetworkUsageInfo)
	lastReadTime = time.Now()
}

// GetNetworkTrafficInfo displays network traffic information
func GetNetworkTrafficInfo(opts models.Options) {
	info := GetNetworkUsage()

	if opts.JSONOutput {
		jsonData, err := json.MarshalIndent(info, "", "  ")
		if err != nil {
			fmt.Printf("Error serializing network traffic data: %v\n", err)
			return
		}
		fmt.Println(string(jsonData))
		return
	}

	// Display network traffic info in compact view
	sections := GetNetworkTrafficInfoSections(opts)
	ui.CompactDisplay(sections)
}

// GetNetworkTrafficInfoSections formats network traffic information for the compact display
func GetNetworkTrafficInfoSections(opts models.Options) map[string][][]string {
	usage := GetNetworkUsage()
	
	// Create a traffic summary section
	summaryData := [][]string{
		{"Interfaces", fmt.Sprintf("%d active", len(usage))},
	}
	
	// Calculate total bytes across all interfaces
	var totalRxBytes, totalTxBytes uint64
	var totalRxRate, totalTxRate float64
	
	for _, iface := range usage {
		totalRxBytes += iface.BytesReceived
		totalTxBytes += iface.BytesSent
		totalRxRate += iface.RxRate
		totalTxRate += iface.TxRate
	}
	
	// Format totals in human-readable format
	summaryData = append(summaryData, []string{
		"Total Download", formatBytes(totalRxBytes),
	})
	summaryData = append(summaryData, []string{
		"Total Upload", formatBytes(totalTxBytes),
	})
	summaryData = append(summaryData, []string{
		"Current Rate", fmt.Sprintf("↓ %.1f Mbps / ↑ %.1f Mbps", totalRxRate, totalTxRate),
	})
	
	// Create interface sections
	interfaceSections := make(map[string][][]string)
	
	for _, iface := range usage {
		// Skip interfaces with no activity if not in verbose mode
		if !opts.VerboseOutput && iface.RxRate < 0.001 && iface.TxRate < 0.001 {
			continue
		}
		
		// Skip loopback interface unless in verbose mode
		if (iface.Interface == "lo" || strings.HasPrefix(iface.Interface, "loop")) && !opts.VerboseOutput {
			continue
		}
		
		// Create data for this interface
		ifaceData := [][]string{
			{"Download", fmt.Sprintf("%s (%.1f Mbps)", formatBytes(iface.BytesReceived), iface.RxRate)},
			{"Upload", fmt.Sprintf("%s (%.1f Mbps)", formatBytes(iface.BytesSent), iface.TxRate)},
		}
		
		// Add packet counts
		if opts.VerboseOutput {
			ifaceData = append(ifaceData, []string{
				"Packets", fmt.Sprintf("↓ %d / ↑ %d", iface.PacketsReceived, iface.PacketsSent),
			})
			
			if iface.Errors > 0 {
				ifaceData = append(ifaceData, []string{
					"Errors", fmt.Sprintf("%d", iface.Errors),
				})
			}
		}
		
		// Add little traffic bars for download/upload
		if !opts.CompactMode {
			// Get download/upload history for this interface
			rxHistory, txHistory := getTrafficHistory(iface.Interface)
			
			// Generate ASCII graph for download
			if len(rxHistory) > 0 {
				ifaceData = append(ifaceData, []string{
					"Download History", generateTrafficGraph(rxHistory, 40, 5),
				})
			}
			
			// Generate ASCII graph for upload
			if len(txHistory) > 0 {
				ifaceData = append(ifaceData, []string{
					"Upload History", generateTrafficGraph(txHistory, 40, 5),
				})
			}
		}
		
		interfaceSections[fmt.Sprintf("Interface: %s", iface.Interface)] = ifaceData
	}
	
	// Build the final sections map
	result := map[string][][]string{
		"Network Traffic": summaryData,
	}
	
	// Add interfaces to result
	for name, section := range interfaceSections {
		result[name] = section
	}
	
	return result
}

// GetNetworkUsage returns current network usage information
func GetNetworkUsage() []models.NetworkUsageInfo {
	networkTrafficMu.Lock()
	defer networkTrafficMu.Unlock()
	
	// Get current network stats
	networkStats, err := readNetworkStats()
	if err != nil {
		return []models.NetworkUsageInfo{}
	}
	
	// Calculate time elapsed since last reading
	now := time.Now()
	elapsed := now.Sub(lastReadTime).Seconds()
	if elapsed < 0.1 {
		elapsed = 0.1 // Prevent division by zero or unrealistically small time periods
	}
	
	result := make([]models.NetworkUsageInfo, 0, len(networkStats))
	
	// Process each interface
	for ifaceName, current := range networkStats {
		// Check if we have a previous reading for this interface
		if prev, ok := lastReadings[ifaceName]; ok {
			// Calculate bytes per second
			rxBytesPerSec := float64(current.BytesReceived-prev.BytesReceived) / elapsed
			txBytesPerSec := float64(current.BytesSent-prev.BytesSent) / elapsed
			
			// Convert to megabits per second (8 bits per byte, 1 million bits per megabit)
			rxMbps := (rxBytesPerSec * 8) / 1000000
			txMbps := (txBytesPerSec * 8) / 1000000
			
			// Update with calculated rates
			current.RxRate = rxMbps
			current.TxRate = txMbps
			
			// Update history
			updateTrafficHistory(ifaceName, rxMbps, txMbps)
		}
		
		result = append(result, current)
	}
	
	// Update stored readings and time
	lastReadings = networkStats
	lastReadTime = now
	
	return result
}

// readNetworkStats reads network statistics from /proc/net/dev on Linux
func readNetworkStats() (map[string]models.NetworkUsageInfo, error) {
	result := make(map[string]models.NetworkUsageInfo)
	
	// On Linux, read from /proc/net/dev
	data, err := os.ReadFile("/proc/net/dev")
	if err != nil {
		return result, err
	}
	
	lines := strings.Split(string(data), "\n")
	
	// Skip the first two lines (headers)
	for _, line := range lines[2:] {
		fields := strings.Fields(strings.TrimSpace(line))
		if len(fields) < 17 {
			continue
		}
		
		// Extract interface name (removing the trailing ':')
		ifaceName := strings.TrimSuffix(fields[0], ":")
		
		// Parse statistics
		rxBytes, _ := strconv.ParseUint(fields[1], 10, 64)
		rxPackets, _ := strconv.ParseUint(fields[2], 10, 64)
		rxErrors, _ := strconv.ParseUint(fields[3], 10, 64)
		txBytes, _ := strconv.ParseUint(fields[9], 10, 64)
		txPackets, _ := strconv.ParseUint(fields[10], 10, 64)
		txErrors, _ := strconv.ParseUint(fields[11], 10, 64)
		
		// Store in result
		result[ifaceName] = models.NetworkUsageInfo{
			Interface:        ifaceName,
			BytesReceived:    rxBytes,
			BytesSent:        txBytes,
			PacketsReceived:  rxPackets,
			PacketsSent:      txPackets,
			Errors:           rxErrors + txErrors,
			RxRate:           0, // Will be calculated later
			TxRate:           0, // Will be calculated later
		}
	}
	
	return result, nil
}

// formatBytes converts bytes to a human-readable string (KB, MB, GB)
func formatBytes(bytes uint64) string {
	const (
		_          = iota
		KB float64 = 1 << (10 * iota)
		MB
		GB
		TB
	)
	
	switch {
	case bytes >= uint64(TB):
		return fmt.Sprintf("%.2f TB", float64(bytes)/TB)
	case bytes >= uint64(GB):
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= uint64(MB):
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= uint64(KB):
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// updateTrafficHistory adds new traffic readings to the history
func updateTrafficHistory(ifaceName string, rxMbps, txMbps float64) {
	// Initialize history for this interface if it doesn't exist
	if _, ok := trafficHistory[ifaceName]; !ok {
		trafficHistory[ifaceName] = make([]models.NetworkUsageInfo, 0, historyLength)
	}
	
	// Add current readings
	historyEntry := models.NetworkUsageInfo{
		Interface: ifaceName,
		RxRate:    rxMbps,
		TxRate:    txMbps,
		Timestamp: time.Now(),
	}
	
	// Add to history and maintain limited size
	history := trafficHistory[ifaceName]
	history = append(history, historyEntry)
	if len(history) > historyLength {
		history = history[1:] // Remove oldest entry
	}
	trafficHistory[ifaceName] = history
}

// getTrafficHistory returns historical traffic data for an interface
func getTrafficHistory(ifaceName string) ([]float64, []float64) {
	history, ok := trafficHistory[ifaceName]
	if !ok || len(history) == 0 {
		return []float64{}, []float64{}
	}
	
	// Extract rx and tx rates
	rxHistory := make([]float64, len(history))
	txHistory := make([]float64, len(history))
	
	for i, entry := range history {
		rxHistory[i] = entry.RxRate
		txHistory[i] = entry.TxRate
	}
	
	return rxHistory, txHistory
}

// generateTrafficGraph creates a simple ASCII graph of network traffic
func generateTrafficGraph(data []float64, width, height int) string {
	if len(data) == 0 {
		return "No data"
	}
	
	// Find the maximum value
	maxVal := data[0]
	for _, val := range data {
		if val > maxVal {
			maxVal = val
		}
	}
	
	// If max is too small, set a minimum scale
	if maxVal < 0.1 {
		maxVal = 0.1
	}
	
	// Ensure we have something to draw
	if maxVal == 0 {
		return strings.Repeat("_", width)
	}
	
	// Create the graph
	result := ""
	
	// Use only the most recent data points that fit in the width
	dataPoints := len(data)
	if dataPoints > width {
		data = data[dataPoints-width:]
	}
	
	// Get color function based on traffic patterns
	var colorFunc func(...interface{}) string
	avgTraffic := averageValue(data)
	switch {
	case avgTraffic < 1.0:
		colorFunc = ui.SuccessColor
	case avgTraffic < 10.0:
		colorFunc = ui.WarningColor
	default:
		colorFunc = ui.DangerColor
	}
	
	// Scale to fit in the available height
	scaleFactor := float64(height) / maxVal
	
	// Use simple ASCII bars
	result = ""
	for _, val := range data {
		barHeight := int(val * scaleFactor)
		if barHeight > height {
			barHeight = height
		}
		
		// Use different characters based on the height
		var barChar string
		switch {
		case barHeight <= 0:
			barChar = "_"
		case barHeight < height/3:
			barChar = "▁"
		case barHeight < 2*height/3:
			barChar = "▄"
		default:
			barChar = "█"
		}
		
		result += colorFunc(barChar)
	}
	
	// Add a legend
	result += fmt.Sprintf(" %.1f Mbps", maxVal)
	
	return result
}

// averageValue calculates the average of a slice of float64
func averageValue(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	
	sum := 0.0
	for _, val := range values {
		sum += val
	}
	
	return sum / float64(len(values))
}
