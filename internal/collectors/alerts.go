package collectors

import (
	"fmt"

	"github.com/tiwariParth/whosay/internal/alerts"
	"github.com/tiwariParth/whosay/internal/models"
	"github.com/tiwariParth/whosay/internal/ui"
)

// Define alertManager variable
var alertManager *alerts.AlertManager

// Initialize the alert manager
func init() {
	alertManager = alerts.NewAlertManager()
}

// GetAlertsInfo displays alerts information
func GetAlertsInfo(opts models.Options) {
	if !opts.EnableAlerts {
		fmt.Println("Alerts are disabled. Use the --alerts flag to enable alerts.")
		return
	}

	// Display alerts information using the compact layout
	sections := GetAlertsInfoSections(opts)
	ui.CompactDisplay(sections)
}

// GetAlertsInfoSections returns formatted alerts information
func GetAlertsInfoSections(opts models.Options) map[string][][]string {
	// Create a summary section
	infoCount, warningCount, criticalCount := alertManager.GetAlertsSummary()
	
	alertSummary := [][]string{
		{"Critical", fmt.Sprintf("%d", criticalCount)},
		{"Warning", fmt.Sprintf("%d", warningCount)},
		{"Info", fmt.Sprintf("%d", infoCount)},
		{"Total", fmt.Sprintf("%d", infoCount + warningCount + criticalCount)},
	}
	
	// Create a thresholds section
	thresholds := alertManager.Thresholds
	thresholdSettings := [][]string{
		{"CPU Warning", fmt.Sprintf("%.1f%%", thresholds.CPUWarning)},
		{"CPU Critical", fmt.Sprintf("%.1f%%", thresholds.CPUCritical)},
		{"Memory Warning", fmt.Sprintf("%.1f%%", thresholds.MemoryWarning)},
		{"Memory Critical", fmt.Sprintf("%.1f%%", thresholds.MemoryCritical)},
		{"Disk Warning", fmt.Sprintf("%.1f%%", thresholds.DiskWarning)},
		{"Disk Critical", fmt.Sprintf("%.1f%%", thresholds.DiskCritical)},
	}
	
	// Create a section for recent alerts
	recentAlerts := [][]string{
		{"Time", "Level", "Resource", "Message"},
	}
	
	// Get active (non-acknowledged) alerts
	activeAlerts := alertManager.GetActiveAlerts()
	
	// Add recent alerts to the table
	alertCount := 0
	for _, alert := range activeAlerts {
		if alertCount >= 10 {
			break // Limit to 10 alerts in the display
		}
		
		// Format alert level
		levelStr := "INFO"
		if alert.Level == 1 {
			levelStr = "WARNING"
		} else if alert.Level == 2 {
			levelStr = "CRITICAL"
		}
		
		// Format time
		timeStr := alert.Time.Format("15:04:05")
		
		// Truncate message if too long
		message := alert.Message
		if len(message) > 50 {
			message = message[:47] + "..."
		}
		
		recentAlerts = append(recentAlerts, []string{
			timeStr,
			levelStr,
			alert.Resource,
			message,
		})
		
		alertCount++
	}
	
	// If no alerts, add a message
	if alertCount == 0 {
		recentAlerts = append(recentAlerts, []string{
			"", "", "", "No active alerts",
		})
	}
	
	// Build the final sections map
	result := map[string][][]string{
		"Alert Summary": alertSummary,
		"Alert Thresholds": thresholdSettings,
	}
	
	// Only add recent alerts if there are any
	if alertCount > 0 {
		result["Recent Alerts"] = recentAlerts
	}
	
	return result
}

// ConfigureAlertThresholds updates the alert thresholds
func ConfigureAlertThresholds(cpuWarn, cpuCrit, memWarn, memCrit, diskWarn, diskCrit float64) {
	// Validate thresholds (ensure warning is less than critical)
	if cpuWarn >= cpuCrit {
		cpuWarn = cpuCrit - 10
	}
	
	if memWarn >= memCrit {
		memWarn = memCrit - 10
	}
	
	if diskWarn >= diskCrit {
		diskWarn = diskCrit - 10
	}
	
	// Update thresholds
	alertManager.Thresholds = alerts.ThresholdConfig{
		CPUWarning:     cpuWarn,
		CPUCritical:    cpuCrit,
		MemoryWarning:  memWarn,
		MemoryCritical: memCrit,
		DiskWarning:    diskWarn,
		DiskCritical:   diskCrit,
	}
	
	// Add an informational alert about the update
	alertManager.AddAlert(
		alerts.Info,
		"Alert Thresholds Updated",
		"The alert thresholds have been updated with new values.",
		"System",
		0,
		0,
	)
}

// AcknowledgeAllAlerts marks all alerts as acknowledged
func AcknowledgeAllAlerts() {
	for i := range alertManager.Alerts {
		alertManager.Alerts[i].Acknowledged = true
	}
	
	// Add an informational alert
	alertManager.AddAlert(
		alerts.Info,
		"Alerts Acknowledged",
		"All alerts have been acknowledged.",
		"System",
		0,
		0,
	)
}
