import React from 'react';
import { Card, EmptyState } from '../../common';

const ContainerTable = ({ 
  containers, 
  showAll, 
  onShowAllChange, 
  onStart, 
  onStop, 
  onRestart 
}) => {
  const headerRight = (
    <label className="checkbox-label">
      <input
        type="checkbox"
        checked={showAll}
        onChange={(e) => onShowAllChange(e.target.checked)}
      />
      显示所有容器
    </label>
  );

  return (
    <Card title="Docker 容器" headerRight={headerRight}>
      {containers.length === 0 ? (
        <EmptyState text="没有找到容器" />
      ) : (
        <div className="containers-table">
          <table>
            <thead>
              <tr>
                <th>容器 ID</th>
                <th>名称</th>
                <th>镜像</th>
                <th>状态</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              {containers.map((container) => (
                <tr key={container.id}>
                  <td>{container.id?.substring(0, 12)}</td>
                  <td>{container.names?.[0] || 'N/A'}</td>
                  <td>{container.image || 'N/A'}</td>
                  <td>
                    <span className={`container-status ${container.state?.toLowerCase()}`}>
                      {container.state || 'unknown'}
                    </span>
                  </td>
                  <td>
                    <div className="container-actions">
                      {container.state === 'running' ? (
                        <>
                          <button
                            className="btn-small btn-stop"
                            onClick={() => onStop(container.id)}
                          >
                            停止
                          </button>
                          <button
                            className="btn-small btn-restart"
                            onClick={() => onRestart(container.id)}
                          >
                            重启
                          </button>
                        </>
                      ) : (
                        <button
                          className="btn-small btn-start"
                          onClick={() => onStart(container.id)}
                        >
                          启动
                        </button>
                      )}
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </Card>
  );
};

export default ContainerTable;
