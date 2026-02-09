import React from 'react';

const EmptyState = ({ text = '暂无数据' }) => (
  <div className="empty-state">
    <p>{text}</p>
  </div>
);

export default EmptyState;
