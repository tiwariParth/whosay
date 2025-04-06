package models

import "time"

type Options struct {
	JSONOutput    bool
	InWatchMode   bool
	VerboseOutput bool
	CompactMode   bool
	EnableAlerts  bool
}

type SystemInfo struct {
	OSName             string `json:"os_name"`
	OSVersion          string `json:"os_version"`
	KernelVersion      string `json:"kernel_version"`
	Hostname           string `json:"hostname"`
	IsVirtualized      bool   `json:"is_virtualized"`
	VirtualizationType string `json:"virtualization_type,omitempty"`
	IsSecureBoot       bool   `json:"is_secure_boot"`
	IsWSL              bool   `json:"is_wsl"`
	WSLVersion         string `json:"wsl_version,omitempty"`
	Uptime             string `json:"uptime"`
	GoVersion          string `json:"go_version"`
}

type CPUInfo struct {
	NumCPU       int     `json:"num_cpu"`
	Usage        float64 `json:"usage_percent"`
	Architecture string  `json:"architecture"`
}

type MemoryInfo struct {
	Total     uint64  `json:"total_bytes"`
	Used      uint64  `json:"used_bytes"`
	UsagePerc float64 `json:"usage_percent"`
}

type DiskInfo struct {
	Path       string  `json:"path"`
	Total      uint64  `json:"total_bytes"`
	Free       uint64  `json:"free_bytes"`
	UsagePerc  float64 `json:"usage_percent"`
}

type NetworkInfo struct {
	Interfaces     []NetworkInterface `json:"interfaces"`
	DefaultGateway string             `json:"default_gateway"`
	DNSServers     []string           `json:"dns_servers"`
}

type NetworkInterface struct {
	Name    string   `json:"name"`
	IPv4    []string `json:"ipv4_addresses"`
	IPv6    []string `json:"ipv6_addresses"`
	MAC     string   `json:"mac_address"`
	Status  string   `json:"status"`
	Speed   string   `json:"speed,omitempty"`
	IsVPN   bool     `json:"is_vpn,omitempty"`
	IsWifi  bool     `json:"is_wifi,omitempty"`
}

type ProcessInfo struct {
	PID         int       `json:"pid"`
	PPID        int       `json:"parent_pid,omitempty"`
	Name        string    `json:"name"`
	User        string    `json:"user"`
	CPU         float64   `json:"cpu_percent"`
	Memory      float64   `json:"memory_percent"`
	MemoryRSS   uint64    `json:"memory_rss_kb,omitempty"`
	Status      string    `json:"status,omitempty"`
	StartTime   time.Time `json:"start_time,omitempty"`
	CommandLine string    `json:"command_line,omitempty"`
}

type ProcessDisplay struct {
	SortBy    string
	Ascending bool
	Filter    string
	Limit     int
}

type ContainerInfo struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Image       string    `json:"image"`
	Command     string    `json:"command,omitempty"`
	Status      string    `json:"status"`
	State       string    `json:"state"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
	StartedAt   time.Time `json:"started_at,omitempty"`
	IPAddress   string    `json:"ip_address,omitempty"`
	Ports       []string  `json:"ports,omitempty"`
	CPUPercent  float64   `json:"cpu_percent,omitempty"`
	MemoryUsage uint64    `json:"memory_usage_bytes,omitempty"`
	MemoryLimit uint64    `json:"memory_limit_bytes,omitempty"`
	MemoryPerc  float64   `json:"memory_percent,omitempty"`
}

type BatteryInfo struct {
	IsPresent      bool    `json:"is_present"`
	Percentage     float64 `json:"percentage"`
	TimeRemaining  string  `json:"time_remaining,omitempty"`
	Status         string  `json:"status"`
	Health         string  `json:"health,omitempty"`
	CycleCount     int     `json:"cycle_count,omitempty"`
	PowerDraw      float64 `json:"power_draw_watts,omitempty"`
	Technology     string  `json:"technology,omitempty"`
	DesignCapacity uint64  `json:"design_capacity_mwh,omitempty"`
	FullCapacity   uint64  `json:"full_capacity_mwh,omitempty"`
}

type TemperatureInfo struct {
	CPU       float64            `json:"cpu_temp"`
	GPU       float64            `json:"gpu_temp,omitempty"`
	Components map[string]float64 `json:"components,omitempty"`
	Units     string             `json:"units"`
}

type TemperatureHistoryRecord struct {
	Timestamp  time.Time           `json:"timestamp"`
	CPU        float64             `json:"cpu_temp"`
	GPU        float64             `json:"gpu_temp,omitempty"`
	Components map[string]float64  `json:"components,omitempty"`
}

type NetworkUsageInfo struct {
	Interface       string    `json:"interface"`
	BytesReceived   uint64    `json:"bytes_received"`
	BytesSent       uint64    `json:"bytes_sent"`
	RxRate          float64   `json:"rx_rate_mbps"`
	TxRate          float64   `json:"tx_rate_mbps"`
	PacketsReceived uint64    `json:"packets_received"`
	PacketsSent     uint64    `json:"packets_sent"`
	Errors          uint64    `json:"errors"`
	Timestamp       time.Time `json:"timestamp,omitempty"`
}

type AlertConfig struct {
	Enabled        bool    `json:"enabled"`
	CPUWarning     float64 `json:"cpu_warning_threshold"`
	CPUCritical    float64 `json:"cpu_critical_threshold"`
	MemoryWarning  float64 `json:"memory_warning_threshold"`
	MemoryCritical float64 `json:"memory_critical_threshold"`
	DiskWarning    float64 `json:"disk_warning_threshold"`
	DiskCritical   float64 `json:"disk_critical_threshold"`
}

type AlertLevel int

const (
	Info AlertLevel = iota
	Warning
	Critical
)

type Alert struct {
	Level       AlertLevel
	Title       string
	Message     string
	Resource    string
	Value       float64
	Threshold   float64
	Time        time.Time
	Acknowledged bool
}

type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Content   string    `json:"content"`
	Level     string    `json:"level"`
	Source    string    `json:"source"`
}
