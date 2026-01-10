package monitor

import "time"

// HostInfo 主机基本信息
type HostInfo struct {
	Hostname     string `json:"hostname"`
	OS           string `json:"os"`
	Architecture string `json:"architecture"`
	GoVersion    string `json:"go_version"`
	NumCPU       int    `json:"num_cpu"`
}

// CPUInfo CPU信息
type CPUInfo struct {
	CoreCount    int     `json:"core_count"`
	UsagePercent float64 `json:"usage_percent"`
}

// MemoryInfo 内存信息
type MemoryInfo struct {
	Total        uint64  `json:"total"`
	Used         uint64  `json:"used"`
	Available    uint64  `json:"available"`
	UsagePercent float64 `json:"usage_percent"`
}

// DiskInfo 磁盘信息
type DiskInfo struct {
	Path         string  `json:"path"`
	Total        uint64  `json:"total"`
	Used         uint64  `json:"used"`
	Free         uint64  `json:"free"`
	UsagePercent float64 `json:"usage_percent"`
}

// NetworkInfo 网络信息
type NetworkInfo struct {
	RxBytes uint64 `json:"rx_bytes"`
	TxBytes uint64 `json:"tx_bytes"`
}

// SystemResources 系统资源信息
type SystemResources struct {
	CPU     *CPUInfo     `json:"cpu"`
	Memory  *MemoryInfo  `json:"memory"`
	Disk    *DiskInfo    `json:"disk"`
	Network *NetworkInfo `json:"network"`
}

// DockerContainer Docker容器信息
type DockerContainer struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Image     string `json:"image"`
	Status    string `json:"status"`
	Ports     string `json:"ports"`
	CreatedAt string `json:"created_at"`
}

// DockerInfo Docker信息
type DockerInfo struct {
	RunningCount int               `json:"running_count"`
	TotalCount   int               `json:"total_count"`
	Containers   []DockerContainer `json:"containers"`
}

// NodeStatus 节点状态
type NodeStatus struct {
	NodeID           string            `json:"node_id"`
	LastUpdate       time.Time         `json:"last_update"`
	Host             *HostInfo         `json:"host"`
	SystemResources  *SystemResources  `json:"system_resources"`
	Docker           *DockerInfo       `json:"docker"`
	Online           bool              `json:"online"`
	Metadata         map[string]string `json:"metadata,omitempty"`
}