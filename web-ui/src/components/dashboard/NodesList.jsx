import React from 'react';
import { Loading, EmptyState } from '../../common';
import NodeCard from './NodeCard';

const NodesList = ({ nodes, nodeStatuses, loading }) => {
  if (loading) {
    return (
      <div className="nodes-card">
        <h2>已连接的节点</h2>
        <Loading />
      </div>
    );
  }

  const nodeEntries = Object.entries(nodes);

  if (nodeEntries.length === 0) {
    return (
      <div className="nodes-card">
        <h2>已连接的节点</h2>
        <EmptyState text="暂无连接的节点" />
      </div>
    );
  }

  return (
    <div className="nodes-card">
      <h2>已连接的节点</h2>
      <div className="nodes-grid">
        {nodeEntries.map(([nodeId, nodeInfo]) => (
          <NodeCard
            key={nodeId}
            nodeId={nodeId}
            nodeInfo={nodeInfo}
            nodeStatus={nodeStatuses[nodeId]}
          />
        ))}
      </div>
    </div>
  );
};

export default NodesList;
