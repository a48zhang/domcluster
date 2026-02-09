import React, { useState } from 'react';
import apiClient from '../api/client';
import { useStatus, useNodes, useToast } from '../hooks';
import { StatusCard, NodesList } from './dashboard';
import Modal from './Modal';
import Toast from './Toast';
import './Dashboard.css';

const Dashboard = ({ onLogout }) => {
  const { status, loading: statusLoading } = useStatus(onLogout);
  const { nodes, nodeStatuses, loading: nodesLoading } = useNodes(onLogout);
  const { toast, showSuccess, showError, hideToast } = useToast();
  
  const [modal, setModal] = useState({ isOpen: false, action: null, title: '' });

  const handleStop = async () => {
    setModal({ isOpen: false, action: null, title: '' });
    try {
      const data = await apiClient.stop();
      if (data.status === 'ok') {
        showSuccess('服务已停止');
      } else {
        showError('停止服务失败');
      }
    } catch (error) {
      showError('停止服务失败');
    }
  };

  const handleRestart = async () => {
    setModal({ isOpen: false, action: null, title: '' });
    try {
      const data = await apiClient.restart();
      if (data.status === 'ok') {
        showWarning('服务正在重启');
      } else {
        showError('重启服务失败');
      }
    } catch (error) {
      showError('重启服务失败');
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
      showError('退出登录失败');
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
        <StatusCard 
          status={status} 
          loading={statusLoading}
          onStop={() => openModal('stop', '确认停止服务')}
          onRestart={() => openModal('restart', '确认重启服务')}
        />
        <NodesList 
          nodes={nodes}
          nodeStatuses={nodeStatuses}
          loading={nodesLoading}
        />
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
          onClose={hideToast}
        />
      )}
    </div>
  );
};

export default Dashboard;
