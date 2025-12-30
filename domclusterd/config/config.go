package config

import (
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Config 配置
type Config struct {
	Address string
	Role    string
	UseTLS  bool
	Timeout time.Duration
}

// Load 加载配置，优先级：命令行参数 > 环境变量 > 配置文件 > 默认值
func Load() (*Config, error) {
	v := viper.New()

	// 设置默认值
	v.SetDefault("domclusterd.config.address", "localhost:50051")
	v.SetDefault("domclusterd.service.role", []string{"judgehost"})
	v.SetDefault("domclusterd.config.use_tls", false)
	v.SetDefault("domclusterd.config.timeout", 10)

	// 绑定命令行参数
	pflag.String("address", "localhost:50051", "服务地址")
	pflag.String("role", "judgehost", "节点角色")
	pflag.Bool("tls", false, "启用TLS")
	pflag.Int("timeout", 10, "连接超时时间(秒)")
	pflag.Parse()

	v.BindPFlags(pflag.CommandLine)

	// 自动读取环境变量
	v.SetEnvPrefix("DOMCLUSTER")
	v.AutomaticEnv()

	// 读取配置文件
	v.SetConfigFile("config.yaml")
	v.ReadInConfig()

	// 解析到结构体
	roles := v.GetStringSlice("domclusterd.service.role")
	role := "judgehost"
	if len(roles) > 0 {
		role = roles[0]
	}

	cfg := &Config{
		Address: v.GetString("domclusterd.config.address"),
		Role:    role,
		UseTLS:  v.GetBool("domclusterd.config.use_tls"),
		Timeout: time.Duration(v.GetInt("domclusterd.config.timeout")) * time.Second,
	}

	return cfg, nil
}

// GetAddress 获取服务地址
func (c *Config) GetAddress() string {
	return c.Address
}

// GetRole 获取角色
func (c *Config) GetRole() string {
	return c.Role
}

// GetUseTLS 获取是否使用TLS
func (c *Config) GetUseTLS() bool {
	return c.UseTLS
}

// GetTimeout 获取连接超时时间
func (c *Config) GetTimeout() time.Duration {
	return c.Timeout
}