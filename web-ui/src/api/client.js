// API客户端工具类
class ApiClient {
    constructor() {
        this.baseUrl = '';
    }

    // 设置API主机地址
    setHost(host) {
        this.baseUrl = `http://${host}`;
        localStorage.setItem('apiHost', host);
    }

    // 获取API主机地址
    getHost() {
        const host = localStorage.getItem('apiHost') || 'localhost:18080';
        this.baseUrl = `http://${host}`;
        return host;
    }

    // 构建完整URL
    buildUrl(path) {
        if (!this.baseUrl) {
            this.getHost();
        }
        return `${this.baseUrl}${path}`;
    }

    // 通用请求方法
    async request(path, options = {}) {
        const url = this.buildUrl(path);
        const response = await fetch(url, {
            ...options,
            // 确保自动携带Cookie
            credentials: 'include',
            headers: {
                'Content-Type': 'application/json',
                ...options.headers,
            },
        });

        // 检查401未授权状态
        if (response.status === 401) {
            throw new Error('UNAUTHORIZED');
        }

        return response;
    }

    // GET 请求
    async get(path) {
        const response = await this.request(path);
        return response.json();
    }

    // POST 请求
    async post(path, data = null) {
        const options = {
            method: 'POST',
        };

        if (data) {
            options.body = JSON.stringify(data);
        }

        const response = await this.request(path, options);
        return response.json();
    }

    // 登录
    async login(password) {
        return this.post('/api/login', { password });
    }

    // 获取状态
    async getStatus() {
        return this.get('/api/status');
    }

    // 获取节点列表
    async getNodes() {
        return this.get('/api/nodes');
    }

    // 获取节点状态
    async getNodeStatus(nodeId) {
        return this.get(`/api/nodes/${nodeId}/status`);
    }

    // 停止服务
    async stop() {
        return this.post('/api/stop');
    }

    // 重启服务
    async restart() {
        return this.post('/api/restart');
    }

    // 退出登录
    async logout() {
        return this.post('/api/logout');
    }

    // ===== Docker 相关 API =====

    // 获取所有可用节点（用于Docker管理）
    async getDockerNodes() {
        return this.get('/api/docker/nodes');
    }

    // 列出指定节点的容器
    async listContainers(nodeId, all = false) {
        return this.get(`/api/docker/containers?node_id=${nodeId}&all=${all}`);
    }

    // 启动容器
    async startContainer(nodeId, containerId) {
        return this.post('/api/docker/start', {
            node_id: nodeId,
            container_id: containerId,
        });
    }

    // 停止容器
    async stopContainer(nodeId, containerId, timeout = 10) {
        return this.post('/api/docker/stop', {
            node_id: nodeId,
            container_id: containerId,
            timeout: timeout,
        });
    }

    // 重启容器
    async restartContainer(nodeId, containerId, timeout = 10) {
        return this.post('/api/docker/restart', {
            node_id: nodeId,
            container_id: containerId,
            timeout: timeout,
        });
    }

    // 获取容器日志
    async getContainerLogs(nodeId, containerId, tail = '100') {
        return this.get(`/api/docker/logs?node_id=${nodeId}&container_id=${containerId}&tail=${tail}`);
    }

    // 获取容器统计信息
    async getContainerStats(nodeId, containerId) {
        return this.get(`/api/docker/stats?node_id=${nodeId}&container_id=${containerId}`);
    }

    // 查看容器详情
    async inspectContainer(nodeId, containerId) {
        return this.get(`/api/docker/inspect?node_id=${nodeId}&container_id=${containerId}`);
    }
}

// 导出单例
const apiClient = new ApiClient();
export default apiClient;
