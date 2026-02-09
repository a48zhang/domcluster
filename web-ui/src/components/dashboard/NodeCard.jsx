import React from 'react';
import { useNavigate } from 'react-router-dom';
import { formatLastUpdate } from '../../utils/format';
import { ResourceBar } from '../../common';

const NodeCard = ({ nodeId, nodeInfo, nodeStatus }) => {
  const navigate = useNavigate();

  return (
    <div 
      className="node-card clickable"
      onClick={() => navigate(`/host/${nodeId}`)}
    >
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
      {nodeStatus?.system_resources && (
        <div className="node-resources">
          <ResourceBar 
            label="CPU" 
            value={nodeStatus.system_resources.cpu?.usage_percent} 
          />
          <ResourceBar 
            label="内存" 
            value={nodeStatus.system_resources.memory?.usage_percent} 
          />
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
};

export default NodeCard;
