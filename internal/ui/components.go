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
