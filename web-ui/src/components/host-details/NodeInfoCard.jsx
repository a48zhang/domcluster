import React from 'react';
import { Card } from '../../common';

const NodeInfoCard = ({ nodeId, nodeInfo }) => (
  <Card title="节点信息">
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
  </Card>
);

export default NodeInfoCard;
