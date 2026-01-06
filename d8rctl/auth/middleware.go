package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware 认证中间件
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if !GetSessionManager().ValidateSession(cookie.Value) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}

// GinAuthMiddleware Gin 认证中间件
func GinAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("session_token")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		if !GetSessionManager().ValidateSession(token) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "session expired or invalid"})
			c.Abort()
			return
		}

		c.Next()
	}
}