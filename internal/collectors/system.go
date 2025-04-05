package collectors

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/tiwariParth/whosay/internal/models"
	"github.com/tiwariParth/whosay/internal/ui"
)

// GetSystemInfo displays detailed system information
func GetSystemInfo(opts models.Options) {
	info := CollectSystemInfo()

	if opts.JSONOutput {
		jsonData, _ := json.MarshalIndent(info, "", "  ")
		fmt.Println(string(jsonData))
		return
	}

	// Display using the compact layout
	sections := GetSystemInfoSections(opts)
	ui.CompactDisplay(sections)
}

// CollectSystemInfo gathers detailed system information
func CollectSystemInfo() models.SystemInfo {
	info := models.SystemInfo{
		OSName:        detectOSName(),
		OSVersion:     detectOSVersion(),
		KernelVersion: detectKernelVersion(),
		Hostname:      detectHostname(),
		GoVersion:     runtime.Version(),
		Uptime:        detectUptime(),
	}

	// Check if running in WSL
	isWSL, wslVersion := detectWSL()
	info.IsWSL = isWSL
	info.WSLVersion = wslVersion

	// Check virtualization
	isVirt, virtType := detectVirtualization()
	info.IsVirtualized = isVirt
	info.VirtualizationType = virtType

	// Check secure boot
	info.IsSecureBoot = detectSecureBoot()

	return info
}

// GetSystemInfoSections returns the system information as formatted sections
func GetSystemInfoSections(opts models.Options) map[string][][]string {
	info := CollectSystemInfo()
	
	// Enhance system data presentation for developer environments
	sysData := [][]string{
		{"OS", fmt.Sprintf("%s %s", info.OSName, info.OSVersion)},
		{"Hostname", info.Hostname},
		{"Kernel", info.KernelVersion},
		{"Architecture", getArchDescription(runtime.GOARCH)},
		{"Uptime", info.Uptime},
	}
	
	// Enhance runtime information with more developer-relevant details
	virtData := [][]string{
		{"Go Runtime", info.GoVersion},
		{"Shell", getDefaultShell()},
		{"Terminal", getTerminalName()},
	}
	
	if info.IsVirtualized {
		virtData = append(virtData, []string{"Virtualization", info.VirtualizationType})
	} else {
		virtData = append(virtData, []string{"Virtualization", "Native Hardware"})
	}
	
	if info.IsWSL {
		virtData = append(virtData, []string{"WSL", info.WSLVersion})
	}
	
	secureBootStatus := "Disabled"
	if info.IsSecureBoot {
		secureBootStatus = "Enabled"
	}
	virtData = append(virtData, []string{"Secure Boot", secureBootStatus})
	
	// Return data for unified display
	return map[string][][]string{
		"System": sysData,
		"Runtime Environment": virtData,
	}
}

// GetArchInfo returns formatted architecture information
func GetArchInfo() string {
	return getArchDescription(runtime.GOARCH)
}

// Helper functions for system information

func getDefaultShell() string {
	shell := os.Getenv("SHELL")
	if shell == "" {
		return "Unknown"
	}
	
	// Extract just the shell name from the path
	parts := strings.Split(shell, "/")
	return parts[len(parts)-1]
}

func getTerminalName() string {
	// Try to detect terminal from environment variables
	term := os.Getenv("TERM")
	if term == "" {
		return "Unknown"
	}
	
	// Clean up common terminal names
	if strings.Contains(term, "xterm") {
		return "xterm"
	} else if strings.Contains(term, "rxvt") {
		return "rxvt"
	} else if strings.Contains(term, "konsole") {
		return "Konsole"
	} else if strings.Contains(term, "gnome") {
		return "GNOME Terminal"
	}
	
	return term
}

// detectOSName detects the operating system name
func detectOSName() string {
	switch runtime.GOOS {
	case "windows":
		return "Microsoft Windows"
	case "darwin":
		return "macOS"
	case "linux":
		// Try to get Linux distribution name
		if distro := readFile("/etc/os-release", "NAME"); distro != "" {
			return strings.Trim(distro, "\"")
		}
		return "Linux"
	default:
		return runtime.GOOS
	}
}

// detectOSVersion gets the OS version
func detectOSVersion() string {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("cmd", "/c", "ver")
		output, err := cmd.Output()
		if err == nil {
			return strings.TrimSpace(string(output))
		}
	case "darwin":
		cmd := exec.Command("sw_vers", "-productVersion")
		output, err := cmd.Output()
		if err == nil {
			return strings.TrimSpace(string(output))
		}
	case "linux":
		// Try to get Linux version
		if version := readFile("/etc/os-release", "VERSION_ID"); version != "" {
			return strings.Trim(version, "\"")
		}
	}
	return "Unknown"
}

// detectKernelVersion gets the kernel version
func detectKernelVersion() string {
	switch runtime.GOOS {
	case "linux":
		cmd := exec.Command("uname", "-r")
		output, err := cmd.Output()
		if err == nil {
			return strings.TrimSpace(string(output))
		}
	case "darwin":
		cmd := exec.Command("uname", "-r")
		output, err := cmd.Output()
		if err == nil {
			return strings.TrimSpace(string(output))
		}
	case "windows":
		return "NT Kernel"
	}
	return "Unknown"
}

// detectHostname gets the system hostname
func detectHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "Unknown"
	}
	return hostname
}

// detectVirtualization checks if running in a virtualized environment
func detectVirtualization() (bool, string) {
	if runtime.GOOS == "linux" {
		// Check systemd-detect-virt
		cmd := exec.Command("systemd-detect-virt")
		output, err := cmd.Output()
		if err == nil {
			virt := strings.TrimSpace(string(output))
			if virt != "none" {
				return true, virt
			}
		}
		
		// Check /proc/cpuinfo for hypervisor flag
		content, err := os.ReadFile("/proc/cpuinfo")
		if err == nil {
			if strings.Contains(string(content), "hypervisor") {
				return true, "Unknown Hypervisor"
			}
		}

		// Check DMI for virtual machine indicators
		dmiContent, err := os.ReadFile("/sys/class/dmi/id/product_name")
		if err == nil {
			product := strings.TrimSpace(string(dmiContent))
			if strings.Contains(product, "VMware") {
				return true, "VMware"
			} else if strings.Contains(product, "VirtualBox") {
				return true, "VirtualBox"
			} else if strings.Contains(product, "Virtual Machine") {
				return true, "Hyper-V"
			} else if strings.Contains(product, "QEMU") {
				return true, "QEMU/KVM"
			}
		}
	}
	return false, ""
}

// detectSecureBoot checks if secure boot is enabled
func detectSecureBoot() bool {
	if runtime.GOOS == "linux" {
		// Check if mokutil exists
		_, err := exec.LookPath("mokutil")
		if err == nil {
			cmd := exec.Command("mokutil", "--sb-state")
			output, err := cmd.Output()
			if err == nil {
				return strings.Contains(string(output), "SecureBoot enabled")
			}
		}
		
		// Alternative check through EFI vars
		_, err = os.Stat("/sys/firmware/efi")
		if err == nil {
			content, err := os.ReadFile("/sys/firmware/efi/efivars/SecureBoot-8be4df61-93ca-11d2-aa0d-00e098032b8c")
			if err == nil {
				// The 5th byte indicates secure boot state (1 for enabled, 0 for disabled)
				if len(content) >= 5 && content[4] == 1 {
					return true
				}
			}
		}
	}
	return false
}

// detectWSL checks if running in Windows Subsystem for Linux
func detectWSL() (bool, string) {
	if runtime.GOOS == "linux" {
		// Check if /proc/version contains Microsoft
		content, err := os.ReadFile("/proc/version")
		if err == nil && strings.Contains(string(content), "Microsoft") {
			// Determine WSL version
			if strings.Contains(string(content), "WSL2") {
				return true, "WSL2"
			}
			return true, "WSL1"
		}
	}
	return false, ""
}

// detectUptime gets the system uptime
func detectUptime() string {
	if runtime.GOOS == "linux" {
		uptime, err := os.ReadFile("/proc/uptime")
		if err == nil {
			parts := strings.Split(string(uptime), " ")
			if len(parts) > 0 {
				seconds := 0.0
				fmt.Sscanf(parts[0], "%f", &seconds)
				
				days := int(seconds / 86400)
				hours := int((seconds - float64(days)*86400) / 3600)
				minutes := int((seconds - float64(days)*86400 - float64(hours)*3600) / 60)
				
				return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
			}
		}
	}
	
	// Fallback
	return "Unknown"
}

// readFile reads a specific value from a file with key=value format
func readFile(path, key string) string {
	file, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, key+"=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				return parts[1]
			}
		}
	}
	return ""
}
