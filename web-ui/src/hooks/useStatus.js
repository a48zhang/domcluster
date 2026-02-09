import { useState, useEffect, useCallback } from 'react';
import apiClient from '../api/client';

export const useStatus = (onUnauthorized) => {
  const [status, setStatus] = useState({
    running: false,
    pid: 0,
    uptime: '-',
    nodes: 0,
    message: '',
  });
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  const loadStatus = useCallback(async () => {
    try {
      const data = await apiClient.getStatus();
      setStatus(data);
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
    loadStatus();
    const interval = setInterval(loadStatus, 5000);
    return () => clearInterval(interval);
  }, [loadStatus]);

  return { status, loading, error, refresh: loadStatus };
};
