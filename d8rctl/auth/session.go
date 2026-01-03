package auth

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

// Session 会话
type Session struct {
	Token     string
	CreatedAt time.Time
	ExpiresAt time.Time
}

// SessionManager 会话管理器
type SessionManager struct {
	sessions map[string]*Session
	mu       sync.RWMutex
}

var (
	sessionInstance *SessionManager
	sessionOnce     sync.Once
)

// GetSessionManager 获取会话管理器单例
func GetSessionManager() *SessionManager {
	sessionOnce.Do(func() {
		sessionInstance = &SessionManager{
			sessions: make(map[string]*Session),
		}
	})
	return sessionInstance
}

// CreateSession 创建新会话
func (sm *SessionManager) CreateSession() (string, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// 生成随机 token
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := hex.EncodeToString(b)

	// 创建会话
	now := time.Now()
	session := &Session{
		Token:     token,
		CreatedAt: now,
		ExpiresAt: now.Add(24 * time.Hour), // 会话有效期 24 小时
	}

	sm.sessions[token] = session

	// 启动清理过期会话的 goroutine
	go sm.cleanupExpiredSessions()

	return token, nil
}

// ValidateSession 验证会话
func (sm *SessionManager) ValidateSession(token string) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, exists := sm.sessions[token]
	if !exists {
		return false
	}

	// 检查是否过期
	if time.Now().After(session.ExpiresAt) {
		return false
	}

	return true
}

// DeleteSession 删除会话
func (sm *SessionManager) DeleteSession(token string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	delete(sm.sessions, token)
}

// cleanupExpiredSessions 清理过期会话
func (sm *SessionManager) cleanupExpiredSessions() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now()
	for token, session := range sm.sessions {
		if now.After(session.ExpiresAt) {
			delete(sm.sessions, token)
		}
	}
}