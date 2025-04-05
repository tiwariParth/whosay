package ui

import (
	"fmt"
	"strings"
)

// RenderProcessTable displays process information in a simple table layout
func RenderProcessTable(processes [][]string, width int, vertChar string) {
    // Calculate column widths based on the first row (headers)
    colWidths := make([]int, len(processes[0]))
    for _, row := range processes {
        for i, col := range row {
            if len(col) > colWidths[i] {
                colWidths[i] = len(col)
            }
        }
    }

    // Ensure the table fits within the available width
    totalWidth := 4 // Account for borders and padding
    for _, colWidth := range colWidths {
        totalWidth += colWidth + 3 // Add padding between columns
    }
    if totalWidth > width {
        // Adjust column widths proportionally to fit within the width
        excess := totalWidth - width
        for i := range colWidths {
            reduction := (colWidths[i] * excess) / totalWidth
            colWidths[i] -= reduction
            if colWidths[i] < 5 { // Minimum column width
                colWidths[i] = 5
            }
        }
    }

    // Render the table
    for idx, row := range processes {
        fmt.Print(" ", vertChar, " ")
        
        // Color the header row differently
        if idx == 0 {
            for i, col := range row {
                fmt.Print(LabelColor(fmt.Sprintf("%-*s", colWidths[i], col)))
                if i < len(row)-1 {
                    fmt.Print(" │ ")
                }
            }
        } else {
            for i, col := range row {
                // Color code percentage values
                if i >= 2 && strings.Contains(col, ".") { // CPU% and Memory% columns
                    val := 0.0
                    fmt.Sscanf(col, "%f", &val)
                    
                    if val < 10.0 {
                        fmt.Print(SuccessColor(fmt.Sprintf("%-*s", colWidths[i], col)))
                    } else if val < 50.0 {
                        fmt.Print(WarningColor(fmt.Sprintf("%-*s", colWidths[i], col)))
                    } else {
                        fmt.Print(DangerColor(fmt.Sprintf("%-*s", colWidths[i], col)))
                    }
                } else {
                    fmt.Printf("%-*s", colWidths[i], col)
                }
                
                if i < len(row)-1 {
                    fmt.Print(" │ ")
                }
            }
        }
        
        fmt.Println(" ", vertChar)

        // Print a separator after the header row
        if idx == 0 {
            fmt.Print(" ", vertChar, " ")
            for i, colWidth := range colWidths {
                fmt.Print(strings.Repeat("─", colWidth))
                if i < len(colWidths)-1 {
                    fmt.Print("─┼─")
                }
            }
            fmt.Println(" ", vertChar)
        }
    }
}

// CompactDisplay renders information sections in a modern, developer-friendly layout
func CompactDisplay(sections map[string][][]string) {
    // Sort section names for consistent display order
    sectionNames := GetSortedSectionNames(sections)
    termWidth := GetTerminalWidth()
    if termWidth > 120 {
        termWidth = 120 // Cap maximum width for better readability
    }

    // Calculate max key length for alignment
    maxKeyLen := 0
    for _, sectionName := range sectionNames {
        for _, pair := range sections[sectionName] {
            if len(pair[0]) > maxKeyLen {
                maxKeyLen = len(pair[0])
            }
        }
    }

    // Print app header with modern style
    fmt.Println()
    headerText := " whosay - Developer System Monitor "
    headerWidth := termWidth - 4
    
    // Create a full-width header for better visibility
    fmt.Printf(" %s\n", HeaderBgColor(strings.Repeat(" ", headerWidth)))
    
    // Center the header text
    headerPadding := (headerWidth - len(headerText)) / 2
    if headerPadding < 0 {
        headerPadding = 0
    }
    
    paddedHeader := strings.Repeat(" ", headerPadding) + headerText + 
                   strings.Repeat(" ", headerWidth - headerPadding - len(headerText))
    fmt.Printf(" %s\n", HeaderBgColor(paddedHeader))
    fmt.Printf(" %s\n\n", HeaderBgColor(strings.Repeat(" ", headerWidth)))

    // Display each section with modern styling
    for i, sectionName := range sectionNames {
        // Calculate section width
        sectionWidth := termWidth - 4
        
        // Choose an icon based on section name
        sectionIcon := GetSectionIcon(sectionName)
        
        // Print section header with box drawing
        fmt.Printf(" %s%s%s\n", 
            BoxTopLeft, 
            strings.Repeat(BoxHorizontal, sectionWidth-2), 
            BoxTopRight)
        
        // Add icon to section title for visual cues
        fmt.Printf(" %s %s %s %s\n", 
            BoxVertical,
            sectionIcon,
            SectionColor(sectionName),
            BoxVertical)
        
        fmt.Printf(" %s%s%s\n", 
            BoxLeftT, 
            strings.Repeat(ThinBoxHorizontal, sectionWidth-2), 
            BoxRightT)
        
        // Special handling for Top Processes section which needs tabular formatting
        if sectionName == "Top Processes" {
            RenderProcessTable(sections[sectionName], sectionWidth, BoxVertical)
        } else {
            // Standard rendering for other sections
            for _, pair := range sections[sectionName] {
                key := pair[0]
                value := pair[1]
                
                // Skip empty rows (used as spacers)
                if key == "" && value == "" {
                    fmt.Printf(" %s%s%s\n", 
                        BoxVertical, 
                        strings.Repeat(" ", sectionWidth-2), 
                        BoxVertical)
                    continue
                }
                
                // Add padding for alignment
                padding := maxKeyLen - len(key) + 2
                if padding < 1 {
                    padding = 1
                }
                
                // Color-code values based on context (for usage metrics)
                formattedValue := FormatValueWithContext(key, value)
                
                // Print key-value pair with consistent alignment
                fmt.Printf(" %s  %s%s%s  %s\n", 
                    BoxVertical,
                    LabelColor(fmt.Sprintf("%s:", key)),
                    strings.Repeat(" ", padding),
                    formattedValue,
                    BoxVertical)
            }
        }
        
        // Close section box
        fmt.Printf(" %s%s%s\n", 
            BoxBottomLeft, 
            strings.Repeat(BoxHorizontal, sectionWidth-2), 
            BoxBottomRight)
        
        // Add spacing between sections if not the last one
        if i < len(sectionNames)-1 {
            fmt.Println() // Add a blank line between sections
        }
    }
}

// GetSortedSectionNames returns section names in a consistent order for display
func GetSortedSectionNames(sections map[string][][]string) []string {
    // Define priority order for sections
    priorityOrder := []string{
        "System",
        "CPU",
        "Memory",
        "Disk",
        "Network",
        "Runtime Environment",
        "Runtime",
        "Top Processes",
        "Processes",
        "Docker",
        "Containers",
    }
    
    // Build result with priority sections first
    result := make([]string, 0, len(sections))
    
    // Add priority sections in order
    for _, name := range priorityOrder {
        if _, exists := sections[name]; exists {
            result = append(result, name)
        }
    }
    
    // Add any remaining sections (network interfaces, etc.)
    for name := range sections {
        // Skip if already added from priority list
        alreadyAdded := false
        for _, added := range result {
            if added == name {
                alreadyAdded = true
                break
            }
        }
        
        if !alreadyAdded {
            result = append(result, name)
        }
    }
    
    return result
}
