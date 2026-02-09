// 格式化字节大小
export const formatBytes = (bytes) => {
  if (!bytes || bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
};

// 格式化时间戳为本地时间字符串
export const formatTime = (timestamp) => {
  const date = new Date(timestamp);
  return date.toLocaleTimeString();
};

// 格式化最后更新时间为相对时间
export const formatLastUpdate = (timestamp) => {
  const now = new Date();
  const update = new Date(timestamp);
  const diff = Math.floor((now - update) / 1000);

  if (diff < 60) return `${diff}秒前`;
  if (diff < 3600) return `${Math.floor(diff / 60)}分钟前`;
  if (diff < 86400) return `${Math.floor(diff / 3600)}小时前`;
  return `${Math.floor(diff / 86400)}天前`;
};

// 根据使用率获取颜色
export const getResourceColor = (usage) => {
  if (usage < 50) return '#28a745';
  if (usage < 80) return '#ffc107';
  return '#dc3545';
};
