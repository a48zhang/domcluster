package services

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
)

// HostProvisionRequest 主机供应请求
type HostProvisionRequest struct {
	SSHConnectionString string `json:"ssh_connection_string"` // user@host:port
	Password            string `json:"password,omitempty"`
	KeyFile             string `json:"key_file,omitempty"`
	D8rctlAddress       string `json:"d8rctl_address"` // d8rctl服务地址，例如 192.168.1.100:50051
}

// HostProvisionResult 主机供应结果
type HostProvisionResult struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	NodeID    string `json:"node_id,omitempty"`
	Hostname  string `json:"hostname,omitempty"`
	OS        string `json:"os,omitempty"`
	Arch      string `json:"arch,omitempty"`
}

// HostProvisioner 主机供应器
type HostProvisioner struct {
	domclusterdBinary string
}

// NewHostProvisioner 创建主机供应器
func NewHostProvisioner() (*HostProvisioner, error) {
	// 查找domclusterd二进制文件
	binary := findDomclusterdBinary()
	if binary == "" {
		return nil, fmt.Errorf("domclusterd binary not found")
	}

	return &HostProvisioner{
		domclusterdBinary: binary,
	}, nil
}

// findDomclusterdBinary 查找domclusterd二进制文件
func findDomclusterdBinary() string {
	// 可能的位置
	possiblePaths := []string{
		"./built/domclusterd",
		"../built/domclusterd",
		"./domclusterd",
		"../domclusterd/domclusterd",
	}

	// 获取当前可执行文件的目录
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		possiblePaths = append([]string{
			filepath.Join(exeDir, "domclusterd"),
			filepath.Join(exeDir, "../built/domclusterd"),
		}, possiblePaths...)
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			absPath, _ := filepath.Abs(path)
			return absPath
		}
	}

	// 尝试在PATH中查找
	if path, err := exec.LookPath("domclusterd"); err == nil {
		return path
	}

	return ""
}

// ProvisionHost 供应主机
func (hp *HostProvisioner) ProvisionHost(req *HostProvisionRequest) (*HostProvisionResult, error) {
	zap.L().Sugar().Infof("Starting host provision for %s", req.SSHConnectionString)

	// 解析SSH连接字符串
	sshConfig, err := ParseSSHConnectionString(req.SSHConnectionString)
	if err != nil {
		return &HostProvisionResult{
			Success: false,
			Message: fmt.Sprintf("Invalid SSH connection string: %v", err),
		}, err
	}

	// 设置认证信息
	sshConfig.Password = req.Password
	sshConfig.KeyFile = req.KeyFile

	// 创建SSH管理器
	sshMgr := NewSSHManager(sshConfig)

	// 测试SSH连接
	zap.L().Sugar().Info("Testing SSH connection...")
	if err := sshMgr.Connect(); err != nil {
		return &HostProvisionResult{
			Success: false,
			Message: fmt.Sprintf("Failed to connect via SSH: %v", err),
		}, err
	}
	defer sshMgr.Close()

	// 获取系统信息
	zap.L().Sugar().Info("Getting system information...")
	sysInfo, err := sshMgr.GetSystemInfo()
	if err != nil {
		return &HostProvisionResult{
			Success: false,
			Message: fmt.Sprintf("Failed to get system info: %v", err),
		}, err
	}

	hostname := sysInfo["hostname"]
	osType := sysInfo["os"]
	arch := sysInfo["arch"]

	zap.L().Sugar().Infof("Remote system: %s (%s/%s)", hostname, osType, arch)

	// 检查操作系统是否支持
	if !strings.EqualFold(osType, "Linux") {
		return &HostProvisionResult{
			Success: false,
			Message: fmt.Sprintf("Unsupported OS: %s (only Linux is supported)", osType),
			Hostname: hostname,
			OS:       osType,
			Arch:     arch,
		}, fmt.Errorf("unsupported OS: %s", osType)
	}

	// 上传domclusterd二进制文件
	zap.L().Sugar().Info("Uploading domclusterd binary...")
	remotePath := "/tmp/domclusterd"
	if err := hp.uploadBinary(sshMgr, remotePath); err != nil {
		return &HostProvisionResult{
			Success: false,
			Message: fmt.Sprintf("Failed to upload domclusterd: %v", err),
			Hostname: hostname,
			OS:       osType,
			Arch:     arch,
		}, err
	}

	// 创建配置目录
	zap.L().Sugar().Info("Creating configuration directory...")
	configDir := "/var/lib/domcluster"
	if _, err := sshMgr.ExecuteCommand(fmt.Sprintf("sudo mkdir -p %s", configDir)); err != nil {
		zap.L().Sugar().Warnf("Failed to create config dir with sudo, trying without: %v", err)
		if _, err := sshMgr.ExecuteCommand(fmt.Sprintf("mkdir -p %s", configDir)); err != nil {
			return &HostProvisionResult{
				Success: false,
				Message: fmt.Sprintf("Failed to create config directory: %v", err),
				Hostname: hostname,
				OS:       osType,
				Arch:     arch,
			}, err
		}
	}

	// 移动二进制文件到最终位置
	zap.L().Sugar().Info("Installing domclusterd...")
	finalPath := "/usr/local/bin/domclusterd"
	if _, err := sshMgr.ExecuteCommand(fmt.Sprintf("sudo mv %s %s", remotePath, finalPath)); err != nil {
		zap.L().Sugar().Warnf("Failed to move with sudo, trying without: %v", err)
		if _, err := sshMgr.ExecuteCommand(fmt.Sprintf("mv %s %s", remotePath, finalPath)); err != nil {
			// 如果移动失败，使用临时路径
			finalPath = remotePath
			zap.L().Sugar().Warnf("Using temporary path: %s", finalPath)
		}
	}

	// 创建配置文件
	zap.L().Sugar().Info("Creating configuration file...")
	configContent := fmt.Sprintf(`server:
  address: "%s"
  cert_file: ""
  key_file: ""
  ca_file: ""

node:
  id: "%s"
  name: "%s"
`, req.D8rctlAddress, hostname, hostname)

	configPath := filepath.Join(configDir, "config.yaml")
	if err := hp.createConfigFile(sshMgr, configPath, configContent); err != nil {
		return &HostProvisionResult{
			Success: false,
			Message: fmt.Sprintf("Failed to create config file: %v", err),
			Hostname: hostname,
			OS:       osType,
			Arch:     arch,
		}, err
	}

	// 启动domclusterd
	zap.L().Sugar().Info("Starting domclusterd...")
	startCmd := fmt.Sprintf("nohup %s --config %s > /tmp/domclusterd.log 2>&1 &", finalPath, configPath)
	if _, err := sshMgr.ExecuteCommand(startCmd); err != nil {
		return &HostProvisionResult{
			Success: false,
			Message: fmt.Sprintf("Failed to start domclusterd: %v", err),
			Hostname: hostname,
			OS:       osType,
			Arch:     arch,
		}, err
	}

	zap.L().Sugar().Infof("Successfully provisioned host %s", hostname)

	return &HostProvisionResult{
		Success:  true,
		Message:  "Host provisioned successfully",
		NodeID:   hostname,
		Hostname: hostname,
		OS:       osType,
		Arch:     arch,
	}, nil
}

// uploadBinary 上传二进制文件
func (hp *HostProvisioner) uploadBinary(sshMgr *SSHManager, remotePath string) error {
	if err := sshMgr.UploadFile(hp.domclusterdBinary, remotePath); err != nil {
		return err
	}
	return nil
}

// createConfigFile 创建配置文件
func (hp *HostProvisioner) createConfigFile(sshMgr *SSHManager, remotePath, content string) error {
	// 使用heredoc创建配置文件
	cmd := fmt.Sprintf("cat > %s << 'EOFCONFIG'\n%s\nEOFCONFIG", remotePath, content)
	if _, err := sshMgr.ExecuteCommand(cmd); err != nil {
		// 尝试使用sudo
		cmd = fmt.Sprintf("sudo bash -c \"cat > %s << 'EOFCONFIG'\n%s\nEOFCONFIG\"", remotePath, content)
		if _, err := sshMgr.ExecuteCommand(cmd); err != nil {
			return err
		}
	}
	return nil
}

// ProvisionHostWithProgress 供应主机并提供进度反馈
func (hp *HostProvisioner) ProvisionHostWithProgress(req *HostProvisionRequest, progressChan chan<- string) (*HostProvisionResult, error) {
	sendProgress := func(msg string) {
		if progressChan != nil {
			progressChan <- msg
		}
		zap.L().Sugar().Info(msg)
	}

	sendProgress("Starting host provisioning...")

	// 解析SSH连接字符串
	sendProgress("Parsing SSH connection string...")
	sshConfig, err := ParseSSHConnectionString(req.SSHConnectionString)
	if err != nil {
		msg := fmt.Sprintf("Invalid SSH connection string: %v", err)
		sendProgress(msg)
		return &HostProvisionResult{Success: false, Message: msg}, err
	}

	sshConfig.Password = req.Password
	sshConfig.KeyFile = req.KeyFile

	// 创建SSH管理器并连接
	sendProgress("Connecting to remote host via SSH...")
	sshMgr := NewSSHManager(sshConfig)
	if err := sshMgr.Connect(); err != nil {
		msg := fmt.Sprintf("Failed to connect via SSH: %v", err)
		sendProgress(msg)
		return &HostProvisionResult{Success: false, Message: msg}, err
	}
	defer sshMgr.Close()

	// 获取系统信息
	sendProgress("Retrieving system information...")
	sysInfo, err := sshMgr.GetSystemInfo()
	if err != nil {
		msg := fmt.Sprintf("Failed to get system info: %v", err)
		sendProgress(msg)
		return &HostProvisionResult{Success: false, Message: msg}, err
	}

	hostname := sysInfo["hostname"]
	osType := sysInfo["os"]
	arch := sysInfo["arch"]
	sendProgress(fmt.Sprintf("Remote system: %s (%s/%s)", hostname, osType, arch))

	// 检查操作系统
	if !strings.EqualFold(osType, "Linux") {
		msg := fmt.Sprintf("Unsupported OS: %s (only Linux is supported)", osType)
		sendProgress(msg)
		return &HostProvisionResult{
			Success: false, Message: msg,
			Hostname: hostname, OS: osType, Arch: arch,
		}, fmt.Errorf("unsupported OS")
	}

	// 上传二进制文件
	sendProgress("Uploading domclusterd binary...")
	remotePath := "/tmp/domclusterd"
	if err := hp.uploadBinary(sshMgr, remotePath); err != nil {
		msg := fmt.Sprintf("Failed to upload domclusterd: %v", err)
		sendProgress(msg)
		return &HostProvisionResult{
			Success: false, Message: msg,
			Hostname: hostname, OS: osType, Arch: arch,
		}, err
	}

	// 创建配置
	sendProgress("Creating configuration...")
	configDir := "/var/lib/domcluster"
	sshMgr.ExecuteCommand(fmt.Sprintf("mkdir -p %s", configDir))

	finalPath := "/usr/local/bin/domclusterd"
	if _, err := sshMgr.ExecuteCommand(fmt.Sprintf("sudo mv %s %s 2>/dev/null || mv %s %s", remotePath, finalPath, remotePath, finalPath)); err != nil {
		finalPath = remotePath
	}

	configContent := fmt.Sprintf(`server:
  address: "%s"
  cert_file: ""
  key_file: ""
  ca_file: ""

node:
  id: "%s"
  name: "%s"
`, req.D8rctlAddress, hostname, hostname)

	configPath := filepath.Join(configDir, "config.yaml")
	if err := hp.createConfigFile(sshMgr, configPath, configContent); err != nil {
		msg := fmt.Sprintf("Failed to create config file: %v", err)
		sendProgress(msg)
		return &HostProvisionResult{
			Success: false, Message: msg,
			Hostname: hostname, OS: osType, Arch: arch,
		}, err
	}

	// 启动服务
	sendProgress("Starting domclusterd service...")
	startCmd := fmt.Sprintf("nohup %s --config %s > /tmp/domclusterd.log 2>&1 &", finalPath, configPath)
	if _, err := sshMgr.ExecuteCommand(startCmd); err != nil {
		msg := fmt.Sprintf("Failed to start domclusterd: %v", err)
		sendProgress(msg)
		return &HostProvisionResult{
			Success: false, Message: msg,
			Hostname: hostname, OS: osType, Arch: arch,
		}, err
	}

	sendProgress("Host provisioned successfully!")

	return &HostProvisionResult{
		Success: true, Message: "Host provisioned successfully",
		NodeID: hostname, Hostname: hostname, OS: osType, Arch: arch,
	}, nil
}

// TestSSHConnection 测试SSH连接
func TestSSHConnection(req *HostProvisionRequest) error {
	sshConfig, err := ParseSSHConnectionString(req.SSHConnectionString)
	if err != nil {
		return err
	}

	sshConfig.Password = req.Password
	sshConfig.KeyFile = req.KeyFile

	sshMgr := NewSSHManager(sshConfig)
	return sshMgr.TestConnection()
}

// StreamProvisionOutput 流式输出供应进度
func (hp *HostProvisioner) StreamProvisionOutput(req *HostProvisionRequest) (*HostProvisionResult, *bytes.Buffer, error) {
	output := &bytes.Buffer{}

	writeLine := func(msg string) {
		output.WriteString(msg + "\n")
		zap.L().Sugar().Info(msg)
	}

	writeLine("Starting host provisioning...")

	result, err := hp.ProvisionHost(req)
	if err != nil {
		writeLine(fmt.Sprintf("Error: %v", err))
	}

	return result, output, err
}
