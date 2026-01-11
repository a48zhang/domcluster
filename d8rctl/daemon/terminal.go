package daemon

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"d8rctl/services"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源，生产环境应该检查origin
	},
}

// TerminalSession 终端会话
type TerminalSession struct {
	nodeID    string
	sessionID string
	conn      *websocket.Conn
	server    *services.DomclusterServer
	closeChan chan struct{}
	mu        sync.Mutex
}

// handleTerminalWebSocket 处理终端 WebSocket 连接
func (hs *HTTPServer) handleTerminalWebSocket(c *gin.Context) {
	nodeID := c.Query("node_id")
	if nodeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "node_id is required"})
		return
	}

	// 升级为 WebSocket 连接
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		zap.L().Sugar().Errorf("Failed to upgrade to websocket: %v", err)
		return
	}
	defer conn.Close()

	sessionID := fmt.Sprintf("terminal_%d", time.Now().UnixNano())
	
	session := &TerminalSession{
		nodeID:    nodeID,
		sessionID: sessionID,
		conn:      conn,
		server:    hs.svc.(*services.DomclusterServer),
		closeChan: make(chan struct{}),
	}

	zap.L().Sugar().Infof("Terminal session %s started for node %s", sessionID, nodeID)

	// 发送欢迎消息
	welcomeMsg := fmt.Sprintf("Connected to node %s\r\n", nodeID)
	if err := session.writeMessage(welcomeMsg); err != nil {
		zap.L().Sugar().Errorf("Failed to send welcome message: %v", err)
		return
	}

	// 处理消息
	go session.handleMessages()

	// 等待连接关闭
	<-session.closeChan
	zap.L().Sugar().Infof("Terminal session %s closed", sessionID)
}

// handleMessages 处理来自客户端的消息
func (s *TerminalSession) handleMessages() {
	defer close(s.closeChan)

	for {
		_, message, err := s.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				zap.L().Sugar().Errorf("WebSocket error: %v", err)
			}
			return
		}

		// 解析消息
		var msg map[string]interface{}
		if err := json.Unmarshal(message, &msg); err != nil {
			zap.L().Sugar().Warnf("Failed to parse message: %v", err)
			continue
		}

		msgType, ok := msg["type"].(string)
		if !ok {
			continue
		}

		switch msgType {
		case "input":
			// 处理用户输入
			input, ok := msg["data"].(string)
			if !ok {
				continue
			}
			s.handleInput(input)

		case "resize":
			// 处理终端大小调整
			// 这里可以实现终端大小调整功能
			zap.L().Sugar().Debugf("Terminal resize: %v", msg)
		}
	}
}

// handleInput 处理用户输入
func (s *TerminalSession) handleInput(input string) {
	// 将输入发送到节点执行
	reqID := fmt.Sprintf("shell_%d", time.Now().UnixNano())
	
	data := map[string]interface{}{
		"command": input,
		"session_id": s.sessionID,
	}
	
	dataBytes, err := json.Marshal(data)
	if err != nil {
		zap.L().Sugar().Errorf("Failed to marshal shell command: %v", err)
		return
	}

	// 注册响应处理器
	responseChan := make(chan []byte, 1)
	s.server.RegisterShellResponse(reqID, responseChan)

	// 发送命令到节点
	if err := s.server.SendToNode(s.nodeID, "shell_exec", reqID, dataBytes); err != nil {
		zap.L().Sugar().Errorf("Failed to send shell command: %v", err)
		s.writeMessage(fmt.Sprintf("Error: %v\r\n", err))
		return
	}

	// 等待响应（带超时）
	select {
	case response := <-responseChan:
		var result map[string]interface{}
		if err := json.Unmarshal(response, &result); err != nil {
			zap.L().Sugar().Errorf("Failed to unmarshal response: %v", err)
			return
		}

		// 发送输出到客户端
		if output, ok := result["output"].(string); ok {
			s.writeMessage(output)
		}
		if stderr, ok := result["stderr"].(string); ok && stderr != "" {
			s.writeMessage(stderr)
		}

	case <-time.After(30 * time.Second):
		s.writeMessage("Command timeout\r\n")
	}
}

// writeMessage 写入消息到 WebSocket
func (s *TerminalSession) writeMessage(message string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	msg := map[string]interface{}{
		"type": "output",
		"data": message,
	}

	return s.conn.WriteJSON(msg)
}
