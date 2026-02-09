import React from 'react';
import { Loading } from '../../common';

const StatusCard = ({ status, loading, onStop, onRestart }) => {
  if (loading) {
    return (
      <div className="status-card">
        <h2>服务状态</h2>
        <Loading />
      </div>
    );
  }

  return (
    <div className="status-card">
      <h2>服务状态</h2>
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
      <div className="actions">
        <button
          className="btn btn-stop"
          onClick={onStop}
          disabled={!status.running}
        >
          停止服务
        </button>
        <button
          className="btn btn-restart"
          onClick={onRestart}
          disabled={!status.running}
        >
          重启服务
        </button>
      </div>
    </div>
  );
};

export default StatusCard;
