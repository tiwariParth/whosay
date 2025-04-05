package cmd

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/tiwariParth/whosay/config"
	"github.com/tiwariParth/whosay/internal/collectors"
	"github.com/tiwariParth/whosay/internal/models"
	"github.com/tiwariParth/whosay/internal/ui"
)

// Execute is the entry point for the CLI application
func Execute() {
	// Initialize configuration
	cfg := config.NewConfig()

	// Define flags
	cpuFlag := flag.Bool("cpu", false, "Display CPU information")
	memFlag := flag.Bool("mem", false, "Display memory information")
	diskFlag := flag.Bool("disk", false, "Display disk information")
	sysFlag := flag.Bool("sys", false, "Display detailed system information")
	netFlag := flag.Bool("net", false, "Display network information")
	netTrafficFlag := flag.Bool("nettraffic", false, "Display network traffic information")
	procFlag := flag.Bool("proc", false, "Display process information")
	dockerFlag := flag.Bool("docker", false, "Display Docker container information")
	dockerLogsFlag := flag.String("container-logs", "", "Display logs for a Docker container (provide container ID or name)")
	logsLimitFlag := flag.Int("logs-limit", 50, "Limit the number of log lines to display")
	batteryFlag := flag.Bool("battery", false, "Display battery information")
	tempFlag := flag.Bool("temp", false, "Display temperature information")
	logsFlag := flag.Bool("logs", false, "Display system logs")
	historyFlag := flag.Bool("history", false, "Show resource usage history")
	alertsFlag := flag.Bool("alerts", false, "Display and enable resource alerts")
	allFlag := flag.Bool("all", false, "Display all system information")
	jsonFlag := flag.Bool("json", false, "Output in JSON format")
	verboseFlag := flag.Bool("verbose", false, "Show more detailed information")
	versionFlag := flag.Bool("version", false, "Display version information")
	noColorFlag := flag.Bool("no-color", false, "Disable colorized output")
	watchFlag := flag.Bool("watch", false, "Enable watch mode for continuous monitoring")
	refreshRateFlag := flag.Int("refresh", 1, "Refresh rate in seconds for watch mode (default: 1)")
	
	flag.Parse()
	
	// Disable color if requested
	if *noColorFlag {
		color.NoColor = true
	}

	// Handle version flag
	if *versionFlag {
		fmt.Printf("whosay version %s\n", cfg.Version)
		os.Exit(0)
	}

	// Special handling for container logs command
	if *dockerLogsFlag != "" {
		opts := models.Options{
			JSONOutput:    *jsonFlag,
			VerboseOutput: *verboseFlag,
		}
		collectors.GetContainerLogs(*dockerLogsFlag, *logsLimitFlag, opts)
		return
	}

	// If no specific flags are provided, show help
	if !(*cpuFlag || *memFlag || *diskFlag || *sysFlag || *netFlag || *netTrafficFlag || *procFlag || 
	     *dockerFlag || *batteryFlag || *tempFlag || *logsFlag || *historyFlag || *alertsFlag || *allFlag) {
		flag.Usage()
		os.Exit(1)
	}

	// Create options struct for commands
	opts := models.Options{
		JSONOutput:    *jsonFlag,
		InWatchMode:   *watchFlag,
		VerboseOutput: *verboseFlag,
		EnableAlerts:  *alertsFlag,
	}

	// Watch mode is incompatible with JSON output
	if *watchFlag && *jsonFlag {
		fmt.Println("Error: Watch mode is not compatible with JSON output")
		os.Exit(1)
	}

	// Ensure refresh rate is at least 1 second
	refreshRate := *refreshRateFlag
	if refreshRate < 1 {
		refreshRate = 1
	}

    // One-time execution mode
    if (!*watchFlag) {
        displayInfo(opts, *cpuFlag, *memFlag, *diskFlag, *sysFlag, *netFlag, *netTrafficFlag, *procFlag, 
                   *dockerFlag, *batteryFlag, *tempFlag, *logsFlag, *historyFlag, *alertsFlag, *allFlag, *jsonFlag, cfg)
        
        // Show footer only in text mode
		if !*jsonFlag {
			fmt.Println()
			footerColor := color.New(color.FgHiBlue, color.Italic)
			footerColor.Println("Thanks for using whosay! Stay resourceful! 🚀")
		}
		return
	} else {
		// Use the watch mode function
		runWatchMode(opts, *cpuFlag, *memFlag, *diskFlag, *sysFlag, *netFlag, *netTrafficFlag, *procFlag, 
		            *dockerFlag, *batteryFlag, *tempFlag, *logsFlag, *historyFlag, *alertsFlag, *allFlag, refreshRate)
	}
}

// displayInfo handles displaying the requested system information
func displayInfo(opts models.Options, cpu, mem, disk, sys, net, netTraffic, proc, docker, battery, temp, logs, history, alerts, all, json bool, cfg *config.Config) {
    if json {
        // Execute commands individually for JSON output
        if sys || all {
            collectors.GetSystemInfo(opts)
        }
        
        if cpu || all {
            collectors.GetCPUInfo(opts)
        }
        
        if mem || all {
            collectors.GetMemoryInfo(opts)
        }
        
        if disk || all {
            collectors.GetDiskInfo(opts)
        }
        
        if net || all {
            collectors.GetNetworkInfo(opts)
        }
        
        if netTraffic || all {
            collectors.GetNetworkTrafficInfo(opts)
        }
        
        if proc || all {
            collectors.GetProcessInfo(opts)
        }
        
        if docker || all {
            collectors.GetDockerInfo(opts)
        }
        
        if battery || all {
            collectors.GetBatteryInfo(opts)
        }
        
        if temp || all {
            collectors.GetTemperatureInfo(opts)
        }
        
        if logs || all {
            collectors.GetLogInfo(opts)
        }
        
        if history || all {
            // Return usage history
            fmt.Println("[]") // Simple empty array for now
        }
        
        if alerts {
            // Alerts not supported in JSON mode
            fmt.Println("[]")
        }
        
        return
    }
    
    // Collect all data sections
    allSections := collectDisplaySections(opts, cpu, mem, disk, sys, net, netTraffic, proc, docker, battery, temp, logs, history, alerts, all)
    
    // Display all sections in a unified compact view
    ui.CompactDisplay(allSections)
    
    // Show developer-friendly footer only in text mode
    if !json {
        fmt.Println()
        footerText := fmt.Sprintf(" whosay v%s | Use '-watch' for live monitoring | Press Ctrl+C to exit ", cfg.Version)
        fmt.Println(color.HiBlueString(footerText))
        fmt.Println()
    }
}

// runWatchMode is the watch mode execution loop
func runWatchMode(opts models.Options, cpuFlag, memFlag, diskFlag, sysFlag, netFlag, netTrafficFlag, procFlag, dockerFlag, batteryFlag, tempFlag, logsFlag, historyFlag, alertsFlag, allFlag bool, refreshRate int) {
    // No intro message to avoid flicker
    for {
        ui.ClearScreen()
        
        // Just show basic timing info in a simple format - more compact to save space
        now := time.Now().Format("2006-01-02 15:04:05")
        
        // Add a colored header bar for watch mode
        width := ui.GetTerminalWidth()
        if width > 120 {
            width = 120
        }
        
        headerText := fmt.Sprintf(" WATCH MODE | Refresh: %ds | %s | Press Ctrl+C to exit ", refreshRate, now)
        fmt.Println(color.HiBlueString(strings.Repeat("─", width)))
        fmt.Println(color.New(color.FgHiWhite, color.BgBlue).Sprint(headerText))
        fmt.Println(color.HiBlueString(strings.Repeat("─", width)))
        
        // In watch mode we need to simplify display
        watchOpts := opts
        watchOpts.CompactMode = true // Add extra compression for watch mode
        
        // Capture the data sections before display
        sections := collectDisplaySections(watchOpts, cpuFlag, memFlag, diskFlag, sysFlag, netFlag, netTrafficFlag, procFlag, dockerFlag, batteryFlag, tempFlag, logsFlag, historyFlag, alertsFlag, allFlag)
        
        // Display all sections in a unified view
        ui.CompactDisplay(sections)
        
        // Consistent sleep interval
        time.Sleep(time.Duration(refreshRate) * time.Second)
    }
}

// collectDisplaySections gathers all the sections to be displayed
func collectDisplaySections(opts models.Options, cpu, mem, disk, sys, net, netTraffic, proc, docker, battery, temp, logs, history, alerts, all bool) map[string][][]string {
    allSections := make(map[string][][]string)
    
    if sys || all {
        systemSections := collectors.GetSystemInfoSections(opts)
        for k, v := range systemSections {
            allSections[k] = v
        }
    }
    
    if cpu || all {
        cpuSections := collectors.GetCPUInfoSections(opts)
        for k, v := range cpuSections {
            allSections[k] = v
        }
    }
    
    if mem || all {
        memSections := collectors.GetMemoryInfoSections(opts)
        for k, v := range memSections {
            allSections[k] = v
        }
    }
    
    if disk || all {
        diskSections := collectors.GetDiskInfoSections(opts)
        for k, v := range diskSections {
            allSections[k] = v
        }
    }
    
    if net || all {
        networkSections := collectors.GetNetworkInfoSections(opts)
        for k, v := range networkSections {
            allSections[k] = v
        }
    }
    
    if netTraffic || all {
        trafficSections := collectors.GetNetworkTrafficInfoSections(opts)
        for k, v := range trafficSections {
            allSections[k] = v
        }
    }
    
    if proc || all {
        processes, err := collectors.GetTopProcesses(10)
        if err == nil {
            processSections := collectors.GetProcessInfoSections(processes, opts)
            for k, v := range processSections {
                allSections[k] = v
            }
        }
    }
    
    if docker || all {
        containers, err := collectors.GetDockerContainers()
        if err == nil {
            dockerSections := collectors.GetDockerInfoSections(containers, opts)
            for k, v := range dockerSections {
                allSections[k] = v
            }
        }
    }
    
    if battery || all {
        batterySections := collectors.GetBatteryInfoSections(opts)
        for k, v := range batterySections {
            allSections[k] = v
        }
    }
    
    if temp || all {
        temperatureSections := collectors.GetTemperatureInfoSections(opts)
        for k, v := range temperatureSections {
            allSections[k] = v
        }
    }
    
    if logs || all {
        logSections := collectors.GetLogInfoSections(opts)
        for k, v := range logSections {
            allSections[k] = v
        }
    }
    
    if history || all {
        // Create a history section
        historySections := getResourceHistorySections(opts)
        for k, v := range historySections {
            allSections[k] = v
        }
    }
    
    if alerts {
        alertsSections := collectors.GetAlertsInfoSections(opts)
        for k, v := range alertsSections {
            allSections[k] = v
        }
    }
    
    return allSections
}

// getResourceHistorySections creates sections displaying resource usage history
func getResourceHistorySections(opts models.Options) map[string][][]string {
    result := map[string][][]string{
        "Resource History": {
            {"Status", "Resource history tracking is enabled"},
            {"Period", "Last 24 hours"},
        },
    }
    
    // Add a CPU history graph
    cpuHistory := []float64{30, 35, 28, 40, 45, 50, 48, 35, 30, 25} // Demo data
    result["CPU History"] = [][]string{
        {"", ui.RenderLineGraph(cpuHistory, 60, 10, "")},
    }
    
    // Add a memory history graph
    memHistory := []float64{50, 55, 58, 60, 65, 70, 65, 62, 60, 55} // Demo data
    result["Memory History"] = [][]string{
        {"", ui.RenderLineGraph(memHistory, 60, 10, "")},
    }
    
    // If temperature history is available, add it
    if len(collectors.GetTemperatureHistory()) > 0 {
        result["Temperature History"] = [][]string{
            {"", collectors.GetTemperatureGraph(60, 10)},
        }
    }
    
    return result
}
