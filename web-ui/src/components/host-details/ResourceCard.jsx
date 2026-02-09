import React from 'react';
import { Card } from '../../common';
import { formatBytes } from '../../utils/format';

const ResourceCard = ({ nodeStatus }) => {
  if (!nodeStatus?.system_resources) return null;

  const { cpu, memory, network } = nodeStatus.system_resources;

  return (
    <Card title="系统资源">
      <div className="resource-grid">
        <div className="resource-item">
          <h3>CPU</h3>
          <div className="resource-details">
            <p>核心数: {cpu?.core_count || 'N/A'}</p>
            <p>使用率: {cpu?.usage_percent?.toFixed(2) || 0}%</p>
          </div>
        </div>
        <div className="resource-item">
          <h3>内存</h3>
          <div className="resource-details">
            <p>总量: {formatBytes(memory?.total_bytes)}</p>
            <p>已使用: {formatBytes(memory?.used_bytes)}</p>
            <p>可用: {formatBytes(memory?.available_bytes)}</p>
            <p>使用率: {memory?.usage_percent?.toFixed(2) || 0}%</p>
          </div>
        </div>
        <div className="resource-item">
          <h3>网络</h3>
          <div className="resource-details">
            <p>接收: {formatBytes(network?.rx_bytes)}</p>
            <p>发送: {formatBytes(network?.tx_bytes)}</p>
          </div>
        </div>
      </div>
    </Card>
  );
};

export default ResourceCard;
