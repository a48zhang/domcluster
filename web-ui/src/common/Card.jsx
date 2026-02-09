import React from 'react';

const Card = ({ title, children, className = '', headerRight = null }) => (
  <div className={`detail-card ${className}`}>
    {(title || headerRight) && (
      <div className="card-header">
        {title && <h2>{title}</h2>}
        {headerRight}
      </div>
    )}
    {children}
  </div>
);

export default Card;
