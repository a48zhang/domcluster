package connections

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	pb "domcluster/api/proto"
)

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
}

// NewManager 创建管理器
func NewManager(config *Config) *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	return &Manager{
		config: config,
		ctx:    ctx,
		cancel: cancel,
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

	return m.send("register", nodeID, dataBytes)
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

	return m.send("heartbeat", nodeID, dataBytes)
}

// send 发送消息
func (m *Manager) send(cmd, reqID string, data []byte) error {
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
			fmt.Printf("Receive error: %v\n", err)
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
	log.Printf("Received response: reporter=%s, req_id=%s, status=%d", resp.Reporter, resp.ReqId, resp.Status)
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
			log.Printf("Failed to connect to server, retrying in %v: %v", connectRetryInterval, err)
			select {
			case <-time.After(connectRetryInterval):
				continue
			case <-ctx.Done():
				return fmt.Errorf("context cancelled while connecting to server: %w", ctx.Err())
			}
		}
		log.Printf("Connected to server")
		break
	}

	// 重试注册节点
	registerRetryInterval := 5 * time.Second
	for {
		if err := m.RegisterNode(nodeID, nodeName); err != nil {
			log.Printf("Failed to register node, retrying in %v: %v", registerRetryInterval, err)
			select {
			case <-time.After(registerRetryInterval):
				continue
			case <-ctx.Done():
				return fmt.Errorf("context cancelled while registering node: %w", ctx.Err())
			}
		}
		log.Printf("Node %s registered successfully", nodeID)
		break
	}

	return nil
}
