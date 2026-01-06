package daemon

import (
	"net/http"

	"d8rctl/auth"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// handleLogin 处理登录请求
func (hs *HTTPServer) handleLogin(c *gin.Context) {
	var req struct {
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if !auth.GetPasswordManager().Verify(req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid password"})
		return
	}

	token, err := auth.GetSessionManager().CreateSession()
	if err != nil {
		zap.L().Sugar().Error("Failed to create session", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session"})
		return
	}

	c.SetCookie(
		"session_token",
		token,
		24*3600,
		"/",
		"",
		true,
		true,
	)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "login successful",
	})

	zap.L().Sugar().Info("User logged in successfully")
}

// handleLogout 处理登出请求
func (hs *HTTPServer) handleLogout(c *gin.Context) {
	token, err := c.Cookie("session_token")
	if err == nil {
		auth.GetSessionManager().DeleteSession(token)
	}

	c.SetCookie(
		"session_token",
		"",
		-1,
		"/",
		"",
		true,
		true,
	)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "logout successful",
	})

	zap.L().Sugar().Info("User logged out")
}