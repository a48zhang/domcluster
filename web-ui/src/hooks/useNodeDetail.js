import { useState, useEffect, useCallback } from 'react';
import apiClient from '../api/client';
import { formatTime } from '../utils/format';

export const useNodeDetail = (nodeId, onUnauthorized) => {
  const [nodeInfo, setNodeInfo] = useState(null);
  const [nodeStatus, setNodeStatus] = useState(null);
  const [cpuData, setCpuData] = useState([]);
  const [memoryData, setMemoryData] = useState([]);
  const [networkData, setNetworkData] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  const loadNodeData = useCallback(async () => {
    try {
      // Get node basic info
      const nodes = await apiClient.getNodes();
      const info = nodes[nodeId];
      if (!info) {
        setError('节点不存在');
        return null;
      }
      setNodeInfo({ ...info, id: nodeId });

      // Get node status
      const status = await apiClient.getNodeStatus(nodeId);
      setNodeStatus(status);

      // Update charts data
      const timestamp = new Date().getTime();

      setCpuData(prev => {
        const newData = [...prev, {
          time: formatTime(timestamp),
          usage: status.system_resources?.cpu?.usage_percent || 0,
        }];
        return newData.slice(-20);
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
          rx: (status.system_resources?.network?.rx_bytes || 0) / 1024 / 1024,
          tx: (status.system_resources?.network?.tx_bytes || 0) / 1024 / 1024,
        }];
        return newData.slice(-20);
      });

      setError(null);
      return status;
    } catch (err) {
      if (err.message === 'UNAUTHORIZED') {
        onUnauthorized?.();
        return null;
      }
      setError(err.message);
      return null;
    } finally {
      setLoading(false);
    }
  }, [nodeId, onUnauthorized]);

  useEffect(() => {
    loadNodeData();
    const interval = setInterval(loadNodeData, 5000);
    return () => clearInterval(interval);
  }, [loadNodeData]);

  return {
    nodeInfo,
    nodeStatus,
    cpuData,
    memoryData,
    networkData,
    loading,
    error,
    refresh: loadNodeData,
  };
};
