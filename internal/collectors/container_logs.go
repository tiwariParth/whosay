package collectors

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/tiwariParth/whosay/internal/models"
	"github.com/tiwariParth/whosay/internal/ui"
)

// GetContainerLogs displays logs from a specified Docker container
func GetContainerLogs(containerID string, tailLines int, opts models.Options) {
	logs, err := fetchContainerLogs(containerID, tailLines)
	if err != nil {
		fmt.Printf("Error fetching container logs: %v\n", err)
		return
	}

	if opts.JSONOutput {
		// Format logs as JSON for programmatic consumption
		type logEntry struct {
			Container string    `json:"container"`
			Timestamp time.Time `json:"timestamp,omitempty"`
			Message   string    `json:"message"`
		}

		entries := make([]logEntry, len(logs))
		for i, log := range logs {
			entries[i] = logEntry{
				Container: containerID,
				Message:   log,
			}
		}

		jsonData, err := json.MarshalIndent(entries, "", "  ")
		if err != nil {
			fmt.Printf("Error serializing log data: %v\n", err)
			return
		}
		fmt.Println(string(jsonData))
		return
	}

	// Display logs in a pretty format
	sections := GetContainerLogSections(containerID, logs, opts)
	ui.CompactDisplay(sections)
}

// GetContainerLogSections formats container logs for display
func GetContainerLogSections(containerID string, logs []string, opts models.Options) map[string][][]string {
	// Get container name if we have a container ID
	containerName := containerID
	if len(containerID) >= 12 {
		// Try to get the actual container name
		nameCmd := exec.Command("docker", "inspect", "--format", "{{.Name}}", containerID)
		nameOutput, err := nameCmd.Output()
		if err == nil {
			name := strings.TrimSpace(string(nameOutput))
			// Remove leading slash from container name
			if strings.HasPrefix(name, "/") {
				name = name[1:]
			}
			if name != "" {
				containerName = name
			}
		}
	}

	// Create header for the logs section
	logData := [][]string{
		{"Container", containerName},
		{"ID", containerID},
		{"Lines", fmt.Sprintf("%d", len(logs))},
		{"", ""}, // Spacer
	}

	// Add log lines
	for _, line := range logs {
		// Color code log lines based on content
		coloredLine := formatLogLine(line)
		logData = append(logData, []string{"", coloredLine})
	}

	// If no logs found
	if len(logs) == 0 {
		logData = append(logData, []string{"", "No logs found for this container"})
	}

	// Build the final sections map
	return map[string][][]string{
		"Container Logs": logData,
	}
}

// fetchContainerLogs gets logs from a Docker container
func fetchContainerLogs(containerID string, tailLines int) ([]string, error) {
	// Validate container ID/name exists
	checkCmd := exec.Command("docker", "container", "inspect", containerID)
	if err := checkCmd.Run(); err != nil {
		return nil, fmt.Errorf("container '%s' not found: %w", containerID, err)
	}

	// Construct command to fetch logs
	tailArg := "all"
	if tailLines > 0 {
		tailArg = fmt.Sprintf("%d", tailLines)
	}

	cmd := exec.Command("docker", "logs", "--tail", tailArg, containerID)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch logs from container '%s': %w", containerID, err)
	}

	// Split logs into lines
	logs := strings.Split(string(output), "\n")
	
	// Remove empty trailing line if present
	if len(logs) > 0 && logs[len(logs)-1] == "" {
		logs = logs[:len(logs)-1]
	}

	return logs, nil
}

// ListContainers lists all running containers
func ListContainers() ([]models.ContainerInfo, error) {
	return GetDockerContainers()
}

// formatLogLine applies colors to log line based on content
func formatLogLine(line string) string {
	line = strings.TrimSpace(line)
	
	lowerLine := strings.ToLower(line)
	switch {
	case strings.Contains(lowerLine, "error") || 
	     strings.Contains(lowerLine, "fail") || 
	     strings.Contains(lowerLine, "exception"):
		return ui.DangerColor(line)
	case strings.Contains(lowerLine, "warn") || 
	     strings.Contains(lowerLine, "warning"):
		return ui.WarningColor(line)
	case strings.Contains(lowerLine, "info") || 
	     strings.Contains(lowerLine, "notice"):
		return ui.InfoColor(line)
	case strings.Contains(lowerLine, "debug"):
		return ui.DimColor(line)
	default:
		return line
	}
}
