import React, { useState, useEffect } from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import apiClient from './api/client';
import Login from './components/Login';
import Dashboard from './components/Dashboard';
import HostDetails from './components/HostDetails';

function App() {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [authChecked, setAuthChecked] = useState(false);

  useEffect(() => {
    checkAuth();
  }, []);

  const checkAuth = async () => {
    try {
      await apiClient.getStatus();
      setIsAuthenticated(true);
    } catch (error) {
      console.error('Auth check failed:', error);
    } finally {
      setAuthChecked(true);
    }
  };

  const handleLogin = () => {
    setIsAuthenticated(true);
  };

  const handleLogout = () => {
    setIsAuthenticated(false);
  };

  if (!authChecked) {
    return <div className="app">Loading...</div>;
  }

  return (
    <BrowserRouter>
      <div className="app">
        {isAuthenticated ? (
          <Routes>
            <Route path="/" element={<Dashboard onLogout={handleLogout} />} />
            <Route path="/host/:nodeId" element={<HostDetails onLogout={handleLogout} />} />
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        ) : (
          <Routes>
            <Route path="*" element={<Login onLogin={handleLogin} />} />
          </Routes>
        )}
      </div>
    </BrowserRouter>
  );
}

export default App;