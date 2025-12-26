# Domcluster 需求文档

## 项目概述

Domcluster 是一个用于快速搭建、部署和管理 DomJudge 竞赛判题系统的自动化工具集。DomJudge 是一个开源的在线判题系统，广泛用于编程竞赛和算法练习。本项目旨在简化 DomJudge 集群的部署、配置和运维过程。

## 项目结构

- D8rctl (集群控制端)
- Domclusterd (节点代理端)

## 功能需求

在集群中应包含多种角色：

- D8rctl
- Domserver
- MariaDB 或 MySQL
- JudgeHost
- CDS（可选）
- Docker Register（可选）

### D8rctl

- 提供web端集群控制面板
- 支持主动连接机器（如：用户输入一台主机的ssh root连接字符串，直接将其加入集群）
- 支持被动加入
- 支持在集群中配置Register，优化内网部署速度与节约带宽资源
- 支持集群内日志统一管理与状态监控
- 未来计划：提供cli管理控制、支持控制平面集群化

### Domclusterd

- 作为后台服务在目标机上运行
- 自动配置Docker，根据网络环境自动配置mirror
- 与ctl建立连接
- 实时上传日志、主机监控
