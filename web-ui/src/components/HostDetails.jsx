import React, { useState, useEffect, useRef } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';
import { Terminal } from '@xterm/xterm';
import { FitAddon } from '@xterm/addon-fit';
import '@xterm/xterm/css/xterm.css';
import apiClient from '../api/client';
import Toast from './Toast';
import './HostDetails.css';

const HostDetails = ({ onLogout }) => {
  const { nodeId } = useParams();
  const navigate = useNavigate();
  const terminalRef = useRef(null);
  const xtermRef = useRef(null);
  const fitAddonRef = useRef(null);

  const [nodeInfo, setNodeInfo] = useState(null);
  const [nodeStatus, setNodeStatus] = useState(null);
  const [containers, setContainers] = useState([]);
  const [cpuData, setCpuData] = useState([]);
  const [memoryData, setMemoryData] = useState([]);
  const [networkData, setNetworkData] = useState([]);
  const [loading, setLoading] = useState(true);
  const [toast, setToast] = useState(null);
  const [showAllContainers, setShowAllContainers] = useState(false);
  const [terminalVisible, setTerminalVisible] = useState(false);

  // Helper function to format bytes
  const formatBytes = (bytes) => {
    if (!bytes || bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  // Helper function to format time
  const formatTime = (timestamp) => {
    const date = new Date(timestamp);
    return date.toLocaleTimeString();
  };

  // Load node info and status
  const loadNodeData = async () => {
    try {
      // Get node basic info
      const nodes = await apiClient.getNodes();
      const info = nodes[nodeId];
      if (!info) {
        setToast({ message: '节点不存在', type: 'error' });
        navigate('/');
        return;
      }
      setNodeInfo({ ...info, id: nodeId });

      // Get node status
      const status = await apiClient.getNodeStatus(nodeId);
      setNodeStatus(status);

      // Update charts data
      const now = new Date().getTime();
      const timestamp = now;

      setCpuData(prev => {
        const newData = [...prev, {
          time: formatTime(timestamp),
          usage: status.system_resources?.cpu?.usage_percent || 0,
        }];
        return newData.slice(-20); // Keep last 20 data points
      });

      setMemoryData(prev => {
        const newData = [...prev, {
          time: formatTime(timestamp),
          usage: status.system_resources?.memory?.usage_percent || 0,
        }];
        return newData.slice(-20);
      });

      setNetworkData(prev => {
        const newData = [...prev, {
          time: formatTime(timestamp),
          rx: (status.system_resources?.network?.rx_bytes || 0) / 1024 / 1024, // Convert to MB
          tx: (status.system_resources?.network?.tx_bytes || 0) / 1024 / 1024,
        }];
        return newData.slice(-20);
      });

    } catch (error) {
      if (error.message === 'UNAUTHORIZED') {
        onLogout();
        return;
      }
      console.error('Failed to load node data:', error);
      setToast({ message: '加载节点数据失败', type: 'error' });
    } finally {
      setLoading(false);
    }
  };

  // Load containers
  const loadContainers = async () => {
    try {
      const data = await apiClient.listContainers(nodeId, showAllContainers);
      setContainers(data.containers || []);
    } catch (error) {
      console.error('Failed to load containers:', error);
      setToast({ message: '加载容器列表失败', type: 'error' });
    }
  };

  // Container control functions
  const handleStartContainer = async (containerId) => {
    try {
      await apiClient.startContainer(nodeId, containerId);
      setToast({ message: '容器启动成功', type: 'success' });
      setTimeout(loadContainers, 1000);
    } catch (error) {
      setToast({ message: '启动容器失败', type: 'error' });
    }
  };

  const handleStopContainer = async (containerId) => {
    try {
      await apiClient.stopContainer(nodeId, containerId);
      setToast({ message: '容器停止成功', type: 'success' });
      setTimeout(loadContainers, 1000);
    } catch (error) {
      setToast({ message: '停止容器失败', type: 'error' });
    }
  };

  const handleRestartContainer = async (containerId) => {
    try {
      await apiClient.restartContainer(nodeId, containerId);
      setToast({ message: '容器重启成功', type: 'success' });
      setTimeout(loadContainers, 1000);
    } catch (error) {
      setToast({ message: '重启容器失败', type: 'error' });
    }
  };

  // Initialize terminal
  const initTerminal = () => {
    if (!terminalRef.current || xtermRef.current) return;

    const term = new Terminal({
      cursorBlink: true,
      fontSize: 14,
      fontFamily: 'Consolas, Monaco, "Courier New", monospace',
      theme: {
        background: '#1e1e1e',
        foreground: '#f8f8f8',
      },
    });

    const fitAddon = new FitAddon();
    term.loadAddon(fitAddon);
    term.open(terminalRef.current);
    fitAddon.fit();

    xtermRef.current = term;
    fitAddonRef.current = fitAddon;

    // Connect to WebSocket
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/api/terminal/ws?node_id=${nodeId}`;
    
    const ws = new WebSocket(wsUrl);
    const wsRef = { current: ws };

    ws.onopen = () => {
      term.writeln('Connected to ' + (nodeInfo?.name || nodeId));
      term.writeln('Type commands and press Enter to execute');
      term.writeln('');
      term.write('$ ');
    };

    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data);
        if (msg.type === 'output' && msg.data) {
          term.write(msg.data);
        }
      } catch (error) {
        console.error('Failed to parse WebSocket message:', error);
      }
    };

    ws.onerror = (error) => {
      term.writeln('\r\nWebSocket error occurred');
      console.error('WebSocket error:', error);
    };

    ws.onclose = () => {
      term.writeln('\r\nConnection closed');
    };

    let currentLine = '';

    // Handle terminal input
    term.onData((data) => {
      const code = data.charCodeAt(0);

      if (code === 13) { // Enter key
        term.write('\r\n');
        if (currentLine.trim() && ws.readyState === WebSocket.OPEN) {
          // Send command to backend
          ws.send(JSON.stringify({
            type: 'input',
            data: currentLine.trim()
          }));
        }
        currentLine = '';
        term.write('$ ');
      } else if (code === 127) { // Backspace
        if (currentLine.length > 0) {
          currentLine = currentLine.slice(0, -1);
          term.write('\b \b');
        }
      } else if (code >= 32) { // Printable characters
        currentLine += data;
        term.write(data);
      }
    });

    // Resize on window resize
    const handleResize = () => {
      if (fitAddonRef.current && terminalVisible) {
        fitAddonRef.current.fit();
      }
    };
    window.addEventListener('resize', handleResize);

    return () => {
      window.removeEventListener('resize', handleResize);
      if (wsRef.current) {
        wsRef.current.close();
      }
      term.dispose();
    };
  };

  useEffect(() => {
    loadNodeData();
    loadContainers();
    
    const statusInterval = setInterval(loadNodeData, 5000);
    const containersInterval = setInterval(loadContainers, 10000);
    
    return () => {
      clearInterval(statusInterval);
      clearInterval(containersInterval);
    };
  }, [nodeId, showAllContainers]);

  useEffect(() => {
    if (terminalVisible) {
      initTerminal();
    }
  }, [terminalVisible, nodeInfo]);

  if (loading) {
    return (
      <div className="host-details">
        <div className="loading">加载中...</div>
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
          <span className={`status-badge ${nodeStatus?.online ? 'online' : 'offline'}`}>
            {nodeStatus?.online ? '在线' : '离线'}
          </span>
        </div>
        <button className="btn btn-logout" onClick={onLogout}>
          退出登录
        </button>
      </header>

      <main className="host-details-main">
        {/* Node Information */}
        <div className="detail-card">
          <h2>节点信息</h2>
          <div className="detail-grid">
            <div className="detail-item">
              <span className="detail-label">节点 ID</span>
              <span className="detail-value">{nodeId}</span>
            </div>
            <div className="detail-item">
              <span className="detail-label">节点名称</span>
              <span className="detail-value">{nodeInfo?.name}</span>
            </div>
            <div className="detail-item">
              <span className="detail-label">角色</span>
              <span className="detail-value">{nodeInfo?.role}</span>
            </div>
            <div className="detail-item">
              <span className="detail-label">版本</span>
              <span className="detail-value">{nodeInfo?.version}</span>
            </div>
          </div>
        </div>

        {/* Resource Information */}
        {nodeStatus?.system_resources && (
          <div className="detail-card">
            <h2>系统资源</h2>
            <div className="resource-grid">
              <div className="resource-item">
                <h3>CPU</h3>
                <div className="resource-details">
                  <p>核心数: {nodeStatus.system_resources.cpu?.core_count || 'N/A'}</p>
                  <p>使用率: {nodeStatus.system_resources.cpu?.usage_percent?.toFixed(2) || 0}%</p>
                </div>
              </div>
              <div className="resource-item">
                <h3>内存</h3>
                <div className="resource-details">
                  <p>总量: {formatBytes(nodeStatus.system_resources.memory?.total_bytes)}</p>
                  <p>已使用: {formatBytes(nodeStatus.system_resources.memory?.used_bytes)}</p>
                  <p>可用: {formatBytes(nodeStatus.system_resources.memory?.available_bytes)}</p>
                  <p>使用率: {nodeStatus.system_resources.memory?.usage_percent?.toFixed(2) || 0}%</p>
                </div>
              </div>
              <div className="resource-item">
                <h3>网络</h3>
                <div className="resource-details">
                  <p>接收: {formatBytes(nodeStatus.system_resources.network?.rx_bytes)}</p>
                  <p>发送: {formatBytes(nodeStatus.system_resources.network?.tx_bytes)}</p>
                </div>
              </div>
            </div>
          </div>
        )}

        {/* CPU Chart */}
        <div className="detail-card chart-card">
          <h2>CPU 使用率 (%)</h2>
          <ResponsiveContainer width="100%" height={250}>
            <LineChart data={cpuData}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="time" />
              <YAxis domain={[0, 100]} />
              <Tooltip />
              <Legend />
              <Line type="monotone" dataKey="usage" stroke="#8884d8" name="CPU使用率" strokeWidth={2} dot={false} />
            </LineChart>
          </ResponsiveContainer>
        </div>

        {/* Memory Chart */}
        <div className="detail-card chart-card">
          <h2>内存使用率 (%)</h2>
          <ResponsiveContainer width="100%" height={250}>
            <LineChart data={memoryData}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="time" />
              <YAxis domain={[0, 100]} />
              <Tooltip />
              <Legend />
              <Line type="monotone" dataKey="usage" stroke="#82ca9d" name="内存使用率" strokeWidth={2} dot={false} />
            </LineChart>
          </ResponsiveContainer>
        </div>

        {/* Network Chart */}
        <div className="detail-card chart-card">
          <h2>网络流量 (MB)</h2>
          <ResponsiveContainer width="100%" height={250}>
            <LineChart data={networkData}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="time" />
              <YAxis />
              <Tooltip />
              <Legend />
              <Line type="monotone" dataKey="rx" stroke="#8884d8" name="接收" strokeWidth={2} dot={false} />
              <Line type="monotone" dataKey="tx" stroke="#82ca9d" name="发送" strokeWidth={2} dot={false} />
            </LineChart>
          </ResponsiveContainer>
        </div>

        {/* Docker Containers */}
        <div className="detail-card">
          <div className="card-header">
            <h2>Docker 容器</h2>
            <label className="checkbox-label">
              <input
                type="checkbox"
                checked={showAllContainers}
                onChange={(e) => setShowAllContainers(e.target.checked)}
              />
              显示所有容器
            </label>
          </div>
          {containers.length === 0 ? (
            <div className="empty-state">
              <p>没有找到容器</p>
            </div>
          ) : (
            <div className="containers-table">
              <table>
                <thead>
                  <tr>
                    <th>容器 ID</th>
                    <th>名称</th>
                    <th>镜像</th>
                    <th>状态</th>
                    <th>操作</th>
                  </tr>
                </thead>
                <tbody>
                  {containers.map((container) => (
                    <tr key={container.id}>
                      <td>{container.id?.substring(0, 12)}</td>
                      <td>{container.names?.[0] || 'N/A'}</td>
                      <td>{container.image || 'N/A'}</td>
                      <td>
                        <span className={`container-status ${container.state?.toLowerCase()}`}>
                          {container.state || 'unknown'}
                        </span>
                      </td>
                      <td>
                        <div className="container-actions">
                          {container.state === 'running' ? (
                            <>
                              <button
                                className="btn-small btn-stop"
                                onClick={() => handleStopContainer(container.id)}
                              >
                                停止
                              </button>
                              <button
                                className="btn-small btn-restart"
                                onClick={() => handleRestartContainer(container.id)}
                              >
                                重启
                              </button>
                            </>
                          ) : (
                            <button
                              className="btn-small btn-start"
                              onClick={() => handleStartContainer(container.id)}
                            >
                              启动
                            </button>
                          )}
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>

        {/* Terminal */}
        <div className="detail-card">
          <div className="card-header">
            <h2>Web Terminal</h2>
            <button
              className="btn btn-toggle"
              onClick={() => setTerminalVisible(!terminalVisible)}
            >
              {terminalVisible ? '隐藏' : '显示'} Terminal
            </button>
          </div>
          {terminalVisible && (
            <div className="terminal-container">
              <div ref={terminalRef} className="terminal"></div>
            </div>
          )}
        </div>
      </main>

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

export default HostDetails;
