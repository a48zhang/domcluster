package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"go.uber.org/zap"
)

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

// SystemResources 系统资源信息
type SystemResources struct {
	CPU    *CPUInfo    `json:"cpu"`
	Memory *MemoryInfo `json:"memory"`
	Disk   *DiskInfo   `json:"disk"`
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

// Monitor 监控器
type Monitor struct {
	ctx      context.Context
	diskPath string
}

// NewMonitor 创建监控器
func NewMonitor(ctx context.Context) *Monitor {
	return &Monitor{
		ctx:      ctx,
		diskPath: "/",
	}
}

// GetHostInfo 获取主机基本信息
func (m *Monitor) GetHostInfo() (*HostInfo, error) {
	hostname, err := os.Hostname()
	if err != nil {
		zap.L().Error("failed to get hostname", zap.Error(err))
		hostname = "unknown"
	}

	return &HostInfo{
		Hostname:     hostname,
		OS:           runtime.GOOS,
		Architecture: runtime.GOARCH,
		GoVersion:    runtime.Version(),
		NumCPU:       runtime.NumCPU(),
	}, nil
}

// GetSystemResources 获取系统资源使用情况
func (m *Monitor) GetSystemResources() (*SystemResources, error) {
	cpuInfo, err := m.getCPUInfo()
	if err != nil {
		zap.L().Warn("failed to get CPU info", zap.Error(err))
	}

	memInfo, err := m.getMemoryInfo()
	if err != nil {
		zap.L().Warn("failed to get memory info", zap.Error(err))
	}

	diskInfo, err := m.getDiskInfo(m.diskPath)
	if err != nil {
		zap.L().Warn("failed to get disk info", zap.Error(err))
	}

	return &SystemResources{
		CPU:    cpuInfo,
		Memory: memInfo,
		Disk:   diskInfo,
	}, nil
}

// GetDockerInfo 获取Docker容器情况
func (m *Monitor) GetDockerInfo() (*DockerInfo, error) {
	containers, err := m.getDockerContainers()
	if err != nil {
		return nil, fmt.Errorf("failed to get docker containers: %w", err)
	}

	runningCount := 0
	for _, c := range containers {
		if strings.Contains(c.Status, "Up") {
			runningCount++
		}
	}

	return &DockerInfo{
		RunningCount: runningCount,
		TotalCount:   len(containers),
		Containers:   containers,
	}, nil
}

// getCPUInfo 获取CPU信息
func (m *Monitor) getCPUInfo() (*CPUInfo, error) {
	cmd := exec.Command("sh", "-c", "top -bn1 | grep 'Cpu(s)' | awk '{print $2}' | cut -d'%' -f1")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var usage float64
	_, err = fmt.Sscanf(strings.TrimSpace(string(output)), "%f", &usage)
	if err != nil {
		return &CPUInfo{UsagePercent: 0}, nil
	}

	return &CPUInfo{UsagePercent: usage}, nil
}

// getMemoryInfo 获取内存信息
func (m *Monitor) getMemoryInfo() (*MemoryInfo, error) {
	cmd := exec.Command("sh", "-c", "free -b | grep Mem")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	fields := strings.Fields(string(output))
	if len(fields) < 4 {
		return &MemoryInfo{}, nil
	}

	var total, used, free, available uint64
	fmt.Sscanf(fields[1], "%d", &total)
	fmt.Sscanf(fields[2], "%d", &used)
	fmt.Sscanf(fields[3], "%d", &free)
	if len(fields) >= 7 {
		fmt.Sscanf(fields[6], "%d", &available)
	} else {
		available = free
	}

	usagePercent := float64(used) / float64(total) * 100

	return &MemoryInfo{
		Total:        total,
		Used:         used,
		Available:    available,
		UsagePercent: usagePercent,
	}, nil
}

// getDiskInfo 获取磁盘信息
func (m *Monitor) getDiskInfo(path string) (*DiskInfo, error) {
	cmd := exec.Command("df", "-B1", path)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return &DiskInfo{}, nil
	}

	fields := strings.Fields(lines[1])
	if len(fields) < 5 {
		return &DiskInfo{}, nil
	}

	var total, used, free uint64
	fmt.Sscanf(fields[1], "%d", &total)
	fmt.Sscanf(fields[2], "%d", &used)
	fmt.Sscanf(fields[3], "%d", &free)

	usagePercent := float64(used) / float64(total) * 100

	return &DiskInfo{
		Path:         path,
		Total:        total,
		Used:         used,
		Free:         free,
		UsagePercent: usagePercent,
	}, nil
}

// getDockerContainers 获取Docker容器列表
func (m *Monitor) getDockerContainers() ([]DockerContainer, error) {
	// 检查docker是否可用
	if _, err := exec.LookPath("docker"); err != nil {
		return nil, fmt.Errorf("docker command not found: %w", err)
	}

	cmd := exec.Command("docker", "ps", "-a", "--format", "{{json .}}")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute docker ps: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	containers := make([]DockerContainer, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var container struct {
			ID      string `json:"ID"`
			Names   string `json:"Names"`
			Image   string `json:"Image"`
			Status  string `json:"Status"`
			Ports   string `json:"Ports"`
			Created string `json:"CreatedAt"`
		}

		if err := json.Unmarshal([]byte(line), &container); err != nil {
			zap.L().Warn("failed to parse docker container info", zap.String("line", line), zap.Error(err))
			continue
		}

		containers = append(containers, DockerContainer{
			ID:        container.ID[:12],
			Name:      strings.TrimPrefix(container.Names, "/"),
			Image:     container.Image,
			Status:    container.Status,
			Ports:     container.Ports,
			CreatedAt: container.Created,
		})
	}

	return containers, nil
}

// FormatBytes 格式化字节数
func FormatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// GetMonitorReport 获取监控报告
func (m *Monitor) GetMonitorReport() (map[string]interface{}, error) {
	hostInfo, err := m.GetHostInfo()
	if err != nil {
		zap.L().Warn("failed to get host info", zap.Error(err))
		hostInfo = &HostInfo{
			Hostname:     "unknown",
			OS:           runtime.GOOS,
			Architecture: runtime.GOARCH,
			GoVersion:    runtime.Version(),
			NumCPU:       runtime.NumCPU(),
		}
	}

	systemResources, err := m.GetSystemResources()
	if err != nil {
		zap.L().Warn("failed to get system resources", zap.Error(err))
		systemResources = &SystemResources{
			CPU:    &CPUInfo{},
			Memory: &MemoryInfo{},
			Disk:   &DiskInfo{},
		}
	}

	dockerInfo, err := m.GetDockerInfo()
	if err != nil {
		zap.L().Warn("failed to get docker info", zap.Error(err))
		dockerInfo = &DockerInfo{}
	}

	return map[string]interface{}{
		"timestamp":        time.Now().Format(time.RFC3339),
		"host":             hostInfo,
		"system_resources": systemResources,
		"docker":           dockerInfo,
	}, nil
}