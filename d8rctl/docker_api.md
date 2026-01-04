# Docker 管理 API 文档

## 概述

d8rctl 提供了一套 RESTful API 用于管理连接到集群的节点上的 Docker 容器。

## 基础信息

- **Base URL**: `http://127.0.0.1:18080/api/docker`
- **认证**: 所有 API 需要通过 Cookie 认证（session_token）
- **响应格式**: JSON

## API 端点

### 1. 获取节点列表

获取所有已注册的节点列表。

**请求**
```
GET /api/docker/nodes
```

**响应示例**
```json
{
  "nodes": {
    "node-001": {
      "name": "Worker Node 1",
      "role": "worker",
      "version": "1.0.0"
    }
  }
}
```

### 2. 列出容器

列出指定节点的所有容器。

**请求**
```
GET /api/docker/containers?node_id={node_id}&all={all}
```

**参数**
- `node_id` (必需): 节点 ID
- `all` (可选): 是否显示所有容器（包括停止的），默认 false

**响应示例**
```json
{
  "containers": [
    {
      "id": "abc123def456",
      "name": "/nginx",
      "image": "nginx:latest",
      "status": "Up 2 hours",
      "state": "running",
      "created": 1234567890,
      "ports": [
        {
          "IP": "0.0.0.0",
          "PrivatePort": 80,
          "PublicPort": 8080,
          "Type": "tcp"
        }
      ],
      "networks": ["bridge"],
      "mounts": []
    }
  ]
}
```

### 3. 启动容器

启动指定节点的容器。

**请求**
```
POST /api/docker/start
Content-Type: application/json

{
  "node_id": "node-001",
  "container_id": "abc123def456"
}
```

**响应示例**
```json
{
  "message": "container started",
  "container_id": "abc123def456"
}
```

### 4. 停止容器

停止指定节点的容器。

**请求**
```
POST /api/docker/stop
Content-Type: application/json

{
  "node_id": "node-001",
  "container_id": "abc123def456",
  "timeout": 10
}
```

**参数**
- `node_id` (必需): 节点 ID
- `container_id` (必需): 容器 ID
- `timeout` (可选): 停止超时时间（秒），默认 10

**响应示例**
```json
{
  "message": "container stopped",
  "container_id": "abc123def456"
}
```

### 5. 重启容器

重启指定节点的容器。

**请求**
```
POST /api/docker/restart
Content-Type: application/json

{
  "node_id": "node-001",
  "container_id": "abc123def456",
  "timeout": 10
}
```

**参数**
- `node_id` (必需): 节点 ID
- `container_id` (必需): 容器 ID
- `timeout` (可选): 重启超时时间（秒），默认 10

**响应示例**
```json
{
  "message": "container restarted",
  "container_id": "abc123def456"
}
```

### 6. 获取容器日志

获取指定节点的容器日志。

**请求**
```
GET /api/docker/logs?node_id={node_id}&container_id={container_id}&tail={tail}
```

**参数**
- `node_id` (必需): 节点 ID
- `container_id` (必需): 容器 ID
- `tail` (可选): 日志行数，默认 "100"

**响应示例**
```json
{
  "container_id": "abc123def456",
  "logs": "container log output..."
}
```

### 7. 获取容器统计信息

获取指定节点的容器统计信息（CPU、内存、网络等）。

**请求**
```
GET /api/docker/stats?node_id={node_id}&container_id={container_id}
```

**参数**
- `node_id` (必需): 节点 ID
- `container_id` (必需): 容器 ID

**响应示例**
```json
{
  "container_id": "abc123def456",
  "cpu": {
    "usage": 12345678,
    "system": 123456789
  },
  "memory": {
    "usage": 52428800,
    "limit": 1073741824
  },
  "network": {
    "eth0": {
      "rx_bytes": 1024,
      "tx_bytes": 2048
    }
  }
}
```

### 8. 查看容器详情

获取指定节点的容器详细信息。

**请求**
```
GET /api/docker/inspect?node_id={node_id}&container_id={container_id}
```

**参数**
- `node_id` (必需): 节点 ID
- `container_id` (必需): 容器 ID

**响应示例**
```json
{
  "id": "abc123def456789...",
  "name": "/nginx",
  "image": "nginx:latest",
  "state": "running",
  "created": "2024-01-01T00:00:00Z",
  "restart_count": 0,
  "ip": "172.17.0.2"
}
```

## 错误响应

所有错误响应都遵循以下格式：

```json
{
  "error": "错误描述信息"
}
```

**常见错误码**
- `400 Bad Request`: 请求参数错误
- `401 Unauthorized`: 未认证或认证失败
- `404 Not Found`: 节点或容器不存在
- `500 Internal Server Error`: 服务器内部错误

## 使用示例

### 使用 curl

```bash
# 1. 登录获取 session token
curl -X POST http://127.0.0.1:18080/api/login \
  -H "Content-Type: application/json" \
  -c cookies.txt \
  -d '{"password":"your_password"}'

# 2. 获取节点列表
curl http://127.0.0.1:18080/api/docker/nodes \
  -b cookies.txt

# 3. 列出容器
curl "http://127.0.0.1:18080/api/docker/containers?node_id=node-001&all=true" \
  -b cookies.txt

# 4. 启动容器
curl -X POST http://127.0.0.1:18080/api/docker/start \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d '{"node_id":"node-001","container_id":"abc123def456"}'
```

### 使用 JavaScript (Fetch)

```javascript
// 登录
const loginResponse = await fetch('http://127.0.0.1:18080/api/login', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({ password: 'your_password' }),
  credentials: 'include',
});

// 获取容器列表
const containersResponse = await fetch(
  'http://127.0.0.1:18080/api/docker/containers?node_id=node-001',
  { credentials: 'include' }
);
const containers = await containersResponse.json();
console.log(containers);
```

## 注意事项

1. **超时设置**: 所有 Docker 操作都有 30 秒的超时限制
2. **节点连接**: 确保目标节点已连接到 d8rctl 服务器
3. **Docker 可用性**: 节点必须能够访问 Docker 引擎
4. **权限**: 确保 d8rctl 有足够的权限管理 Docker 容器
5. **并发**: 支持同时对多个节点进行 Docker 操作