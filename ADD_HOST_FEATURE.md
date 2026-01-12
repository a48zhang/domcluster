# 添加被控主机功能 (Add Controlled Host Feature)

## 功能概述

该功能允许用户通过 SSH 连接字符串和密码或密钥文件，自动完成以下操作：
1. 连接到指定的远程主机
2. 在远程主机上安装 domclusterd
3. 配置 domclusterd 连接到当前的 d8rctl 控制端
4. 启动 domclusterd 服务

该功能可以通过命令行（CLI）和 Web UI 两种方式使用。

## 使用方法

### 命令行（CLI）

#### 添加主机

使用密码认证：
```bash
d8rctl host add root@192.168.1.100 --password mypassword --d8rctl-address 192.168.1.1:50051
```

使用密钥文件认证：
```bash
d8rctl host add user@server.com:22 --key-file /root/.ssh/id_rsa --d8rctl-address 192.168.1.1:50051
```

参数说明：
- `<user@host[:port]>`: SSH 连接字符串，格式为 `用户名@主机地址[:端口]`，端口默认为 22
- `--password <password>`: SSH 密码（与 --key-file 二选一）
- `--key-file <path>`: SSH 私钥文件路径（与 --password 二选一）
- `--d8rctl-address <address>`: d8rctl 服务地址，远程主机需要能够访问该地址（默认: localhost:50051）

#### 列出所有主机

```bash
d8rctl host list
```

或使用原有命令：
```bash
d8rctl pod list
```

#### 移除主机

```bash
d8rctl host remove <node_id>
```

注意：当前需要手动在远程主机上停止 domclusterd 服务。

### Web UI

1. 登录 Domcluster Dashboard
2. 在"已连接的节点"区域，点击右上角的 **"+ 添加主机"** 按钮
3. 在弹出的对话框中填写以下信息：
   - SSH 连接字符串（例如：`root@192.168.1.100` 或 `user@server.com:22`）
   - 选择认证方式：
     - **密码**：输入 SSH 密码
     - **密钥文件**：输入服务器上密钥文件的绝对路径
   - D8rctl 服务地址（远程主机可访问的地址）
4. 点击 **"添加主机"** 按钮
5. 等待添加完成，成功后对话框会显示节点信息并自动关闭

## 技术实现

### 后端（Go）

#### SSH 连接管理器
- 文件：`d8rctl/services/ssh_manager.go`
- 功能：
  - 解析 SSH 连接字符串
  - 支持密码和密钥文件两种认证方式
  - 提供命令执行、文件上传等功能

#### 主机供应器
- 文件：`d8rctl/services/host_provisioner.go`
- 功能：
  - 查找本地的 domclusterd 二进制文件
  - 通过 SSH 连接到远程主机
  - 上传 domclusterd 到远程主机
  - 创建配置文件
  - 启动 domclusterd 服务

#### API 端点
- HTTP API: `POST /api/hosts/add`（需要认证）
- CLI Server: `/hosts/add`（通过 Unix Socket，无需认证）

### 前端（React）

#### AddHostModal 组件
- 文件：`web-ui/src/components/AddHostModal.jsx`
- 功能：
  - 表单验证
  - 认证方式切换（密码/密钥）
  - 实时显示添加进度和结果
  - 成功后自动关闭

#### Dashboard 集成
- 在节点列表区域添加 "添加主机" 按钮
- 空状态时显示 "添加第一个主机" 提示

## 工作流程

```
用户输入 SSH 信息
    ↓
d8rctl 通过 SSH 连接到远程主机
    ↓
获取远程主机系统信息（OS, Arch, Hostname）
    ↓
检查操作系统是否支持（仅支持 Linux）
    ↓
上传 domclusterd 二进制文件到 /tmp/domclusterd
    ↓
创建配置目录 /var/lib/domcluster
    ↓
移动二进制文件到 /usr/local/bin/domclusterd
    ↓
创建配置文件 /var/lib/domcluster/config.yaml
    ↓
启动 domclusterd 服务
    ↓
domclusterd 自动注册到 d8rctl
    ↓
完成！主机出现在节点列表中
```

## 安全性考虑

1. **SSH 认证**：支持密码和密钥文件两种安全的认证方式
2. **主机密钥验证**：当前使用 `InsecureIgnoreHostKey`，生产环境应该验证主机密钥
3. **权限管理**：
   - Web UI 接口需要登录认证
   - CLI 接口通过 Unix Socket 本地访问，权限由操作系统控制
4. **配置隔离**：每个主机的配置独立存储在 `/var/lib/domcluster/config.yaml`

## 系统要求

### 控制端（d8rctl）
- 需要访问 domclusterd 二进制文件（自动查找）
- 能够访问远程主机的 SSH 端口

### 被控主机
- 操作系统：Linux（自动检测）
- SSH 服务：已启动并可访问
- 网络：能够访问 d8rctl 的 gRPC 端口（默认 50051）
- 权限：能够创建目录、移动文件（最好使用 sudo 权限的用户）

## 故障排查

### 连接失败
- 检查 SSH 连接字符串格式是否正确
- 验证密码或密钥文件是否正确
- 确认远程主机 SSH 服务是否运行
- 检查网络连接和防火墙规则

### 安装失败
- 确认远程主机是 Linux 系统
- 检查用户是否有足够的权限
- 验证 domclusterd 二进制文件是否存在

### 节点未出现
- 检查 d8rctl-address 是否正确
- 确认远程主机能够访问该地址
- 查看远程主机的 domclusterd 日志：`cat /tmp/domclusterd.log`

## 未来改进

1. 支持批量添加多个主机
2. 添加主机健康检查
3. 实现完整的主机移除功能（自动停止远程服务）
4. 支持主机密钥验证
5. 添加进度条显示详细的安装步骤
6. 支持自定义 domclusterd 安装路径和配置
