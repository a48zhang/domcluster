package services

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// SSHConfig SSH连接配置
type SSHConfig struct {
	Host       string
	Port       string
	User       string
	Password   string
	KeyFile    string
	Timeout    time.Duration
}

// SSHManager SSH管理器
type SSHManager struct {
	config *SSHConfig
	client *ssh.Client
}

// NewSSHManager 创建SSH管理器
func NewSSHManager(config *SSHConfig) *SSHManager {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	return &SSHManager{
		config: config,
	}
}

// ParseSSHConnectionString 解析SSH连接字符串
// 支持格式: user@host:port 或 user@host
func ParseSSHConnectionString(connStr string) (*SSHConfig, error) {
	// 解析 user@host:port 格式
	parts := strings.Split(connStr, "@")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid SSH connection string format, expected user@host[:port]")
	}

	user := parts[0]
	hostPort := parts[1]

	var host, port string
	if strings.Contains(hostPort, ":") {
		hostParts := strings.Split(hostPort, ":")
		host = hostParts[0]
		port = hostParts[1]
	} else {
		host = hostPort
		port = "22"
	}

	return &SSHConfig{
		Host: host,
		Port: port,
		User: user,
	}, nil
}

// Connect 建立SSH连接
func (sm *SSHManager) Connect() error {
	var authMethods []ssh.AuthMethod

	// 优先使用密钥认证
	if sm.config.KeyFile != "" {
		key, err := os.ReadFile(sm.config.KeyFile)
		if err != nil {
			return fmt.Errorf("failed to read key file: %w", err)
		}

		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return fmt.Errorf("failed to parse private key: %w", err)
		}

		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	// 使用密码认证
	if sm.config.Password != "" {
		authMethods = append(authMethods, ssh.Password(sm.config.Password))
	}

	if len(authMethods) == 0 {
		return fmt.Errorf("no authentication method provided (password or key required)")
	}

	clientConfig := &ssh.ClientConfig{
		User:            sm.config.User,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 生产环境应该验证主机密钥
		Timeout:         sm.config.Timeout,
	}

	addr := net.JoinHostPort(sm.config.Host, sm.config.Port)
	client, err := ssh.Dial("tcp", addr, clientConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", addr, err)
	}

	sm.client = client
	return nil
}

// ExecuteCommand 执行命令
func (sm *SSHManager) ExecuteCommand(cmd string) (string, error) {
	if sm.client == nil {
		return "", fmt.Errorf("SSH client not connected")
	}

	session, err := sm.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return string(output), fmt.Errorf("command failed: %w", err)
	}

	return string(output), nil
}

// UploadFile 上传文件到远程主机
func (sm *SSHManager) UploadFile(localPath, remotePath string) error {
	if sm.client == nil {
		return fmt.Errorf("SSH client not connected")
	}

	// 读取本地文件
	data, err := os.ReadFile(localPath)
	if err != nil {
		return fmt.Errorf("failed to read local file: %w", err)
	}

	// 创建远程目录
	remoteDir := remotePath[:strings.LastIndex(remotePath, "/")]
	if remoteDir != "" {
		_, err = sm.ExecuteCommand(fmt.Sprintf("mkdir -p %s", remoteDir))
		if err != nil {
			return fmt.Errorf("failed to create remote directory: %w", err)
		}
	}

	// 使用SCP协议上传文件
	session, err := sm.client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	go func() {
		w, _ := session.StdinPipe()
		defer w.Close()
		fmt.Fprintf(w, "C0755 %d %s\n", len(data), remotePath[strings.LastIndex(remotePath, "/")+1:])
		w.Write(data)
		fmt.Fprint(w, "\x00")
	}()

	cmd := fmt.Sprintf("scp -t %s", remotePath)
	if err := session.Run(cmd); err != nil {
		// SCP可能不可用，尝试使用cat命令
		_, err = sm.ExecuteCommand(fmt.Sprintf("cat > %s << 'EOF'\n%s\nEOF", remotePath, string(data)))
		if err != nil {
			return fmt.Errorf("failed to upload file: %w", err)
		}
	}

	// 设置可执行权限
	_, err = sm.ExecuteCommand(fmt.Sprintf("chmod +x %s", remotePath))
	if err != nil {
		return fmt.Errorf("failed to set executable permission: %w", err)
	}

	return nil
}

// Close 关闭SSH连接
func (sm *SSHManager) Close() error {
	if sm.client != nil {
		return sm.client.Close()
	}
	return nil
}

// TestConnection 测试SSH连接
func (sm *SSHManager) TestConnection() error {
	if err := sm.Connect(); err != nil {
		return err
	}
	defer sm.Close()

	_, err := sm.ExecuteCommand("echo 'connection test'")
	return err
}

// GetSystemInfo 获取远程系统信息
func (sm *SSHManager) GetSystemInfo() (map[string]string, error) {
	if sm.client == nil {
		return nil, fmt.Errorf("SSH client not connected")
	}

	info := make(map[string]string)

	// 获取操作系统信息
	if output, err := sm.ExecuteCommand("uname -s"); err == nil {
		info["os"] = strings.TrimSpace(output)
	}

	// 获取架构信息
	if output, err := sm.ExecuteCommand("uname -m"); err == nil {
		info["arch"] = strings.TrimSpace(output)
	}

	// 获取主机名
	if output, err := sm.ExecuteCommand("hostname"); err == nil {
		info["hostname"] = strings.TrimSpace(output)
	}

	return info, nil
}

// StreamCommand 执行命令并实时输出
func (sm *SSHManager) StreamCommand(cmd string, writer io.Writer) error {
	if sm.client == nil {
		return fmt.Errorf("SSH client not connected")
	}

	session, err := sm.client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	session.Stdout = writer
	session.Stderr = writer

	if err := session.Run(cmd); err != nil {
		return fmt.Errorf("command failed: %w", err)
	}

	return nil
}
