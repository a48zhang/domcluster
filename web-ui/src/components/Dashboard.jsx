import React, { useState, useEffect } from 'react';
import apiClient from '../api/client';
import Modal from './Modal';
import Toast from './Toast';
import './Dashboard.css';

const Dashboard = ({ onLogout }) => {
  // 辅助函数：根据使用率获取颜色
  const getResourceColor = (usage) => {
    if (usage < 50) return '#28a745';
    if (usage < 80) return '#ffc107';
    return '#dc3545';
  };

  // 辅助函数：格式化最后更新时间
  const formatLastUpdate = (timestamp) => {
    const now = new Date();
    const update = new Date(timestamp);
    const diff = Math.floor((now - update) / 1000);

    if (diff < 60) return `${diff}秒前`;
    if (diff < 3600) return `${Math.floor(diff / 60)}分钟前`;
    if (diff < 86400) return `${Math.floor(diff / 3600)}小时前`;
    return `${Math.floor(diff / 86400)}天前`;
  };
  const [status, setStatus] = useState({
    running: false,
    pid: 0,
    uptime: '-',
    nodes: 0,
    message: '',
  });
  const [nodes, setNodes] = useState({});
  const [nodeStatuses, setNodeStatuses] = useState({});
  const [loading, setLoading] = useState(true);
  const [nodesLoading, setNodesLoading] = useState(true);
  const [modal, setModal] = useState({ isOpen: false, action: null, title: '' });
  const [toast, setToast] = useState(null);

  const loadStatus = async () => {
    try {
      const data = await apiClient.getStatus();
      setStatus(data);
    } catch (error) {
      if (error.message === 'UNAUTHORIZED') {
        onLogout();
        return;
      }
      console.error('Failed to load status:', error);
      setToast({ message: '加载状态失败', type: 'error' });
    } finally {
      setLoading(false);
    }
  };

  const loadNodes = async () => {
    try {
      const data = await apiClient.getNodes();
      setNodes(data);

      // 加载每个节点的状态
      const nodeIds = Object.keys(data);
      const statuses = {};
      await Promise.all(
        nodeIds.map(async (nodeId) => {
          try {
            const statusData = await apiClient.getNodeStatus(nodeId);
            statuses[nodeId] = statusData;
          } catch (error) {
            console.error(`Failed to load status for node ${nodeId}:`, error);
          }
        })
      );
      setNodeStatuses(statuses);
    } catch (error) {
      if (error.message === 'UNAUTHORIZED') {
        onLogout();
        return;
      }
      console.error('Failed to load nodes:', error);
      setToast({ message: '加载节点列表失败', type: 'error' });
    } finally {
      setNodesLoading(false);
    }
  };

  useEffect(() => {
    loadStatus();
    loadNodes();
    const statusInterval = setInterval(loadStatus, 5000);
    const nodesInterval = setInterval(loadNodes, 10000);
    return () => {
      clearInterval(statusInterval);
      clearInterval(nodesInterval);
    };
  }, []);

  const handleStop = async () => {
    setModal({ isOpen: false, action: null, title: '' });

    try {
      const data = await apiClient.stop();

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
      const data = await apiClient.restart();

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
      await apiClient.logout();
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

        <div className="nodes-card">
          <h2>已连接的节点</h2>
          {nodesLoading ? (
            <div className="loading">加载中...</div>
          ) : Object.keys(nodes).length === 0 ? (
            <div className="empty-state">
              <p>暂无连接的节点</p>
            </div>
          ) : (
            <div className="nodes-grid">
              {Object.entries(nodes).map(([nodeId, nodeInfo]) => {
                const nodeStatus = nodeStatuses[nodeId];
                return (
                  <div key={nodeId} className="node-card">
                    <div className="node-header">
                      <h3 className="node-name">{nodeInfo.name}</h3>
                      <span className="node-id">{nodeId}</span>
                    </div>
                    <div className="node-info">
                      <div className="node-info-item">
                        <span className="node-info-label">角色</span>
                        <span className="node-info-value">{nodeInfo.role || 'worker'}</span>
                      </div>
                      <div className="node-info-item">
                        <span className="node-info-label">版本</span>
                        <span className="node-info-value">{nodeInfo.version || 'unknown'}</span>
                      </div>
                    </div>
                    {nodeStatus && nodeStatus.system_resources && (
                      <div className="node-resources">
                        <div className="resource-item">
                          <span className="resource-label">CPU</span>
                          <div className="resource-bar">
                            <div
                              className="resource-fill"
                              style={{
                                width: `${nodeStatus.system_resources.cpu?.usage_percent || 0}%`,
                                backgroundColor: getResourceColor(nodeStatus.system_resources.cpu?.usage_percent || 0)
                              }}
                            ></div>
                          </div>
                          <span className="resource-value">
                            {nodeStatus.system_resources.cpu?.usage_percent?.toFixed(1) || 0}%
                          </span>
                        </div>
                        <div className="resource-item">
                          <span className="resource-label">内存</span>
                          <div className="resource-bar">
                            <div
                              className="resource-fill"
                              style={{
                                width: `${nodeStatus.system_resources.memory?.usage_percent || 0}%`,
                                backgroundColor: getResourceColor(nodeStatus.system_resources.memory?.usage_percent || 0)
                              }}
                            ></div>
                          </div>
                          <span className="resource-value">
                            {nodeStatus.system_resources.memory?.usage_percent?.toFixed(1) || 0}%
                          </span>
                        </div>
                        {nodeStatus.docker && (
                          <div className="docker-info">
                            <span className="docker-label">容器</span>
                            <span className="docker-value">
                              {nodeStatus.docker.running_count}/{nodeStatus.docker.total_count}
                            </span>
                          </div>
                        )}
                      </div>
                    )}
                    <div className="node-status">
                      <span className={`status-indicator ${nodeStatus?.online ? 'online' : 'offline'}`}></span>
                      <span className="status-text">
                        {nodeStatus?.online ? '在线' : '离线'}
                      </span>
                      {nodeStatus?.last_update && (
                        <span className="last-update">
                          {formatLastUpdate(nodeStatus.last_update)}
                        </span>
                      )}
                    </div>
                  </div>
                );
              })}
            </div>
          )}
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