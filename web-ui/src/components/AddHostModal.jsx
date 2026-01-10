import React, { useState } from 'react';
import './AddHostModal.css';

const AddHostModal = ({ isOpen, onClose, onSubmit }) => {
  const [formData, setFormData] = useState({
    sshConnection: '',
    authType: 'password', // 'password' or 'keyfile'
    password: '',
    keyFile: '',
    d8rctlAddress: 'localhost:50051',
  });

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [result, setResult] = useState(null);

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: value
    }));
  };

  const handleAuthTypeChange = (type) => {
    setFormData(prev => ({
      ...prev,
      authType: type,
      password: '',
      keyFile: ''
    }));
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError(null);
    setResult(null);

    try {
      const resultData = await onSubmit(formData);
      setResult(resultData);
      
      // 如果成功，3秒后自动关闭
      if (resultData.success) {
        setTimeout(() => {
          handleClose();
        }, 3000);
      }
    } catch (err) {
      setError(err.message || '添加主机失败');
    } finally {
      setLoading(false);
    }
  };

  const handleClose = () => {
    setFormData({
      sshConnection: '',
      authType: 'password',
      password: '',
      keyFile: '',
      d8rctlAddress: 'localhost:50051',
    });
    setError(null);
    setResult(null);
    setLoading(false);
    onClose();
  };

  if (!isOpen) return null;

  return (
    <div className="add-host-modal-overlay" onClick={handleClose}>
      <div className="add-host-modal-content" onClick={e => e.stopPropagation()}>
        <div className="add-host-modal-header">
          <h2>添加被控主机</h2>
          <button className="add-host-modal-close" onClick={handleClose}>×</button>
        </div>

        <form onSubmit={handleSubmit} className="add-host-form">
          <div className="form-group">
            <label htmlFor="sshConnection">SSH 连接字符串</label>
            <input
              type="text"
              id="sshConnection"
              name="sshConnection"
              value={formData.sshConnection}
              onChange={handleChange}
              placeholder="user@host:port 或 user@host"
              required
              disabled={loading || (result && result.success)}
            />
            <div className="form-hint">例如: root@192.168.1.100 或 user@server.com:22</div>
          </div>

          <div className="form-group">
            <label>认证方式</label>
            <div className="auth-type-selector">
              <button
                type="button"
                className={`auth-type-btn ${formData.authType === 'password' ? 'active' : ''}`}
                onClick={() => handleAuthTypeChange('password')}
                disabled={loading || (result && result.success)}
              >
                密码
              </button>
              <button
                type="button"
                className={`auth-type-btn ${formData.authType === 'keyfile' ? 'active' : ''}`}
                onClick={() => handleAuthTypeChange('keyfile')}
                disabled={loading || (result && result.success)}
              >
                密钥文件
              </button>
            </div>
          </div>

          {formData.authType === 'password' ? (
            <div className="form-group">
              <label htmlFor="password">SSH 密码</label>
              <input
                type="password"
                id="password"
                name="password"
                value={formData.password}
                onChange={handleChange}
                placeholder="输入 SSH 密码"
                required
                disabled={loading || (result && result.success)}
              />
            </div>
          ) : (
            <div className="form-group">
              <label htmlFor="keyFile">SSH 密钥文件路径</label>
              <input
                type="text"
                id="keyFile"
                name="keyFile"
                value={formData.keyFile}
                onChange={handleChange}
                placeholder="例如: /home/user/.ssh/id_rsa"
                required
                disabled={loading || (result && result.success)}
              />
              <div className="form-hint">服务器上密钥文件的绝对路径</div>
            </div>
          )}

          <div className="form-group">
            <label htmlFor="d8rctlAddress">D8rctl 服务地址</label>
            <input
              type="text"
              id="d8rctlAddress"
              name="d8rctlAddress"
              value={formData.d8rctlAddress}
              onChange={handleChange}
              placeholder="例如: 192.168.1.100:50051"
              required
              disabled={loading || (result && result.success)}
            />
            <div className="form-hint">被控主机可访问的 D8rctl 服务地址</div>
          </div>

          {error && (
            <div className="add-host-error">
              <strong>错误:</strong> {error}
            </div>
          )}

          {result && (
            <div className={`add-host-result ${result.success ? 'success' : 'error'}`}>
              {result.success ? (
                <>
                  <div className="result-icon">✓</div>
                  <div className="result-message">
                    <strong>主机添加成功！</strong>
                    <div className="result-details">
                      <div>节点 ID: {result.node_id}</div>
                      <div>主机名: {result.hostname}</div>
                      <div>操作系统: {result.os} / {result.arch}</div>
                    </div>
                  </div>
                </>
              ) : (
                <>
                  <div className="result-icon">✗</div>
                  <div className="result-message">
                    <strong>添加失败</strong>
                    <div>{result.message}</div>
                    {result.hostname && (
                      <div className="result-details">
                        <div>主机名: {result.hostname}</div>
                        {result.os && <div>操作系统: {result.os} / {result.arch}</div>}
                      </div>
                    )}
                  </div>
                </>
              )}
            </div>
          )}

          <div className="add-host-modal-footer">
            <button
              type="button"
              className="btn btn-secondary"
              onClick={handleClose}
              disabled={loading}
            >
              {result && result.success ? '关闭' : '取消'}
            </button>
            {!(result && result.success) && (
              <button
                type="submit"
                className="btn btn-primary"
                disabled={loading}
              >
                {loading ? '正在添加...' : '添加主机'}
              </button>
            )}
          </div>
        </form>
      </div>
    </div>
  );
};

export default AddHostModal;
