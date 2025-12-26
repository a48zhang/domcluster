package connections

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// Config 客户端配置
type Config struct {
	Address  string        // D8rctl 地址
	CertFile string        // 客户端证书
	KeyFile  string        // 客户端密钥
	CAFile   string        // CA 证书
	Timeout  time.Duration // 连接超时
}

// Client gRPC 客户端
type Client struct {
	conn *grpc.ClientConn
}

// NewClient 创建客户端
func NewClient(config *Config) (*Client, error) {
	var opts []grpc.DialOption

	// TLS 配置
	if config.CertFile != "" && config.KeyFile != "" {
		tlsConfig := &tls.Config{}

		cert, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load client cert: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}

		if config.CAFile != "" {
			caCert, err := os.ReadFile(config.CAFile)
			if err != nil {
				return nil, fmt.Errorf("failed to read CA cert: %w", err)
			}
			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(caCert)
			tlsConfig.RootCAs = caCertPool
		}

		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// Keepalive
	opts = append(opts,
		grpc.WithKeepaliveParams(
			keepalive.ClientParameters{
				PermitWithoutStream: true,
			}))

	conn, err := grpc.NewClient(config.Address, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", config.Address, err)
	}

	return &Client{conn: conn}, nil
}

// GetConn 获取 gRPC 连接
func (c *Client) GetConn() *grpc.ClientConn {
	return c.conn
}

// Close 关闭连接
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
