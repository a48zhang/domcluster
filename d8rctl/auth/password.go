package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"sync"

	"go.uber.org/zap"
)

const (
	passwordFile = ".d8rctl_password"
	passwordLen  = 16
)

var (
	instance *PasswordManager
	once     sync.Once
)

// PasswordManager 密码管理器
type PasswordManager struct {
	passwordHash string
	mu           sync.RWMutex
}

// GetPasswordManager 获取密码管理器单例
func GetPasswordManager() *PasswordManager {
	once.Do(func() {
		instance = &PasswordManager{}
	})
	return instance
}

// Init 初始化密码管理器
func (pm *PasswordManager) Init() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.passwordFileExists() {
		hash, err := pm.readPasswordHash()
		if err != nil {
			return fmt.Errorf("failed to read password hash: %w", err)
		}
		pm.passwordHash = hash
		zap.L().Sugar().Info("Password loaded from file")
	} else {
		password, err := pm.generatePassword()
		if err != nil {
			return fmt.Errorf("failed to generate password: %w", err)
		}

		hash := pm.hashPassword(password)
		pm.passwordHash = hash

		if err := pm.savePasswordHash(hash); err != nil {
			return fmt.Errorf("failed to save password hash: %w", err)
		}

		zap.L().Sugar().Warnf("========================================")
		zap.L().Sugar().Warnf("INITIAL PASSWORD: %s", password)
		zap.L().Sugar().Warnf("Please save this password to access the web interface")
		zap.L().Sugar().Warnf("Use 'd8rctl password' command to view this password later")
		zap.L().Sugar().Warnf("========================================")
	}

	return nil
}

// Verify 验证密码
func (pm *PasswordManager) Verify(password string) bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	// 每次验证时从文件读取最新的密码哈希
	currentHash, err := pm.readPasswordHash()
	if err != nil {
		zap.L().Sugar().Error("Failed to read password hash", zap.Error(err))
		return false
	}

	hash := pm.hashPassword(password)
	return hash == currentHash
}

// GetPassword 获取密码（仅用于 CLI 命令显示）
func (pm *PasswordManager) GetPassword() (string, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	return "", fmt.Errorf("password cannot be retrieved from hash. Please check the logs for initial password or reset it")
}

// ResetPassword 重置密码
func (pm *PasswordManager) ResetPassword() (string, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	password, err := pm.generatePassword()
	if err != nil {
		return "", fmt.Errorf("failed to generate password: %w", err)
	}

	hash := pm.hashPassword(password)
	pm.passwordHash = hash

	if err := pm.savePasswordHash(hash); err != nil {
		return "", fmt.Errorf("failed to save password hash: %w", err)
	}

	zap.L().Sugar().Warnf("Password reset. New password: %s", password)

	return password, nil
}

// passwordFileExists 检查密码文件是否存在
func (pm *PasswordManager) passwordFileExists() bool {
	_, err := os.Stat(pm.getPasswordFilePath())
	return err == nil
}

// readPasswordHash 读取密码哈希
func (pm *PasswordManager) readPasswordHash() (string, error) {
	data, err := os.ReadFile(pm.getPasswordFilePath())
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// savePasswordHash 保存密码哈希
func (pm *PasswordManager) savePasswordHash(hash string) error {
	return os.WriteFile(pm.getPasswordFilePath(), []byte(hash), 0600)
}

// getPasswordFilePath 获取密码文件路径
func (pm *PasswordManager) getPasswordFilePath() string {
	return "/run/d8rctl/password"
}

// generatePassword 生成随机密码
func (pm *PasswordManager) generatePassword() (string, error) {
	b := make([]byte, passwordLen)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b)[:16], nil
}

// hashPassword 计算密码哈希
func (pm *PasswordManager) hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}