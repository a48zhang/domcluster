import { useState, useEffect, useCallback } from 'react';
import apiClient from '../api/client';

export const useContainers = (nodeId, showAll = false) => {
  const [containers, setContainers] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  const loadContainers = useCallback(async () => {
    if (!nodeId) return;
    
    setLoading(true);
    try {
      const data = await apiClient.listContainers(nodeId, showAll);
      setContainers(data.containers || []);
      setError(null);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }, [nodeId, showAll]);

  useEffect(() => {
    loadContainers();
    const interval = setInterval(loadContainers, 10000);
    return () => clearInterval(interval);
  }, [loadContainers]);

  const startContainer = async (containerId) => {
    await apiClient.startContainer(nodeId, containerId);
    setTimeout(loadContainers, 1000);
  };

  const stopContainer = async (containerId) => {
    await apiClient.stopContainer(nodeId, containerId);
    setTimeout(loadContainers, 1000);
  };

  const restartContainer = async (containerId) => {
    await apiClient.restartContainer(nodeId, containerId);
    setTimeout(loadContainers, 1000);
  };

  return {
    containers,
    loading,
    error,
    refresh: loadContainers,
    startContainer,
    stopContainer,
    restartContainer,
  };
};
