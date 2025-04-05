package collectors

import (
	"encoding/json"
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"runtime"
	"strings"

	"github.com/tiwariParth/whosay/internal/models"
	"github.com/tiwariParth/whosay/internal/ui"
)

// GetNetworkInfo collects and displays network information
func GetNetworkInfo(opts models.Options) {
	info, err := collectNetworkInfo()
	if err != nil {
		// Handle error gracefully, don't crash
		fmt.Printf("Error collecting network information: %v\n", err)
		return
	}

	if opts.JSONOutput {
		jsonData, err := json.MarshalIndent(info, "", "  ")
		if err != nil {
			fmt.Printf("Error serializing network data: %v\n", err)
			return
		}
		fmt.Println(string(jsonData))
		return
	}

	// Format and display network info in compact view
	sections := GetNetworkInfoSections(opts)
	ui.CompactDisplay(sections)
}

// GetNetworkInfoSections formats network information for the compact display
func GetNetworkInfoSections(opts models.Options) map[string][][]string {
	info, err := collectNetworkInfo()
	if err != nil {
		// Return minimal error section if we can't get network info
		return map[string][][]string{
			"Network": {
				{"Status", fmt.Sprintf("Error: %v", err)},
			},
		}
	}

	// Create main network section
	networkSection := [][]string{
		{"Gateway", info.DefaultGateway},
		{"DNS", strings.Join(info.DNSServers, ", ")},
	}

	// Create interface sections
	interfaceSections := make(map[string][][]string)

	for _, iface := range info.Interfaces {
		// Skip interfaces with no IP addresses if not in verbose mode
		if len(iface.IPv4) == 0 && len(iface.IPv6) == 0 && !opts.VerboseOutput {
			continue
		}

		// Skip loopback interface unless in verbose mode
		if (iface.Name == "lo" || strings.HasPrefix(iface.Name, "loop")) && !opts.VerboseOutput {
			continue
		}

		ifaceType := "Interface"
		if iface.IsVPN {
			ifaceType = "VPN"
		} else if iface.IsWifi {
			ifaceType = "Wifi"
		}

		ifaceData := [][]string{
			{"Status", iface.Status},
			{"MAC", iface.MAC},
		}

		// Add IPv4 addresses
		if len(iface.IPv4) > 0 {
			ifaceData = append(ifaceData, []string{"IPv4", strings.Join(iface.IPv4, ", ")})
		}

		// Add IPv6 addresses if we have them
		if len(iface.IPv6) > 0 {
			// Only add first IPv6 to save space
			ifaceData = append(ifaceData, []string{"IPv6", truncateIPv6(iface.IPv6[0])})
		}

		// Add speed if available
		if iface.Speed != "" {
			ifaceData = append(ifaceData, []string{"Speed", iface.Speed})
		}

		interfaceSections[fmt.Sprintf("%s: %s", ifaceType, iface.Name)] = ifaceData
	}

	// Combine network section with interfaces
	result := map[string][][]string{
		"Network": networkSection,
	}

	// Add interfaces to result
	for name, section := range interfaceSections {
		result[name] = section
	}

	return result
}

// collectNetworkInfo gathers information about network interfaces and config
func collectNetworkInfo() (models.NetworkInfo, error) {
	var info models.NetworkInfo

	// Get network interfaces
	ifaces, err := net.Interfaces()
	if err != nil {
		return info, fmt.Errorf("failed to get network interfaces: %w", err)
	}

	info.Interfaces = make([]models.NetworkInterface, 0, len(ifaces))
	for _, iface := range ifaces {
		// Skip interfaces that are down if needed
		if iface.Flags&net.FlagUp == 0 {
			// Only include if it's important enough
			continue
		}

		netIface := models.NetworkInterface{
			Name:   iface.Name,
			MAC:    iface.HardwareAddr.String(),
			Status: getInterfaceStatus(iface),
			Speed:  getInterfaceSpeed(iface.Name),
			IPv4:   make([]string, 0),
			IPv6:   make([]string, 0),
			IsVPN:  isVPNInterface(iface.Name),
			IsWifi: isWifiInterface(iface.Name),
		}

		// Get IP addresses
		addrs, err := iface.Addrs()
		if err == nil {
			for _, addr := range addrs {
				// Parse CIDR notation
				ipNet, ok := addr.(*net.IPNet)
				if !ok {
					continue
				}

				// Skip local link addresses
				if ipNet.IP.IsLinkLocalUnicast() || ipNet.IP.IsLinkLocalMulticast() {
					continue
				}

				// Separate IPv4 and IPv6 addresses
				if ipNet.IP.To4() != nil {
					netIface.IPv4 = append(netIface.IPv4, ipNet.IP.String())
				} else {
					netIface.IPv6 = append(netIface.IPv6, ipNet.IP.String())
				}
			}
		}

		info.Interfaces = append(info.Interfaces, netIface)
	}

	// Get default gateway
	info.DefaultGateway = getDefaultGateway()

	// Get DNS servers
	info.DNSServers = getDNSServers()

	return info, nil
}

// getInterfaceStatus returns a human-readable interface status
func getInterfaceStatus(iface net.Interface) string {
	if iface.Flags&net.FlagUp != 0 {
		return "Up"
	}
	return "Down"
}

// getDefaultGateway tries to determine the default gateway
func getDefaultGateway() string {
	// Try multiple methods based on OS
	switch runtime.GOOS {
	case "linux":
		cmd := exec.Command("ip", "route", "show", "default")
		output, err := cmd.Output()
		if err == nil {
			// Parse output to extract gateway
			fields := strings.Fields(string(output))
			for i, field := range fields {
				if field == "via" && i+1 < len(fields) {
					return fields[i+1]
				}
			}
		}
	case "darwin":
		cmd := exec.Command("netstat", "-nr")
		output, err := cmd.Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "default") || strings.HasPrefix(line, "0.0.0.0") {
					fields := strings.Fields(line)
					if len(fields) >= 2 {
						return fields[1]
					}
				}
			}
		}
	case "windows":
		cmd := exec.Command("cmd", "/c", "ipconfig")
		output, err := cmd.Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				if strings.Contains(line, "Default Gateway") {
					parts := strings.Split(line, ":")
					if len(parts) >= 2 {
						return strings.TrimSpace(parts[1])
					}
				}
			}
		}
	}

	return "Unknown"
}

// getDNSServers attempts to get the DNS server list
func getDNSServers() []string {
	dnsServers := []string{}

	// Try to read resolv.conf on Unix-like systems
	if runtime.GOOS != "windows" {
		// Use exec.Command instead of ioutil to read the file
		cmd := exec.Command("cat", "/etc/resolv.conf")
		output, err := cmd.Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			nameserverRegex := regexp.MustCompile(`^nameserver\s+(.+)`)

			for _, line := range lines {
				matches := nameserverRegex.FindStringSubmatch(line)
				if len(matches) == 2 {
					dnsServers = append(dnsServers, matches[1])
				}
			}
		}
	} else {
		// Windows-specific DNS server detection
		cmd := exec.Command("cmd", "/c", "ipconfig", "/all")
		output, err := cmd.Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			for lineIdx, line := range lines {
				if strings.Contains(line, "DNS Servers") {
					parts := strings.Split(line, ":")
					if len(parts) >= 2 {
						dnsServer := strings.TrimSpace(parts[1])
						if dnsServer != "" {
							dnsServers = append(dnsServers, dnsServer)
						}

						// Check for additional DNS servers on subsequent lines
						for j := lineIdx + 1; j < len(lines); j++ {
							if !strings.Contains(lines[j], ":") && strings.TrimSpace(lines[j]) != "" {
								dnsServers = append(dnsServers, strings.TrimSpace(lines[j]))
							} else {
								break
							}
						}
					}
				}
			}
		}
	}

	// If no DNS servers found, return a default
	if len(dnsServers) == 0 {
		dnsServers = append(dnsServers, "Unknown")
	}

	return dnsServers
}

// getInterfaceSpeed tries to determine network interface speed
func getInterfaceSpeed(ifaceName string) string {
	if runtime.GOOS == "linux" {
		cmd := exec.Command("cat", fmt.Sprintf("/sys/class/net/%s/speed", ifaceName))
		output, err := cmd.Output()
		if err == nil {
			speed := strings.TrimSpace(string(output))
			if speed != "" {
				// Convert to a readable format
				speedInt := 0
				fmt.Sscanf(speed, "%d", &speedInt)
				if speedInt > 0 {
					if speedInt >= 1000 {
						return fmt.Sprintf("%.1f Gbps", float64(speedInt)/1000.0)
					}
					return fmt.Sprintf("%d Mbps", speedInt)
				}
			}
		}
	}
	return ""
}

// isVPNInterface tries to determine if an interface is a VPN
func isVPNInterface(name string) bool {
	// Simple heuristic based on common VPN interface names
	vpnPatterns := []string{"tun", "tap", "ppp", "vpn", "wg", "ipsec", "wireguard"}
	nameLower := strings.ToLower(name)
	
	for _, pattern := range vpnPatterns {
		if strings.Contains(nameLower, pattern) {
			return true
		}
	}
	return false
}

// isWifiInterface detects if interface is wireless
func isWifiInterface(name string) bool {
	// Check for common wireless interface naming patterns
	if strings.HasPrefix(name, "wl") || strings.Contains(name, "wifi") || strings.Contains(name, "wlan") {
		return true
	}
	
	// On Linux, check if there's a wireless directory in sysfs
	if runtime.GOOS == "linux" {
		cmd := exec.Command("ls", fmt.Sprintf("/sys/class/net/%s/wireless", name))
		if err := cmd.Run(); err == nil {
			return true
		}
	}
	
	return false
}

// truncateIPv6 shortens IPv6 addresses for display
func truncateIPv6(ip string) string {
	// If it's already short enough, return as is
	if len(ip) <= 25 {
		return ip
	}
	
	// Otherwise truncate the middle part
	parsedIP := net.ParseIP(ip)
	if parsedIP != nil {
		return parsedIP.String()
	}
	
	// Fallback to simple truncation
	return ip[:10] + "..." + ip[len(ip)-10:]
}
