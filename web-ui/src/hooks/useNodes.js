import { useState, useEffect, useCallback } from 'react';
import apiClient from '../api/client';

export const useNodes = (onUnauthorized) => {
  const [nodes, setNodes] = useState({});
  const [nodeStatuses, setNodeStatuses] = useState({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  const loadNodes = useCallback(async () => {
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
          } catch (err) {
            console.error(`Failed to load status for node ${nodeId}:`, err);
          }
        })
      );
      setNodeStatuses(statuses);
      setError(null);
    } catch (err) {
      if (err.message === 'UNAUTHORIZED') {
        onUnauthorized?.();
        return;
      }
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }, [onUnauthorized]);

  useEffect(() => {
    loadNodes();
    const interval = setInterval(loadNodes, 10000);
    return () => clearInterval(interval);
  }, [loadNodes]);

  return { nodes, nodeStatuses, loading, error, refresh: loadNodes };
};
