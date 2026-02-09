import React from 'react';
import { getResourceColor } from '../utils/format';

const ResourceBar = ({ label, value, suffix = '%' }) => (
  <div className="resource-item">
    <span className="resource-label">{label}</span>
    <div className="resource-bar">
      <div
        className="resource-fill"
        style={{
          width: `${value || 0}%`,
          backgroundColor: getResourceColor(value || 0),
        }}
      ></div>
    </div>
    <span className="resource-value">
      {(value || 0).toFixed(1)}{suffix}
    </span>
  </div>
);

export default ResourceBar;
