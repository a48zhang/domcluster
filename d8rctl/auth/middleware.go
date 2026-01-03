package auth

import (
	"encoding/json"
	"net/http"
)

// AuthMiddleware 认证中间件
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 检查 session cookie
		cookie, err := r.Cookie("session_token")
		if err != nil {
			// 未登录，返回 401
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
			return
		}

		// 验证会话
		if !GetSessionManager().ValidateSession(cookie.Value) {
			// 会话无效或过期
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "session expired or invalid"})
			return
		}

		// 会话有效，继续处理请求
		next(w, r)
	}
}