package daemon

import (
	"encoding/json"
	"net/http"
	"time"

	"d8rctl/auth"

	"go.uber.org/zap"
)

// handleLogin 处理登录请求
func (hs *HTTPServer) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 处理登录
	var req struct {
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request"})
		return
	}

	// 验证密码
	if !auth.GetPasswordManager().Verify(req.Password) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid password"})
		return
	}

	// 创建会话
	token, err := auth.GetSessionManager().CreateSession()
	if err != nil {
		zap.L().Sugar().Error("Failed to create session", zap.Error(err))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to create session"})
		return
	}

	// 设置 cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "login successful",
	})

	zap.L().Sugar().Info("User logged in successfully")
}

// handleLogout 处理登出请求
func (hs *HTTPServer) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 获取并删除 session
	cookie, err := r.Cookie("session_token")
	if err == nil {
		auth.GetSessionManager().DeleteSession(cookie.Value)
	}

	// 清除 cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "logout successful",
	})

	zap.L().Sugar().Info("User logged out")
}

// handleIndex 处理根路径，重定向到登录或仪表板
func (hs *HTTPServer) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 检查是否已登录
	cookie, err := r.Cookie("session_token")
	if err != nil || !auth.GetSessionManager().ValidateSession(cookie.Value) {
		// 未登录，重定向到登录页面
		http.Redirect(w, r, "/login.html", http.StatusFound)
		return
	}

	// 已登录，重定向到仪表板
	http.Redirect(w, r, "/dashboard.html", http.StatusFound)
}