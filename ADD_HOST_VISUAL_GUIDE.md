# Add Host Feature - Visual Guide

## Command Line Interface (CLI)

### 1. View Help
```bash
$ ./built/d8rctl host

Usage: d8rctl host <command>

Commands:
  add <user@host[:port]>   Add a new host
    Options:
      --password <password>       SSH password
      --key-file <path>           SSH private key file
      --d8rctl-address <address>  D8rctl server address (default: localhost:50051)

  list                     List all hosts
  remove <node_id>         Remove a host

Example:
  d8rctl host add root@192.168.1.100 --password mypassword
  d8rctl host add user@server.com:22 --key-file ~/.ssh/id_rsa
```

### 2. Add a Host (Password Authentication)
```bash
$ ./built/d8rctl host add root@192.168.1.100 --password mypassword --d8rctl-address 192.168.1.1:50051

✓ Host added successfully!
  Node ID:  server-100
  Hostname: server-100
  OS:       Linux
  Arch:     x86_64

The host should now appear in the node list.
Use 'd8rctl pod list' to verify.
```

### 3. List Hosts
```bash
$ ./built/d8rctl pod list

Connected nodes: 1

Node ID: server-100
  Name:    server-100
  Role:    worker
  Version: v1.0.0
```

## Web User Interface

### 1. Dashboard - Empty State
```
┌─────────────────────────────────────────────────────────────┐
│ Domcluster Dashboard                        [退出登录]      │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  服务状态                                                    │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ 运行状态: 运行中  进程ID: 12345  运行时间: 1h 23m  │   │
│  │ [停止服务] [重启服务]                                │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                               │
│  已连接的节点                         [+ 添加主机]          │
│  ┌─────────────────────────────────────────────────────┐   │
│  │                                                       │   │
│  │              暂无连接的节点                          │   │
│  │         [添加第一个主机]                              │   │
│  │                                                       │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                               │
└─────────────────────────────────────────────────────────────┘
```

### 2. Add Host Modal
```
┌─────────────────────────────────────────────────────────────┐
│  添加被控主机                                           [×] │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  SSH 连接字符串                                             │
│  [root@192.168.1.100                              ]         │
│  例如: root@192.168.1.100 或 user@server.com:22             │
│                                                               │
│  认证方式                                                   │
│  [  密码  ] [ 密钥文件 ]                                    │
│                                                               │
│  SSH 密码                                                   │
│  [••••••••••                                      ]         │
│                                                               │
│  D8rctl 服务地址                                            │
│  [192.168.1.1:50051                               ]         │
│  被控主机可访问的 D8rctl 服务地址                          │
│                                                               │
│                                       [取消]  [添加主机]     │
└─────────────────────────────────────────────────────────────┘
```

### 3. Adding in Progress
```
┌─────────────────────────────────────────────────────────────┐
│  添加被控主机                                           [×] │
├─────────────────────────────────────────────────────────────┤
│  [All fields as above, but disabled]                        │
│                                                               │
│                                       [取消]  [正在添加...]  │
└─────────────────────────────────────────────────────────────┘
```

### 4. Success Result
```
┌─────────────────────────────────────────────────────────────┐
│  添加被控主机                                           [×] │
├─────────────────────────────────────────────────────────────┤
│  [All fields as above, disabled]                            │
│                                                               │
│  ┌───────────────────────────────────────────────────────┐ │
│  │ ✓  主机添加成功！                                    │ │
│  │    节点 ID: server-100                                │ │
│  │    主机名: server-100                                 │ │
│  │    操作系统: Linux / x86_64                          │ │
│  └───────────────────────────────────────────────────────┘ │
│                                                               │
│                                                       [关闭]  │
└─────────────────────────────────────────────────────────────┘
```

### 5. Dashboard - With Nodes
```
┌─────────────────────────────────────────────────────────────┐
│ Domcluster Dashboard                        [退出登录]      │
├─────────────────────────────────────────────────────────────┤
│  已连接的节点                         [+ 添加主机]          │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  ┌────────────────────────────────────────────────┐ │   │
│  │  │ server-100                      │ abc123-def456│ │   │
│  │  │ 角色: worker  版本: v1.0.0                     │ │   │
│  │  │                                                 │ │   │
│  │  │ CPU    ████████░░░░░░░░░░░  45.2%             │ │   │
│  │  │ 内存   ██████████░░░░░░░░░  62.8%             │ │   │
│  │  │ 容器   3/5                                     │ │   │
│  │  │                                                 │ │   │
│  │  │ ● 在线  2分钟前                                │ │   │
│  │  └────────────────────────────────────────────────┘ │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

## Key Features Demonstrated

### CLI
- ✅ Simple command syntax
- ✅ Support for password and key file authentication
- ✅ Clear success/error messages
- ✅ Automatic verification through pod list

### Web UI
- ✅ User-friendly modal interface
- ✅ Form validation
- ✅ Toggle between password and key file
- ✅ Real-time feedback
- ✅ Success confirmation with details
- ✅ Auto-close after success
- ✅ Integration with dashboard
- ✅ Empty state handling

## Architecture Benefits

1. **Dual Interface**: Both CLI and Web UI support
2. **Security**: Authenticated API endpoints
3. **Automation**: One-command host provisioning
4. **Feedback**: Clear progress and error messages
5. **Integration**: Seamlessly integrates with existing node management
