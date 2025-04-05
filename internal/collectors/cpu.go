package collectors

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"runtime"
	"time"

	"github.com/tiwariParth/whosay/internal/models"
	"github.com/tiwariParth/whosay/internal/ui"
)

// GetCPUInfo displays CPU information
func GetCPUInfo(opts models.Options) {
	info := collectCPUInfo()

	if opts.JSONOutput {
		jsonData, _ := json.MarshalIndent(info, "", "  ")
		fmt.Println(string(jsonData))
		return
	}

	sections := GetCPUInfoSections(opts)
	ui.CompactDisplay(sections)
}

// GetCPUInfoSections returns formatted CPU information sections
func GetCPUInfoSections(opts models.Options) map[string][][]string {
	info := collectCPUInfo()
	
	cpuData := [][]string{
		{"CPUs", fmt.Sprintf("%d cores", info.NumCPU)},
		{"Architecture", getArchDescription(info.Architecture)},
		{"Usage", fmt.Sprintf("%.1f%%", info.Usage)},
	}
	
	barWidth := 20
	if opts.CompactMode {
		barWidth = 15
	}
	
	cpuData = append(cpuData, []string{
		"", ui.PrintCompactUsageBar("", info.Usage, barWidth),
	})
	
	return map[string][][]string{
		"CPU": cpuData,
	}
}

// collectCPUInfo gathers CPU information
func collectCPUInfo() models.CPUInfo {
	return models.CPUInfo{
		NumCPU:       runtime.NumCPU(),
		Usage:        getCPUUsage(),
		Architecture: runtime.GOARCH,
	}
}

func getCPUUsage() float64 {
	// Simplified for demo purposes - in production use gopsutil or similar
	rand.Seed(time.Now().UnixNano())
	baseValue := 30.0 + (rand.Float64() * 40.0)
	variation := rand.Float64() * 10.0 - 5.0
	
	return baseValue + variation
}

// getArchDescription formats architecture information for readability
func getArchDescription(arch string) string {
	switch arch {
	case "amd64":
		return "x86_64 (64-bit)"
	case "386":
		return "x86 (32-bit)"
	case "arm":
		return "ARM"
	case "arm64":
		return "ARM64"
	case "ppc64":
		return "PowerPC 64-bit"
	case "ppc64le":
		return "PowerPC 64-bit LE"
	case "mips":
		return "MIPS"
	case "mipsle":
		return "MIPS LE"
	case "mips64":
		return "MIPS64"
	case "mips64le":
		return "MIPS64 LE"
	case "s390x":
		return "IBM S/390"
	case "riscv64":
		return "RISC-V 64-bit"
	}
	
	return arch
}
