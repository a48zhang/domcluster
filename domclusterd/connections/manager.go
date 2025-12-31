package connections

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	pb "domcluster/api/proto"
	"go.uber.org/zap"
)

// HandlerFunc 请求处理函数类型
type HandlerFunc func(*pb.PublishResponse) error

// Manager 连接管理器
type Manager struct {
	config    *Config
	client    *Client
	rpcClient pb.DomclusterServiceClient
	ctx       context.Context
	cancel    context.CancelFunc
	stream    pb.DomclusterService_PublishClient
	nodeID    string
	mu        sync.RWMutex
	connected bool
	handlers  map[string]HandlerFunc
}

// NewManager 创建管理器
func NewManager(config *Config) *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	return &Manager{
		config:   config,
		ctx:      ctx,
		cancel:   cancel,
		handlers: make(map[string]HandlerFunc),
	}
}

// Connect 连接到 D8rctl
func (m *Manager) Connect() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	client, err := NewClient(m.config)
	if err != nil {
		return err
	}
	m.client = client
	rpcClient := pb.NewDomclusterServiceClient(client.GetConn())

	// 建立流式连接
	stream, err := rpcClient.Publish(context.Background())
	if err != nil {
		return fmt.Errorf("failed to create stream: %w", err)
	}
	m.rpcClient = rpcClient
	m.stream = stream
	m.connected = true

	// 启动接收协程
	go m.receiveLoop()

	return nil
}

// RegisterNode 注册节点
func (m *Manager) RegisterNode(nodeID, name string) error {
	m.mu.Lock()
	m.nodeID = nodeID
	m.mu.Unlock()

	data := map[string]interface{}{
		"name":    name,
		"version": "1.0.0",
	}
	dataBytes, _ := json.Marshal(data)

	return m.Send("register", nodeID, dataBytes)
}

// SendHeartbeat 发送心跳
func (m *Manager) SendHeartbeat() error {
	m.mu.RLock()
	nodeID := m.nodeID
	m.mu.RUnlock()

	if nodeID == "" {
		return fmt.Errorf("node not registered")
	}

	dataBytes, _ := json.Marshal(map[string]interface{}{
		"timestamp": time.Now().Unix(),
	})

	return m.Send("heartbeat", nodeID, dataBytes)
}

// Send 发送消息
func (m *Manager) Send(cmd, reqID string, data []byte) error {
	m.mu.RLock()
	stream := m.stream
	m.mu.RUnlock()

	if stream == nil {
		return fmt.Errorf("not connected")
	}

	req := &pb.PublishRequest{
		Issuer: m.nodeID,
		ReqId:  reqID,
		Cmd:    cmd,
		Data:   data,
	}

	return stream.Send(req)
}

// receiveLoop 接收消息循环
func (m *Manager) receiveLoop() {
	for {
		resp, err := m.stream.Recv()
		if err != nil {
			zap.L().Error("Receive error", zap.Error(err))
			m.mu.Lock()
			m.connected = false
			m.mu.Unlock()
			return
		}

		m.handleResponse(resp)
	}
}

// handleResponse 处理响应
func (m *Manager) handleResponse(resp *pb.PublishResponse) {
	zap.L().Sugar().Debugf("Received response: reporter=%s, req_id=%s, status=%d", resp.Reporter, resp.ReqId, resp.Status)

	// 获取 handler 并直接调用
	if handler, ok := m.getHandler(resp); ok {
		if err := handler(resp); err != nil {
			zap.L().Sugar().Errorf("Failed to handle response: reporter=%s, error=%v", resp.Reporter, err)
		}
	}
}

// getHandler 获取处理函数
func (m *Manager) getHandler(resp *pb.PublishResponse) (HandlerFunc, bool) {
	// 从响应数据中提取命令类型
	var data map[string]interface{}
	if err := json.Unmarshal(resp.Data, &data); err == nil {
		if cmd, ok := data["cmd"].(string); ok && cmd != "" {
			m.mu.RLock()
			handler, ok := m.handlers[cmd]
			m.mu.RUnlock()
			return handler, ok
		}
	}

	// 默认使用 reporter 作为命令类型
	m.mu.RLock()
	handler, ok := m.handlers[resp.Reporter]
	m.mu.RUnlock()
	return handler, ok
}

// RegisterHandler 注册请求处理函数
func (m *Manager) RegisterHandler(cmd string, handler HandlerFunc) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handlers[cmd] = handler
	zap.L().Sugar().Infof("Handler registered for command: %s", cmd)
}

// UnregisterHandler 取消注册请求处理函数
func (m *Manager) UnregisterHandler(cmd string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.handlers, cmd)
	zap.L().Sugar().Infof("Handler unregistered for command: %s", cmd)
}

// Close 关闭连接
func (m *Manager) Close() {
	m.cancel()
	if m.client != nil {
		m.client.Close()
	}
}


func (m *Manager) Start(ctx context.Context, nodeID, nodeName string) error {
	// 重试连接服务器
	connectRetryInterval := 5 * time.Second
	for {
		if err := m.Connect(); err != nil {
			zap.L().Sugar().Warnf("Failed to connect to server, retrying in %v: %v", connectRetryInterval, err)
			select {
			case <-time.After(connectRetryInterval):
				continue
			case <-ctx.Done():
				return fmt.Errorf("context cancelled while connecting to server: %w", ctx.Err())
			}
		}
		zap.L().Sugar().Infof("Connected to server")
		break
	}

	// 重试注册节点
	registerRetryInterval := 5 * time.Second
	for {
		if err := m.RegisterNode(nodeID, nodeName); err != nil {
			zap.L().Sugar().Warnf("Failed to register node, retrying in %v: %v", registerRetryInterval, err)
			select {
			case <-time.After(registerRetryInterval):
				continue
			case <-ctx.Done():
				return fmt.Errorf("context cancelled while registering node: %w", ctx.Err())
			}
		}
		zap.L().Sugar().Infof("Node %s registered successfully", nodeID)
		break
	}

	return nil
}
