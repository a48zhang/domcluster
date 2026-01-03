import React, { useState, useEffect } from 'react';
import Modal from './Modal';
import Toast from './Toast';
import './Dashboard.css';

const Dashboard = ({ onLogout }) => {
  const [status, setStatus] = useState({
    running: false,
    pid: 0,
    uptime: '-',
    nodes: 0,
    message: '',
  });
  const [loading, setLoading] = useState(true);
  const [modal, setModal] = useState({ isOpen: false, action: null, title: '' });
  const [toast, setToast] = useState(null);

  const loadStatus = async () => {
    try {
      const response = await fetch('/api/status');
      if (response.status === 401) {
        onLogout();
        return;
      }
      const data = await response.json();
      setStatus(data);
    } catch (error) {
      console.error('Failed to load status:', error);
      setToast({ message: '加载状态失败', type: 'error' });
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadStatus();
    const interval = setInterval(loadStatus, 5000);
    return () => clearInterval(interval);
  }, []);

  const handleStop = async () => {
    setModal({ isOpen: false, action: null, title: '' });

    try {
      const response = await fetch('/api/stop', { method: 'POST' });
      const data = await response.json();

      if (data.status === 'ok') {
        setToast({ message: '服务已停止', type: 'success' });
        setTimeout(() => loadStatus(), 1000);
      } else {
        setToast({ message: '停止服务失败', type: 'error' });
      }
    } catch (error) {
      setToast({ message: '停止服务失败', type: 'error' });
    }
  };

  const handleRestart = async () => {
    setModal({ isOpen: false, action: null, title: '' });

    try {
      const response = await fetch('/api/restart', { method: 'POST' });
      const data = await response.json();

      if (data.status === 'ok') {
        setToast({ message: '服务正在重启', type: 'warning' });
        setTimeout(() => loadStatus(), 2000);
      } else {
        setToast({ message: '重启服务失败', type: 'error' });
      }
    } catch (error) {
      setToast({ message: '重启服务失败', type: 'error' });
    }
  };

  const openModal = (action, title) => {
    setModal({ isOpen: true, action, title });
  };

  const closeModal = () => {
    setModal({ isOpen: false, action: null, title: '' });
  };

  const handleLogout = async () => {
    try {
      await fetch('/api/logout', { method: 'POST' });
      onLogout();
    } catch (error) {
      console.error('Logout failed:', error);
      setToast({ message: '退出登录失败', type: 'error' });
    }
  };

  return (
    <div className="dashboard">
      <header className="dashboard-header">
        <h1>Domcluster Dashboard</h1>
        <button className="btn btn-logout" onClick={handleLogout}>
          退出登录
        </button>
      </header>

      <main className="dashboard-main">
        <div className="status-card">
          <h2>服务状态</h2>
          {loading ? (
            <div className="loading">加载中...</div>
          ) : (
            <div className="status-info">
              <div className="info-item">
                <div className="info-label">运行状态</div>
                <div className={`info-value ${status.running ? 'status-running' : 'status-stopped'}`}>
                  {status.running ? '运行中' : '已停止'}
                </div>
              </div>
              <div className="info-item">
                <div className="info-label">进程 ID</div>
                <div className="info-value">{status.pid}</div>
              </div>
              <div className="info-item">
                <div className="info-label">运行时间</div>
                <div className="info-value">{status.uptime}</div>
              </div>
              <div className="info-item">
                <div className="info-label">节点数量</div>
                <div className="info-value">{status.nodes}</div>
              </div>
            </div>
          )}
          <div className="actions">
            <button
              className="btn btn-stop"
              onClick={() => openModal('stop', '确认停止服务')}
              disabled={!status.running}
            >
              停止服务
            </button>
            <button
              className="btn btn-restart"
              onClick={() => openModal('restart', '确认重启服务')}
              disabled={!status.running}
            >
              重启服务
            </button>
          </div>
        </div>
      </main>

      <Modal
        isOpen={modal.isOpen}
        onClose={closeModal}
        title={modal.title}
        onConfirm={modal.action === 'stop' ? handleStop : handleRestart}
        confirmText="确认"
      >
        <p>此操作将{modal.action === 'stop' ? '停止' : '重启'} Domcluster 服务，确定要继续吗？</p>
      </Modal>

      {toast && (
        <Toast
          message={toast.message}
          type={toast.type}
          onClose={() => setToast(null)}
        />
      )}
    </div>
  );
};

export default Dashboard;