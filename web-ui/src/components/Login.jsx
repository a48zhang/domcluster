import React, { useState } from 'react';
import apiClient from '../api/client';
import './Login.css';

const Login = ({ onLogin }) => {
  const [host, setHost] = useState(apiClient.getHost());
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      // 设置API主机地址
      apiClient.setHost(host);

      const data = await apiClient.login(password);

      if (data.success) {
        onLogin(host);
      } else {
        setError(data.error || '登录失败，请检查密码');
        setPassword('');
      }
    } catch (err) {
      setError('网络错误，请检查主机地址和端口是否正确');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="login-container">
      <div className="login-card">
        <div className="login-header">
          <h1>Domcluster</h1>
          <p>请输入密码访问管理界面</p>
        </div>
        {error && <div className="error-message">{error}</div>}
        <form className="login-form" onSubmit={handleSubmit}>
          <div className="form-group">
            <label htmlFor="host">主机</label>
            <input
              type="text"
              id="host"
              value={host}
              onChange={(e) => setHost(e.target.value)}
              required
              autoFocus
              placeholder="例如: localhost:50051"
            />
          </div>
          <div className="form-group">
            <label htmlFor="password">密码</label>
            <input
              type="password"
              id="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              placeholder="请输入密码"
            />
          </div>
          <button type="submit" className="login-button" disabled={loading}>
            {loading ? '登录中...' : '登录'}
          </button>
        </form>
      </div>
    </div>
  );
};

export default Login;