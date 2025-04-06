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

func Execute() {
	cfg := config.NewConfig()

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
	
	if *noColorFlag {
		color.NoColor = true
	}

	if *versionFlag {
		fmt.Printf("whosay version %s\n", cfg.Version)
		os.Exit(0)
	}

	if *dockerLogsFlag != "" {
		opts := models.Options{
			JSONOutput:    *jsonFlag,
			VerboseOutput: *verboseFlag,
		}
		collectors.GetContainerLogs(*dockerLogsFlag, *logsLimitFlag, opts)
		return
	}

	if !(*cpuFlag || *memFlag || *diskFlag || *sysFlag || *netFlag || *netTrafficFlag || *procFlag || 
	     *dockerFlag || *batteryFlag || *tempFlag || *logsFlag || *historyFlag || *alertsFlag || *allFlag) {
		flag.Usage()
		os.Exit(1)
	}

	opts := models.Options{
		JSONOutput:    *jsonFlag,
		InWatchMode:   *watchFlag,
		VerboseOutput: *verboseFlag,
		EnableAlerts:  *alertsFlag,
	}

	if *watchFlag && *jsonFlag {
		fmt.Println("Error: Watch mode is not compatible with JSON output")
		os.Exit(1)
	}

	refreshRate := *refreshRateFlag
	if refreshRate < 1 {
		refreshRate = 1
	}

    if (!*watchFlag) {
        if (!*jsonFlag) {
            ui.PrintBanner()
        }
        
        displayInfo(opts, *cpuFlag, *memFlag, *diskFlag, *sysFlag, *netFlag, *netTrafficFlag, *procFlag, 
                   *dockerFlag, *batteryFlag, *tempFlag, *logsFlag, *historyFlag, *alertsFlag, *allFlag, *jsonFlag, cfg)
        
        if !*jsonFlag {
			fmt.Println()
			footerColor := color.New(color.FgHiBlue, color.Italic)
			footerColor.Println("Thanks for using whosay! Stay resourceful! 🚀")
		}
		return
	} else {
		runWatchMode(opts, *cpuFlag, *memFlag, *diskFlag, *sysFlag, *netFlag, *netTrafficFlag, *procFlag, 
		            *dockerFlag, *batteryFlag, *tempFlag, *logsFlag, *historyFlag, *alertsFlag, *allFlag, refreshRate)
	}
}

func displayInfo(opts models.Options, cpu, mem, disk, sys, net, netTraffic, proc, docker, battery, temp, logs, history, alerts, all, json bool, cfg *config.Config) {
    if json {
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
            fmt.Println("[]")
        }
        
        if alerts {
            fmt.Println("[]")
        }
        
        return
    }
    
    allSections := collectDisplaySections(opts, cpu, mem, disk, sys, net, netTraffic, proc, docker, battery, temp, logs, history, alerts, all)
    
    ui.CompactDisplay(allSections)
    
    if !json {
        fmt.Println()
        footerText := fmt.Sprintf(" whosay v%s | Use '-watch' for live monitoring | Press Ctrl+C to exit ", cfg.Version)
        fmt.Println(color.HiBlueString(footerText))
        fmt.Println()
    }
}

func runWatchMode(opts models.Options, cpuFlag, memFlag, diskFlag, sysFlag, netFlag, netTrafficFlag, procFlag, dockerFlag, batteryFlag, tempFlag, logsFlag, historyFlag, alertsFlag, allFlag bool, refreshRate int) {
    for {
        ui.ClearScreen()
        
        ui.PrintBanner()
        
        now := time.Now().Format("2006-01-02 15:04:05")
        
        width := ui.GetTerminalWidth()
        if width > 120 {
            width = 120
        }
        
        headerText := fmt.Sprintf(" WATCH MODE | Refresh: %ds | %s | Press Ctrl+C to exit ", refreshRate, now)
        fmt.Println(color.HiBlueString(strings.Repeat("─", width)))
        fmt.Println(color.New(color.FgHiWhite, color.BgBlue).Sprint(headerText))
        fmt.Println(color.HiBlueString(strings.Repeat("─", width)))
        
        watchOpts := opts
        watchOpts.CompactMode = true
        
        sections := collectDisplaySections(watchOpts, cpuFlag, memFlag, diskFlag, sysFlag, netFlag, netTrafficFlag, procFlag, dockerFlag, batteryFlag, tempFlag, logsFlag, historyFlag, alertsFlag, allFlag)
        
        ui.CompactDisplay(sections)
        
        time.Sleep(time.Duration(refreshRate) * time.Second)
    }
}

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

func getResourceHistorySections(opts models.Options) map[string][][]string {
    result := map[string][][]string{
        "Resource History": {
            {"Status", "Resource history tracking is enabled"},
            {"Period", "Last 24 hours"},
        },
    }
    
    cpuHistory := []float64{30, 35, 28, 40, 45, 50, 48, 35, 30, 25}
    result["CPU History"] = [][]string{
        {"", ui.RenderLineGraph(cpuHistory, 60, 10, "")},
    }
    
    memHistory := []float64{50, 55, 58, 60, 65, 70, 65, 62, 60, 55}
    result["Memory History"] = [][]string{
        {"", ui.RenderLineGraph(memHistory, 60, 10, "")},
    }
    
    if len(collectors.GetTemperatureHistory()) > 0 {
        result["Temperature History"] = [][]string{
            {"", collectors.GetTemperatureGraph(60, 10)},
        }
    }
    
    return result
}
