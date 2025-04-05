package collectors

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/tiwariParth/whosay/internal/models"
	"github.com/tiwariParth/whosay/internal/ui"
)

// GetDockerInfo displays information about running Docker containers
func GetDockerInfo(opts models.Options) {
	containers, err := GetDockerContainers()
	if err != nil {
		// Check if it's just docker not being available
		if isDockerNotInstalled(err) {
			if opts.JSONOutput {
				fmt.Println("[]")
			} else {
				fmt.Println("Docker doesn't appear to be installed or isn't running.")
				fmt.Println("Install Docker or start the Docker service to see container information.")
			}
			return
		}
		
		// Otherwise it's some other error
		fmt.Printf("Error getting Docker containers: %v\n", err)
		return
	}

	if opts.JSONOutput {
		jsonData, err := json.MarshalIndent(containers, "", "  ")
		if err != nil {
			fmt.Printf("Error serializing container data: %v\n", err)
			return
		}
		fmt.Println(string(jsonData))
		return
	}

	// Format and display containers in compact view
	sections := GetDockerInfoSections(containers, opts)
	ui.CompactDisplay(sections)
}

// GetDockerInfoSections formats container information for the compact display
func GetDockerInfoSections(containers []models.ContainerInfo, opts models.Options) map[string][][]string {
	// Create the main docker section
	dockerData := [][]string{
		{"Containers", fmt.Sprintf("%d running", len(containers))},
	}

	if len(containers) == 0 {
		dockerData = append(dockerData, []string{"Status", "No containers running"})
	}

	// Create container sections
	containerSections := [][]string{}

	// Create a container table
	for _, container := range containers {
		// Format name (remove leading slash if present)
		name := container.Name
		if strings.HasPrefix(name, "/") {
			name = name[1:]
		}

		// Calculate memory in MB
		memoryMB := float64(container.MemoryUsage) / 1024 / 1024
		memoryLimitMB := float64(container.MemoryLimit) / 1024 / 1024
		
		// Truncate image name if needed
		image := container.Image
		if len(image) > 25 {
			// Try to keep the repository and tag, remove middle part
			parts := strings.Split(image, ":")
			if len(parts) > 1 {
				if len(parts[0]) > 20 {
					parts[0] = parts[0][:17] + "..."
				}
				image = parts[0] + ":" + parts[1]
			} else {
				image = image[:22] + "..."
			}
		}

		containerSections = append(containerSections, []string{
			name,
			fmt.Sprintf("%-25s %s", image, container.Status),
		})
		
		// Add detailed info per container
		containerSections = append(containerSections, []string{
			"CPU",
			fmt.Sprintf("%.1f%%", container.CPUPercent),
		})
		
		// Add memory info
		if memoryLimitMB > 0 {
			containerSections = append(containerSections, []string{
				"Memory",
				fmt.Sprintf("%.1f MB / %.1f MB (%.1f%%)", memoryMB, memoryLimitMB, container.MemoryPerc),
			})
		} else {
			containerSections = append(containerSections, []string{
				"Memory",
				fmt.Sprintf("%.1f MB", memoryMB),
			})
		}
		
		// Add IP and ports if available
		if container.IPAddress != "" {
			containerSections = append(containerSections, []string{
				"IP",
				container.IPAddress,
			})
		}
		
		if len(container.Ports) > 0 {
			// Join ports with commas if there are more than two
			portsStr := container.Ports[0]
			if len(container.Ports) > 1 {
				if len(container.Ports) > 3 {
					portsStr = strings.Join(container.Ports[:2], ", ") + fmt.Sprintf(" (+%d more)", len(container.Ports)-2)
				} else {
					portsStr = strings.Join(container.Ports, ", ")
				}
			}
			
			containerSections = append(containerSections, []string{
				"Ports",
				portsStr,
			})
		}
		
		// Add spacer between containers
		containerSections = append(containerSections, []string{"", ""})
	}

	// Build the final sections map
	result := map[string][][]string{
		"Docker": dockerData,
	}

	// Add containers section if we have containers
	if len(containers) > 0 {
		result["Containers"] = containerSections
	}

	return result
}

// GetDockerContainers returns information about running Docker containers
func GetDockerContainers() ([]models.ContainerInfo, error) {
	result := []models.ContainerInfo{}

	// Check if docker is installed
	if _, err := exec.LookPath("docker"); err != nil {
		return result, fmt.Errorf("docker command not found: %w", err)
	}

	// Get list of containers
	cmd := exec.Command("docker", "ps", "--format", "{{.ID}}|{{.Names}}|{{.Image}}|{{.Status}}|{{.Command}}|{{.CreatedAt}}")
	output, err := cmd.Output()
	if err != nil {
		return result, fmt.Errorf("failed to run docker ps: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) < 5 {
			continue
		}

		// Parse container info
		container := models.ContainerInfo{
			ID:      parts[0],
			Name:    parts[1],
			Image:   parts[2],
			Status:  parts[3],
			Command: parts[4],
		}

		// Set state based on status
		if strings.HasPrefix(container.Status, "Up") {
			container.State = "running"
		} else if strings.HasPrefix(container.Status, "Exited") {
			container.State = "exited"
		} else if strings.HasPrefix(container.Status, "Created") {
			container.State = "created"
		} else {
			container.State = "unknown"
		}

		// If the container is running, get more detailed info
		if container.State == "running" {
			// Get container stats
			statCmd := exec.Command("docker", "stats", "--no-stream", "--format", 
			                        "{{.CPUPerc}}|{{.MemUsage}}|{{.MemPerc}}", container.ID)
			statOutput, err := statCmd.Output()
			if err == nil {
				statParts := strings.Split(strings.TrimSpace(string(statOutput)), "|")
				if len(statParts) >= 3 {
					// Parse CPU percentage
					cpuStr := strings.TrimSuffix(statParts[0], "%")
					cpu, err := strconv.ParseFloat(cpuStr, 64)
					if err == nil {
						container.CPUPercent = cpu
					}

					// Parse memory usage
					memParts := strings.Fields(statParts[1])
					if len(memParts) >= 3 {
						mem, unit := memParts[0], memParts[1]
						memVal, err := strconv.ParseFloat(mem, 64)
						if err == nil {
							// Convert to bytes based on unit
							switch unit {
							case "KiB", "KB":
								container.MemoryUsage = uint64(memVal * 1024)
							case "MiB", "MB":
								container.MemoryUsage = uint64(memVal * 1024 * 1024)
							case "GiB", "GB":
								container.MemoryUsage = uint64(memVal * 1024 * 1024 * 1024)
							}
						}

						// Parse memory limit if available
						if len(memParts) >= 5 {
							limit, unit := memParts[2], memParts[3]
							limitVal, err := strconv.ParseFloat(limit, 64)
							if err == nil {
								// Convert to bytes based on unit
								switch unit {
								case "KiB", "KB":
									container.MemoryLimit = uint64(limitVal * 1024)
								case "MiB", "MB":
									container.MemoryLimit = uint64(limitVal * 1024 * 1024)
								case "GiB", "GB":
									container.MemoryLimit = uint64(limitVal * 1024 * 1024 * 1024)
								}
							}
						}
					}

					// Parse memory percentage
					memPercStr := strings.TrimSuffix(statParts[2], "%")
					memPerc, err := strconv.ParseFloat(memPercStr, 64)
					if err == nil {
						container.MemoryPerc = memPerc
					}
				}
			}

			// Get network info
			inspectCmd := exec.Command("docker", "inspect", "--format", "{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}", container.ID)
			inspectOutput, err := inspectCmd.Output()
			if err == nil {
				container.IPAddress = strings.TrimSpace(string(inspectOutput))
			}

			// Get port mappings
			portCmd := exec.Command("docker", "port", container.ID)
			portOutput, err := portCmd.Output()
			if err == nil {
				portLines := strings.Split(string(portOutput), "\n")
				for _, portLine := range portLines {
					if portLine != "" {
						container.Ports = append(container.Ports, strings.TrimSpace(portLine))
					}
				}
			}
		}

		result = append(result, container)
	}

	return result, nil
}

// isDockerNotInstalled checks if the error is due to Docker not being installed
func isDockerNotInstalled(err error) bool {
	if err == nil {
		return false
	}
	
	// Check common error messages that indicate Docker isn't available
	errorMsg := err.Error()
	notInstalledPatterns := []string{
		"command not found",
		"docker daemon is not running",
		"connection refused",
		"Is the docker daemon running",
		"docker.sock: no such file",
		"Cannot connect to the Docker daemon",
	}
	
	for _, pattern := range notInstalledPatterns {
		if strings.Contains(errorMsg, pattern) {
			return true
		}
	}
	
	return false
}
