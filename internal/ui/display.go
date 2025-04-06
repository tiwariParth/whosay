package ui

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"

	"github.com/schollz/progressbar/v3"
)

func PrintBanner() {
	banner := `
 __      __.__           _________              __  
/  \    /  \  |__   ____ /   _____/____  ___.__.\ \ 
\   \/\/   /  |  \ /  _ \\_____  \\__  \<   |  | | |
 \        /|   Y  (  <_> )        \/ __ \\___  | | |
  \__/\  / |___|  /\____/_______  (____  / ____| | |
       \/       \/              \/     \/\/      /_/
                                                    
`
	fmt.Println(TitleColor(banner))
}

// PrintTitle prints a decorated section title with more compact design
func PrintTitle(title string) {
	width := len(title) + 6
	// Modern box-style title
	fmt.Printf("\n %s%s%s\n", BoxTopLeft, strings.Repeat(BoxHorizontal, width), BoxTopRight)
	fmt.Printf(" %s  %s  %s\n", BoxVertical, TitleColor(title), BoxVertical)
	fmt.Printf(" %s%s%s\n", BoxBottomLeft, strings.Repeat(BoxHorizontal, width), BoxBottomRight)
}

// PrintKeyValue prints a key-value pair with nice formatting
func PrintKeyValue(key string, value interface{}) {
	fmt.Printf("  %s %s %s\n", 
		AccentColor(BulletPoint),
		LabelColor(fmt.Sprintf("%s:", key)), 
		ValueColor(fmt.Sprintf("%v", value)))
}

// PrintCompactKeyValue prints a key-value pair in a shorter format
func PrintCompactKeyValue(key string, value interface{}, maxKeyWidth int) {
	keyStr := fmt.Sprintf("%s:", key)
	padding := maxKeyWidth - len(keyStr) + 1
	if padding < 1 {
		padding = 1
	}
	
	fmt.Printf("  %s %s%s%s\n", 
		AccentColor(BulletPoint),
		LabelColor(keyStr), 
		strings.Repeat(" ", padding),
		ValueColor(fmt.Sprintf("%v", value)))
}

// PrintUsageBar prints a more compact progress bar for resource usage
func PrintUsageBar(label string, percentage float64) {
	// Choose color based on usage percentage
	var colorFunc func(a ...interface{}) string
	switch {
	case percentage < 60:
		colorFunc = SuccessColor
	case percentage < 80:
		colorFunc = WarningColor
	default:
		colorFunc = DangerColor
	}
	
	fmt.Printf("  %s %s: ", AccentColor(BulletPoint), LabelColor(label))
	
	// Create and render a more compact progress bar
	bar := progressbar.NewOptions(100,
		progressbar.OptionSetWidth(20),
		progressbar.OptionShowCount(),
		progressbar.OptionSetDescription(""),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "█",
			SaucerHead:    "█",
			SaucerPadding: "░",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)
	
	// Set the percentage and render
	bar.Set(int(percentage))
	fmt.Printf(" %s\n", colorFunc(fmt.Sprintf("%.1f%%", percentage)))
}

// PrintCompactUsageBar prints a very compact progress bar for resource usage
func PrintCompactUsageBar(label string, percentage float64, width int) string {
    // Choose color based on usage percentage
    var colorFunc func(a ...interface{}) string
    
    switch {
    case percentage < 60:
        colorFunc = SuccessColor
    case percentage < 85:
        colorFunc = WarningColor
    default:
        colorFunc = DangerColor
    }
    
    // Create a compact visual bar with consistent width
    barWidth := width - 9 // Adjusted to ensure percentage fits
    if barWidth < 5 {
        barWidth = 5
    }
    
    // Ensure percentage is within bounds
    if percentage < 0 {
        percentage = 0
    } else if percentage > 100 {
        percentage = 100
    }
    
    filledWidth := int(float64(barWidth) * percentage / 100)
    if filledWidth > barWidth {
        filledWidth = barWidth
    }
    
    emptyWidth := barWidth - filledWidth
    
    // Use more modern looking bar characters - solid blocks for better visibility
    bar := colorFunc(strings.Repeat("■", filledWidth)) + DimColor(strings.Repeat("□", emptyWidth))
    
    // Format percentage with consistent width and add a visual indicator
    return fmt.Sprintf("%s %s", 
        bar,
        colorFunc(fmt.Sprintf("%5.1f%%", percentage)))
}

// ClearScreen clears the terminal screen
func ClearScreen() {
	var cmd *exec.Cmd
	
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "cls")
	default: // Linux, macOS, etc.
		cmd = exec.Command("clear")
	}
	
	cmd.Stdout = os.Stdout
	_ = cmd.Run()
}

// GetTerminalWidth tries to determine the width of the terminal
func GetTerminalWidth() int {
	// Default to 80 columns if we can't determine
	defaultWidth := 80
	
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	if err != nil {
		return defaultWidth
	}
	
	var rows, cols int
	_, err = fmt.Sscanf(string(out), "%d %d", &rows, &cols)
	if err != nil || cols <= 0 {
		return defaultWidth
	}
	
	return cols
}

// RuneDisplayLength calculates the display width of a string accounting for ANSI codes
func RuneDisplayLength(s string) int {
    // Strip any potential ANSI color codes
    ansiRegex := regexp.MustCompile("\x1b\\[[0-9;]*m")
    cleanStr := ansiRegex.ReplaceAllString(s, "")
    
    return len(cleanStr)
}

// FormatValueWithContext adds visual context (colors) to values based on their meaning
func FormatValueWithContext(key, value string) string {
    // Color-code usage metrics
    if strings.Contains(key, "Usage") && strings.Contains(value, "%") {
        // Extract percentage for coloring
        percentStr := strings.TrimSpace(strings.Split(value, " ")[0])
        percent := 0.0
        fmt.Sscanf(percentStr, "%f", &percent)
        
        if percent < 60.0 {
            return SuccessColor(value)
        } else if percent < 85.0 {
            return WarningColor(value)
        } else {
            return DangerColor(value)
        }
    }
    
    // Color-code status values
    if key == "Status" {
        if value == "Up" || value == "Running" || value == "Active" || strings.Contains(value, "OK") {
            return SuccessColor(value)
        } else if value == "Down" || value == "Stopped" || strings.Contains(value, "Error") {
            return DangerColor(value)
        }
    }
    
    return ValueColor(value)
}

// GetSectionIcon returns an appropriate icon for a section
func GetSectionIcon(sectionName string) string {
    switch sectionName {
    case "System", "Runtime Environment", "Runtime":
        return Info + " "
    case "CPU":
        return Cpu + " "
    case "Memory":
        return Memory + " "
    case "Disk":
        return Disk + " "
    case "Network":
        return Network + " "
    case "Top Processes", "Processes":
        return "⏺ "
    case "Docker", "Containers":
        return "🐳" // Docker whale icon
    default:
        return BulletPoint + " "
    }
}
