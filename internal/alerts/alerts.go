package alerts

import (
	"fmt"
	"time"

	"github.com/fatih/color"
)

// AlertLevel defines the severity of an alert
type AlertLevel int

const (
	// Info is for informational alerts
	Info AlertLevel = iota
	// Warning is for concerning but non-critical alerts
	Warning
	// Critical is for urgent alerts requiring immediate attention
	Critical
)

// Alert represents a monitoring alert
type Alert struct {
	Level       AlertLevel
	Title       string
	Message     string
	Resource    string
	Value       float64
	Threshold   float64
	Time        time.Time
	Acknowledged bool
}

// ThresholdConfig defines alert thresholds for resources
type ThresholdConfig struct {
	CPUWarning     float64
	CPUCritical    float64
	MemoryWarning  float64
	MemoryCritical float64
	DiskWarning    float64
	DiskCritical   float64
}

// DefaultThresholds returns sensible default threshold values
func DefaultThresholds() ThresholdConfig {
	return ThresholdConfig{
		CPUWarning:     75.0,
		CPUCritical:    90.0,
		MemoryWarning:  80.0,
		MemoryCritical: 95.0,
		DiskWarning:    85.0,
		DiskCritical:   95.0,
	}
}

// AlertManager handles creation and storage of alerts
type AlertManager struct {
	Alerts       []Alert
	Thresholds   ThresholdConfig
	MaxAlerts    int
	EnableAlerts bool
}

// NewAlertManager creates a new alert manager with default settings
func NewAlertManager() *AlertManager {
	return &AlertManager{
		Alerts:       make([]Alert, 0),
		Thresholds:   DefaultThresholds(),
		MaxAlerts:    100,
		EnableAlerts: true,
	}
}

// AddAlert adds a new alert to the manager
func (am *AlertManager) AddAlert(level AlertLevel, title, message, resource string, value, threshold float64) {
	if !am.EnableAlerts {
		return
	}

	alert := Alert{
		Level:       level,
		Title:       title,
		Message:     message,
		Resource:    resource,
		Value:       value,
		Threshold:   threshold,
		Time:        time.Now(),
		Acknowledged: false,
	}

	// Add to beginning of slice for reverse chronological order
	am.Alerts = append([]Alert{alert}, am.Alerts...)

	// Truncate if we exceed the maximum number of alerts
	if len(am.Alerts) > am.MaxAlerts {
		am.Alerts = am.Alerts[:am.MaxAlerts]
	}

	// Print the alert
	am.PrintAlert(alert)
}

// CheckCPU creates alerts for CPU usage if needed
func (am *AlertManager) CheckCPU(usage float64) {
	if usage >= am.Thresholds.CPUCritical {
		am.AddAlert(
			Critical,
			"Critical CPU Usage",
			fmt.Sprintf("CPU usage is at %.1f%%, exceeding the critical threshold of %.1f%%", usage, am.Thresholds.CPUCritical),
			"CPU",
			usage,
			am.Thresholds.CPUCritical,
		)
	} else if usage >= am.Thresholds.CPUWarning {
		am.AddAlert(
			Warning,
			"High CPU Usage",
			fmt.Sprintf("CPU usage is at %.1f%%, exceeding the warning threshold of %.1f%%", usage, am.Thresholds.CPUWarning),
			"CPU",
			usage,
			am.Thresholds.CPUWarning,
		)
	}
}

// CheckMemory creates alerts for memory usage if needed
func (am *AlertManager) CheckMemory(usage float64) {
	if usage >= am.Thresholds.MemoryCritical {
		am.AddAlert(
			Critical,
			"Critical Memory Usage",
			fmt.Sprintf("Memory usage is at %.1f%%, exceeding the critical threshold of %.1f%%", usage, am.Thresholds.MemoryCritical),
			"Memory",
			usage,
			am.Thresholds.MemoryCritical,
		)
	} else if usage >= am.Thresholds.MemoryWarning {
		am.AddAlert(
			Warning,
			"High Memory Usage",
			fmt.Sprintf("Memory usage is at %.1f%%, exceeding the warning threshold of %.1f%%", usage, am.Thresholds.MemoryWarning),
			"Memory",
			usage,
			am.Thresholds.MemoryWarning,
		)
	}
}

// CheckDisk creates alerts for disk usage if needed
func (am *AlertManager) CheckDisk(usage float64, path string) {
	if usage >= am.Thresholds.DiskCritical {
		am.AddAlert(
			Critical,
			"Critical Disk Usage",
			fmt.Sprintf("Disk usage for %s is at %.1f%%, exceeding the critical threshold of %.1f%%", path, usage, am.Thresholds.DiskCritical),
			"Disk",
			usage,
			am.Thresholds.DiskCritical,
		)
	} else if usage >= am.Thresholds.DiskWarning {
		am.AddAlert(
			Warning,
			"High Disk Usage",
			fmt.Sprintf("Disk usage for %s is at %.1f%%, exceeding the warning threshold of %.1f%%", path, usage, am.Thresholds.DiskWarning),
			"Disk",
			usage,
			am.Thresholds.DiskWarning,
		)
	}
}

// GetAlertsByLevel returns alerts filtered by level
func (am *AlertManager) GetAlertsByLevel(level AlertLevel) []Alert {
	filtered := make([]Alert, 0)
	for _, alert := range am.Alerts {
		if alert.Level == level {
			filtered = append(filtered, alert)
		}
	}
	return filtered
}

// GetActiveAlerts returns all non-acknowledged alerts
func (am *AlertManager) GetActiveAlerts() []Alert {
	active := make([]Alert, 0)
	for _, alert := range am.Alerts {
		if !alert.Acknowledged {
			active = append(active, alert)
		}
	}
	return active
}

// AcknowledgeAlert marks an alert as acknowledged
func (am *AlertManager) AcknowledgeAlert(index int) {
	if index >= 0 && index < len(am.Alerts) {
		am.Alerts[index].Acknowledged = true
	}
}

// PrintAlert displays an alert to the console
func (am *AlertManager) PrintAlert(alert Alert) {
	var prefix string
	var alertColor *color.Color
	
	switch alert.Level {
	case Info:
		prefix = "INFO"
		alertColor = color.New(color.FgHiBlue)
	case Warning:
		prefix = "WARNING"
		alertColor = color.New(color.FgHiYellow)
	case Critical:
		prefix = "CRITICAL"
		alertColor = color.New(color.FgHiRed, color.Bold)
	}
	
	timeStr := alert.Time.Format("15:04:05")
	alertColor.Printf("\n[%s] %s - %s\n", prefix, timeStr, alert.Title)
	fmt.Printf("  %s\n\n", alert.Message)
}

// GetAlertsSummary returns alert counts by severity level
func (am *AlertManager) GetAlertsSummary() (int, int, int) {
	infoCount := 0
	warningCount := 0
	criticalCount := 0
	
	for _, alert := range am.Alerts {
		if !alert.Acknowledged {
			switch alert.Level {
			case Info:
				infoCount++
			case Warning:
				warningCount++
			case Critical:
				criticalCount++
			}
		}
	}
	
	return infoCount, warningCount, criticalCount
}
