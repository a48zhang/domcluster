package connections

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"go.uber.org/zap"
)

// Config 服务器配置
type Config struct {
	Address   string // 监听地址
	CertFile  string // 服务器证书
	KeyFile   string // 服务器密钥
	CAFile    string // CA 证书
}

// Server gRPC 服务器
type Server struct {
	server *grpc.Server
	config *Config
}

// NewServer 创建服务器
func NewServer(config *Config) (*Server, error) {
	var opts []grpc.ServerOption

	// TLS 配置
	if config.CertFile != "" && config.KeyFile != "" {
		tlsConfig := &tls.Config{
			ClientAuth: tls.RequireAndVerifyClientCert,
		}

		cert, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load server cert: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}

		if config.CAFile != "" {
			caCert, err := os.ReadFile(config.CAFile)
			if err != nil {
				return nil, fmt.Errorf("failed to read CA cert: %w", err)
			}
			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(caCert)
			tlsConfig.ClientCAs = caCertPool
		}

		opts = append(opts, grpc.Creds(credentials.NewTLS(tlsConfig)))
	}

	// Keepalive
	opts = append(opts, grpc.KeepaliveParams(keepalive.ServerParameters{}))

	opts = append(opts, grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
		MinTime:             10 * time.Second,
		PermitWithoutStream: true,
	}))

	return &Server{
		config: config,
		server: grpc.NewServer(opts...),
	}, nil
}

// Start 启动服务器
func (s *Server) Start(ctx context.Context) error {
	lis, err := net.Listen("tcp", s.config.Address)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	zap.L().Sugar().Infof("Server listening on %s", s.config.Address)
	
	// 启动 goroutine 监听 context 取消
	go func() {
		<-ctx.Done()
		zap.L().Sugar().Info("Context cancelled, stopping server...")
		s.Stop()
	}()
	
	return s.server.Serve(lis)
}

// Stop 停止服务器
func (s *Server) Stop() {
	s.server.GracefulStop()
}

// GetServer 获取 gRPC 服务器实例
func (s *Server) GetServer() *grpc.Server {
	return s.server
}