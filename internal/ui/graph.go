package ui

import (
	"fmt"
	"strings"
)

// RenderBarGraph creates a horizontal bar graph with labels
func RenderBarGraph(data []float64, labels []string, width int, title string) string {
	if len(data) == 0 || len(data) != len(labels) {
		return "No data or mismatched data/labels"
	}
	
	// Find max value for scaling
	maxVal := data[0]
	for _, val := range data {
		if val > maxVal {
			maxVal = val
		}
	}
	
	// Find max label length for alignment
	maxLabelLen := 0
	for _, label := range labels {
		if len(label) > maxLabelLen {
			maxLabelLen = len(label)
		}
	}
	
	// Define available width for bars
	barWidth := width - maxLabelLen - 15 // Allow space for label, value and padding
	if barWidth < 10 {
		barWidth = 10 // Minimum bar width
	}
	
	// Build the graph
	var sb strings.Builder
	
	// Add title if provided
	if title != "" {
		sb.WriteString(SectionColor(title))
		sb.WriteString("\n")
	}
	
	// Draw bars
	for i, val := range data {
		// Add label with padding
		sb.WriteString(fmt.Sprintf("  %s%s ", 
			LabelColor(labels[i]), 
			strings.Repeat(" ", maxLabelLen-len(labels[i]))))
		
		// Calculate bar length
		barLen := 1
		if maxVal > 0 {
			barLen = int((val / maxVal) * float64(barWidth))
		}
		if barLen < 1 {
			barLen = 1
		}
		
		// Color code based on percentage of max
		var barColor func(...interface{}) string
		percentage := (val / maxVal) * 100
		switch {
		case percentage < 30:
			barColor = SuccessColor
		case percentage < 70:
			barColor = WarningColor
		default:
			barColor = DangerColor
		}
		
		// Draw the bar
		sb.WriteString(barColor(strings.Repeat("█", barLen)))
		
		// Add value at the end
		sb.WriteString(fmt.Sprintf(" %s\n", ValueColor(fmt.Sprintf("%.2f", val))))
	}
	
	return sb.String()
}

// RenderLineGraph creates a simple ASCII line graph for time series data
func RenderLineGraph(data []float64, width, height int, title string) string {
	if len(data) == 0 {
		return "No data"
	}
	
	// Find max value for scaling
	maxVal := data[0]
	for _, val := range data {
		if val > maxVal {
			maxVal = val
		}
	}
	
	// Add a small buffer to the max
	maxVal *= 1.1
	
	// Ensure we have a non-zero max
	if maxVal < 0.001 {
		maxVal = 0.001
	}
	
	// Define the graph characters
	const (
		empty      = " "
		dataPoint  = "•"
		axisVert   = "│"
		axisHoriz  = "─"
		axisCorner = "└"
		axisTop    = "┬"
		axisLeft   = "├"
	)
	
	// Scale data to fit height
	scaledData := make([]int, len(data))
	for i, val := range data {
		scaled := int((val / maxVal) * float64(height))
		if scaled > height {
			scaled = height
		}
		scaledData[i] = scaled
	}
	
	// Determine how many data points to display
	dataPoints := len(data)
	if dataPoints > width-2 {
		data = data[dataPoints-(width-2):]
		scaledData = scaledData[dataPoints-(width-2):]
		dataPoints = width - 2
	}
	
	// Create grid
	grid := make([][]string, height+1)
	for i := range grid {
		grid[i] = make([]string, dataPoints+2) // +2 for axis
		for j := range grid[i] {
			grid[i][j] = empty
		}
	}
	
	// Add axis
	for i := 0; i < height+1; i++ {
		grid[i][0] = axisVert
	}
	for j := 0; j < dataPoints+2; j++ {
		grid[height][j] = axisHoriz
	}
	grid[height][0] = axisCorner
	
	// Add data points
	for i, scaled := range scaledData {
		if scaled > 0 {
			y := height - scaled
			x := i + 1
			grid[y][x] = dataPoint
		}
	}
	
	// Build the graph string
	var sb strings.Builder
	
	// Add title if provided
	if title != "" {
		sb.WriteString(SectionColor(title))
		sb.WriteString("\n")
	}
	
	// Add max value at top
	sb.WriteString(fmt.Sprintf("%s %.2f\n", axisTop, maxVal))
	
	// Add middle value in the middle
	middleValue := maxVal / 2
	middleRow := height / 2
	sb.WriteString(fmt.Sprintf("%s\n", strings.Join(grid[0][:], "")))
	for i := 1; i < height; i++ {
		if i == middleRow {
			sb.WriteString(fmt.Sprintf("%s %.2f %s\n", 
				axisLeft, middleValue, strings.Join(grid[i][1:], "")))
		} else {
			sb.WriteString(fmt.Sprintf("%s\n", strings.Join(grid[i][:], "")))
		}
	}
	
	// Add bottom of graph with zero
	sb.WriteString(fmt.Sprintf("%s %.2f", axisCorner, 0.0))
	sb.WriteString("\n")
	
	return sb.String()
}

// RenderSparkline creates a simple inline sparkline graph
func RenderSparkline(data []float64, width int) string {
	if len(data) == 0 {
		return ""
	}
	
	// Use a simpler set of block characters for the sparkline
	sparkChars := []rune{' ', '▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}
	
	// Find max value for scaling
	maxVal := data[0]
	for _, val := range data {
		if val > maxVal {
			maxVal = val
		}
	}
	
	// Ensure non-zero max
	if maxVal < 0.001 {
		maxVal = 0.001
	}
	
	// Determine how many data points to display
	dataPoints := len(data)
	if dataPoints > width {
		data = data[dataPoints-width:]
		dataPoints = width
	}
	
	// Build the sparkline
	var sb strings.Builder
	for _, val := range data {
		// Scale value to the sparkline character set
		idx := 0
		if maxVal > 0 {
			idx = int((val / maxVal) * float64(len(sparkChars)-1))
		}
		if idx < 0 {
			idx = 0
		} else if idx >= len(sparkChars) {
			idx = len(sparkChars) - 1
		}
		
		// Add the appropriate character
		sb.WriteRune(sparkChars[idx])
	}
	
	return sb.String()
}
