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
	sessions      map[string]*Session
	mu            sync.RWMutex
	cleanupOnce   sync.Once
	cleanupDone   chan struct{}
}

var (
	sessionInstance *SessionManager
	sessionOnce     sync.Once
)

// GetSessionManager 获取会话管理器单例
func GetSessionManager() *SessionManager {
	sessionOnce.Do(func() {
		sessionInstance = &SessionManager{
			sessions:    make(map[string]*Session),
			cleanupDone: make(chan struct{}),
		}
	})
	return sessionInstance
}

// CreateSession 创建新会话
func (sm *SessionManager) CreateSession() (string, error) {
	sm.cleanupOnce.Do(func() {
		go sm.runCleanupRoutine()
	})

	sm.mu.Lock()
	defer sm.mu.Unlock()

	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := hex.EncodeToString(b)

	now := time.Now()
	session := &Session{
		Token:     token,
		CreatedAt: now,
		ExpiresAt: now.Add(24 * time.Hour),
	}

	sm.sessions[token] = session

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

// Shutdown 关闭会话管理器，停止清理 goroutine
func (sm *SessionManager) Shutdown() {
	if sm.cleanupDone != nil {
		close(sm.cleanupDone)
	}
}

// runCleanupRoutine 定期清理过期会话
func (sm *SessionManager) runCleanupRoutine() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sm.cleanupExpiredSessions()
		case <-sm.cleanupDone:
			return
		}
	}
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