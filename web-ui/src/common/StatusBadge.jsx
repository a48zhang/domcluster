import React from 'react';

const StatusBadge = ({ online, textOnline = '在线', textOffline = '离线' }) => (
  <span className={`status-badge ${online ? 'online' : 'offline'}`}>
    {online ? textOnline : textOffline}
  </span>
);

export default StatusBadge;
