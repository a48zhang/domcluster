import React, { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import apiClient from '../api/client';
import { useNodeDetail, useContainers, useToast } from '../hooks';
import { Loading, StatusBadge } from '../common';
import { 
  NodeInfoCard, 
  ResourceCard, 
  CpuChart, 
  MemoryChart, 
  NetworkChart,
  ContainerTable,
  Terminal 
} from './host-details';
import Toast from './Toast';
import './HostDetails.css';

const HostDetails = ({ onLogout }) => {
  const { nodeId } = useParams();
  const navigate = useNavigate();
  
  const [showAllContainers, setShowAllContainers] = useState(false);
  const [terminalVisible, setTerminalVisible] = useState(false);
  
  const { 
    nodeInfo, 
    nodeStatus, 
    cpuData, 
    memoryData, 
    networkData, 
    loading 
  } = useNodeDetail(nodeId, onLogout);
  
  const { 
    containers, 
    startContainer, 
    stopContainer, 
    restartContainer 
  } = useContainers(nodeId, showAllContainers);
  
  const { toast, showSuccess, showError, hideToast } = useToast();

  const handleStart = async (containerId) => {
    try {
      await startContainer(containerId);
      showSuccess('容器启动成功');
    } catch (error) {
      showError('启动容器失败');
    }
  };

  const handleStop = async (containerId) => {
    try {
      await stopContainer(containerId);
      showSuccess('容器停止成功');
    } catch (error) {
      showError('停止容器失败');
    }
  };

  const handleRestart = async (containerId) => {
    try {
      await restartContainer(containerId);
      showSuccess('容器重启成功');
    } catch (error) {
      showError('重启容器失败');
    }
  };

  const handleLogout = async () => {
    try {
      await apiClient.logout();
      onLogout();
    } catch (error) {
      showError('退出登录失败');
    }
  };

  if (loading) {
    return (
      <div className="host-details">
        <Loading />
      </div>
    );
  }

  return (
    <div className="host-details">
      <header className="host-details-header">
        <div className="header-left">
          <button className="btn btn-back" onClick={() => navigate('/')}>
            ← 返回
          </button>
          <h1>{nodeInfo?.name || nodeId}</h1>
          <StatusBadge online={nodeStatus?.online} />
        </div>
        <button className="btn btn-logout" onClick={handleLogout}>
          退出登录
        </button>
      </header>

      <main className="host-details-main">
        <NodeInfoCard nodeId={nodeId} nodeInfo={nodeInfo} />
        <ResourceCard nodeStatus={nodeStatus} />
        <CpuChart data={cpuData} />
        <MemoryChart data={memoryData} />
        <NetworkChart data={networkData} />
        <ContainerTable
          containers={containers}
          showAll={showAllContainers}
          onShowAllChange={setShowAllContainers}
          onStart={handleStart}
          onStop={handleStop}
          onRestart={handleRestart}
        />
        <Terminal
          nodeId={nodeId}
          nodeInfo={nodeInfo}
          visible={terminalVisible}
          onToggle={() => setTerminalVisible(!terminalVisible)}
        />
      </main>

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

export default HostDetails;