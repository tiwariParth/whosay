package ui

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"sort"
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
	fmt.Println()
}

func PrintTitle(title string) {
	width := len(title) + 6
	fmt.Printf("\n %s%s%s\n", BoxTopLeft, strings.Repeat(BoxHorizontal, width), BoxTopRight)
	fmt.Printf(" %s  %s  %s\n", BoxVertical, TitleColor(title), BoxVertical)
	fmt.Printf(" %s%s%s\n", BoxBottomLeft, strings.Repeat(BoxHorizontal, width), BoxBottomRight)
}

func PrintKeyValue(key string, value interface{}) {
	fmt.Printf("  %s %s %s\n", 
		AccentColor(BulletPoint),
		LabelColor(fmt.Sprintf("%s:", key)), 
		ValueColor(fmt.Sprintf("%v", value)))
}

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

func PrintUsageBar(label string, percentage float64) {
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
	
	bar.Set(int(percentage))
	fmt.Printf(" %s\n", colorFunc(fmt.Sprintf("%.1f%%", percentage)))
}

func PrintCompactUsageBar(label string, percentage float64, width int) string {
    var colorFunc func(a ...interface{}) string
    
    switch {
    case percentage < 60:
        colorFunc = SuccessColor
    case percentage < 85:
        colorFunc = WarningColor
    default:
        colorFunc = DangerColor
    }
    
    barWidth := width - 9
    if barWidth < 5 {
        barWidth = 5
    }
    
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
    
    bar := colorFunc(strings.Repeat("■", filledWidth)) + DimColor(strings.Repeat("□", emptyWidth))
    
    return fmt.Sprintf("%s %s", 
        bar,
        colorFunc(fmt.Sprintf("%5.1f%%", percentage)))
}

func ClearScreen() {
	var cmd *exec.Cmd
	
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "cls")
	default:
		cmd = exec.Command("clear")
	}
	
	cmd.Stdout = os.Stdout
	_ = cmd.Run()
}

func GetTerminalWidth() int {
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

func RuneDisplayLength(s string) int {
    ansiRegex := regexp.MustCompile("\x1b\\[[0-9;]*m")
    cleanStr := ansiRegex.ReplaceAllString(s, "")
    
    return len(cleanStr)
}

func FormatValueWithContext(key, value string) string {
    if strings.Contains(key, "Usage") && strings.Contains(value, "%") {
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
    
    if key == "Status" {
        if value == "Up" || value == "Running" || value == "Active" || strings.Contains(value, "OK") {
            return SuccessColor(value)
        } else if value == "Down" || value == "Stopped" || strings.Contains(value, "Error") {
            return DangerColor(value)
        }
    }
    
    return ValueColor(value)
}

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
        return "🐳"
    default:
        return BulletPoint + " "
    }
}

func isTopLevelSection(sectionName string) bool {
	topLevelSections := map[string]bool{
		"System": true,
		"Runtime Environment": true,
		"Overview": true,
	}
	
	return topLevelSections[sectionName]
}

func getSortedSectionNames(sections map[string][][]string) []string {
	sectionOrder := map[string]int{
		"System":            1,
		"Runtime Environment": 2,
		"CPU":               3,
		"Memory":            4,
		"Disk":              5,
		"Network":           6,
		"Network Traffic":   7,
		"Top Processes":     8,
		"Processes":         9,
		"Docker":            10,
		"Containers":        11,
		"Battery":           12,
		"Temperature":       13,
		"System Logs":       14,
		"Resource History":  15,
	}
	
	names := make([]string, 0, len(sections))
	for name := range sections {
		names = append(names, name)
	}
	
	sort.Slice(names, func(i, j int) bool {
		orderI, existsI := sectionOrder[names[i]]
		orderJ, existsJ := sectionOrder[names[j]]
		
		if !existsI && !existsJ {
			return names[i] < names[j]
		} else if !existsI {
			return false
		} else if !existsJ {
			return true
		}
		
		return orderI < orderJ
	})
	
	return names
}

func CompactDisplay(sections map[string][][]string) {
	termWidth := GetTerminalWidth()
	
	for _, sectionName := range getSortedSectionNames(sections) {
		data := sections[sectionName]
		if len(data) == 0 {
			continue
		}
		
		if !isTopLevelSection(sectionName) { 
			icon := GetSectionIcon(sectionName)
			fmt.Printf(" %s%s %s %s\n", 
				BoxTopLeft, 
				BoxHorizontal,
				SectionColor(icon + sectionName),
				BoxHorizontal)
			
			fmt.Printf(" %s%s%s%s%s\n",
				BoxLeftT,
				strings.Repeat(ThinBoxHorizontal, termWidth-4),
				BoxRightT,
				"",
				"")
		} else {
			fmt.Println()
		}
		
		printSectionData(data, termWidth)
	}
}

// Add the printSectionData function that was referenced
func printSectionData(data [][]string, termWidth int) {
	if len(data) == 0 {
		return
	}
	
	maxLabelWidth := 0
	for _, row := range data {
		if len(row) >= 2 && len(row[0]) > maxLabelWidth {
			maxLabelWidth = len(row[0])
		}
	}
	
	for _, row := range data {
		if len(row) == 0 {
			continue
		}
		
		if len(row) == 1 {
			fmt.Printf("  %s\n", ValueColor(row[0]))
			continue
		}
		
		label := row[0]
		value := row[1]
		
		if label == "" && value == "" {
			fmt.Println()
			continue
		}
		
		if label == "" {
			fmt.Printf("  %s\n", ValueColor(value))
		} else {
			PrintCompactKeyValue(label, value, maxLabelWidth)
		}
	}
	
	fmt.Println()
}
