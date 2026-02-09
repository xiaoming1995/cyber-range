import axios from 'axios';
import { getAdminToken } from '../admin/auth';

// API Base URL
const API_BASE = '/api';

// 创建 Axios 实例
const adminApi = axios.create({
    baseURL: `${API_BASE}/admin`,
    timeout: 10000,
});

// 请求拦截器 - 自动添加 Token
adminApi.interceptors.request.use(
    (config) => {
        const token = getAdminToken();
        if (token) {
            config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
    },
    (error) => {
        return Promise.reject(error);
    }
);

// 响应拦截器 - 统一错误处理
adminApi.interceptors.response.use(
    (response) => {
        return response;
    },
    (error) => {
        // 401 未授权 - 可能 Token 过期
        if (error.response?.status === 401) {
            // 清除本地存储的认证信息
            localStorage.removeItem('cyber_range_admin_token');
            localStorage.removeItem('cyber_range_admin_info');

            // 跳转到登录页
            if (window.location.pathname !== '/admin/login') {
                window.location.href = '/admin/login';
            }
        }

        return Promise.reject(error);
    }
);

export default adminApi;

// API Response 格式
interface APIResponse<T = any> {
    code: number;
    msg: string;
    data?: T;
}

// ========== 认证相关 ==========

export interface Admin {
    id: string;
    username: string;
    email: string;
    name: string;
}

export interface LoginResponse {
    token: string;
    admin: Admin;
}

/**
 * 管理员登录（不使用拦截器，因为登录时还没有 Token）
 */
export async function adminLogin(username: string, password: string): Promise<LoginResponse> {
    const response = await axios.post<APIResponse<LoginResponse>>(
        `${API_BASE}/admin/login`,
        { username, password }
    );

    if (response.data.code !== 200) {
        throw new Error(response.data.msg || '登录失败');
    }

    return response.data.data!;
}

// ========== 题目管理 ==========

export interface Challenge {
    id: string;
    title: string;
    description: string;
    hint?: string;
    category: string;
    difficulty: string;
    image: string;
    image_id?: string;
    port: number;
    points: number;
    flag: string;
    docker_host_id?: string;
    memory_limit?: number;
    cpu_limit?: number;
    status: 'published' | 'unpublished';
    published_at?: string;
    unpublished_at?: string;
    created_at: string;
    updated_at: string;
}

export interface ChallengeListQuery {
    page?: number;
    pageSize?: number;
    category?: string;
    difficulty?: string;
    status?: string;
    search?: string;
}

export interface ChallengeListResponse {
    list: Challenge[];
    total: number;
    page: number;
    pageSize: number;
}

/**
 * 获取题目列表（自动带 Token）
 */
export async function listChallenges(query: ChallengeListQuery): Promise<ChallengeListResponse> {
    const response = await adminApi.get<APIResponse<ChallengeListResponse>>('/challenges', {
        params: query,
    });

    if (response.data.code !== 200) {
        throw new Error(response.data.msg || '获取失败');
    }

    return response.data.data!;
}

/**
 * 获取题目详情（自动带 Token）
 */
export async function getChallenge(id: string): Promise<Challenge> {
    const response = await adminApi.get<APIResponse<Challenge>>(`/challenges/${id}`);

    if (response.data.code !== 200) {
        throw new Error(response.data.msg || '获取失败');
    }

    return response.data.data!;
}

export interface ChallengeFormData {
    title: string;
    descriptionHtml: string;
    hintHtml?: string;
    category: string;
    difficulty: string;
    image?: string;         // 可选，后端会自动根据 image_id 填充
    image_id?: string;      // 关联镜像 ID
    docker_host_id: string; // Docker 主机 ID
    port: number;
    memory_limit?: number;  // 内存限制
    cpu_limit?: number;     // CPU 限制
    flag: string;
    points: number;
    status: 'published' | 'unpublished';
}

/**
 * 创建题目（自动带 Token）
 */
export async function createChallenge(data: ChallengeFormData): Promise<any> {
    const response = await adminApi.post<APIResponse>('/challenges', data);

    if (response.data.code !== 200) {
        throw new Error(response.data.msg || '创建失败');
    }

    return response.data.data;
}

/**
 * 更新题目（自动带 Token）
 */
export async function updateChallenge(id: string, data: ChallengeFormData): Promise<void> {
    const response = await adminApi.put<APIResponse>(`/challenges/${id}`, data);

    if (response.data.code !== 200) {
        throw new Error(response.data.msg || '更新失败');
    }
}

/**
 * 删除题目（自动带 Token）
 */
export async function deleteChallenge(id: string): Promise<void> {
    const response = await adminApi.delete<APIResponse>(`/challenges/${id}`);

    if (response.data.code !== 200) {
        throw new Error(response.data.msg || '删除失败');
    }
}

/**
 * 更新题目状态（自动带 Token）
 */
export async function updateChallengeStatus(
    id: string,
    status: 'published' | 'unpublished'
): Promise<void> {
    const response = await adminApi.put<APIResponse>(`/challenges/${id}/status`, { status });

    if (response.data.code !== 200) {
        throw new Error(response.data.msg || '状态更新失败');
    }
}

// ========== 实例管理 ==========

export interface InstanceQuery {
    status?: string;
    challenge?: string;
    page?: number;
    pageSize?: number;
}

export interface InstanceListResponse {
    list: any[];
    total: number;
    page: number;
    pageSize: number;
}

/**
 * 获取实例列表（自动带 Token）
 */
export async function listInstances(query: InstanceQuery): Promise<InstanceListResponse> {
    const response = await adminApi.get<APIResponse<InstanceListResponse>>('/instances', {
        params: query,
    });

    if (response.data.code !== 200) {
        throw new Error(response.data.msg || '获取失败');
    }

    return response.data.data!;
}

// 容器资源统计
export interface ContainerStats {
    container_id: string;
    cpu_percent: number;
    memory_usage: number;   // 字节
    memory_limit: number;   // 字节
    memory_percent: number;
    network_rx: number;     // 接收字节
    network_tx: number;     // 发送字节
}

/**
 * 获取实例实时资源统计（自动带 Token）
 */
export async function getInstanceStats(instanceId: string): Promise<ContainerStats> {
    const response = await adminApi.get<APIResponse<ContainerStats>>(`/instances/${instanceId}/stats`);

    if (response.data.code !== 200) {
        throw new Error(response.data.msg || '获取统计失败');
    }

    return response.data.data!;
}

// 容器日志
export interface ContainerLogs {
    logs: string;
    container_id: string;
}

/**
 * 获取实例容器日志（自动带 Token）
 */
export async function getInstanceLogs(instanceId: string, tail: number = 200): Promise<ContainerLogs> {
    const response = await adminApi.get<APIResponse<ContainerLogs>>(`/instances/${instanceId}/logs`, {
        params: { tail },
    });

    if (response.data.code !== 200) {
        throw new Error(response.data.msg || '获取日志失败');
    }

    return response.data.data!;
}

// ========== 提交记录 ==========

export interface SubmissionQuery {
    user?: string;
    challenge?: string;
    result?: string;
    page?: number;
    pageSize?: number;
}

export interface SubmissionListResponse {
    list: any[];
    total: number;
    page: number;
    pageSize: number;
}

/**
 * 获取提交记录列表（自动带 Token）
 */
export async function listSubmissions(query: SubmissionQuery): Promise<SubmissionListResponse> {
    const response = await adminApi.get<APIResponse<SubmissionListResponse>>('/submissions', {
        params: query,
    });

    if (response.data.code !== 200) {
        throw new Error(response.data.msg || '获取失败');
    }

    return response.data.data!;
}

// ========== 总览统计 ==========

export interface SubmissionView {
    id: string;
    userDisplayName: string;
    challengeTitle: string;
    result: string;
    createdAt: string;
}

export interface HotChallenge {
    title: string;
    count: number;
}

export interface OverviewStats {
    todayInstances: number;
    runningInstances: number;
    todaySubmissions: number;
    todayCorrectRate: number;
    recentSubmissions: SubmissionView[];
    hotChallenges: HotChallenge[];
}

/**
 * 获取总览统计数据（自动带 Token）
 */
export async function getOverviewStats(): Promise<OverviewStats> {
    const response = await adminApi.get<APIResponse<OverviewStats>>('/overview/stats');

    if (response.data.code !== 200) {
        throw new Error(response.data.msg || '获取失败');
    }

    return response.data.data!;
}

// ========== Docker 镜像管理 ==========

export interface DockerImage {
    id: string;
    name: string;
    tag: string;
    registry: string;
    size?: number;
    digest?: string;
    description?: string;
    recommended_memory?: number;
    recommended_cpu?: number;
    is_available: boolean;
    last_sync_at?: string;
    created_at: string;
    updated_at: string;
}

/**
 * 获取镜像列表
 */
export async function getImages(): Promise<DockerImage[]> {
    const response = await adminApi.get<APIResponse<DockerImage[]>>('/images');

    if (response.data.code !== 200) {
        throw new Error(response.data.msg || '获取镜像列表失败');
    }

    return response.data.data || [];
}

/**
 * 注册镜像
 */
export async function registerImage(data: {
    name: string;
    tag: string;
    description?: string;
}): Promise<DockerImage> {
    const response = await adminApi.post<APIResponse<DockerImage>>('/images', data);

    if (response.data.code !== 200) {
        throw new Error(response.data.msg || '注册镜像失败');
    }

    return response.data.data!;
}

/**
 * 手动预加载镜像到所有主机
 */
export async function preloadImages(): Promise<void> {
    const response = await adminApi.post<APIResponse>('/images/preload');

    if (response.data.code !== 200) {
        throw new Error(response.data.msg || '预加载失败');
    }
}

// 同步结果
export interface SyncImagesResult {
    synced_count: number;
    registry_url: string;
}

/**
 * 从 Registry 同步镜像到数据库
 */
export async function syncImagesFromRegistry(registryUrl?: string): Promise<SyncImagesResult> {
    const response = await adminApi.post<APIResponse<SyncImagesResult>>('/images/sync', {
        registry_url: registryUrl,
    });

    if (response.data.code !== 200) {
        throw new Error(response.data.msg || '同步失败');
    }

    return response.data.data!;
}

// 镜像上传结果
export interface ImageUploadResult {
    image_name: string;
    registry_tag: string;
    pushed: boolean;
}

/**
 * 上传并导入镜像文件（.tar 或 .tar.gz）
 */
export async function uploadImage(
    file: File,
    tag: string = '',
    onProgress?: (percent: number) => void
): Promise<ImageUploadResult> {
    const formData = new FormData();
    formData.append('file', file);
    if (tag) {
        formData.append('tag', tag);
    }

    const response = await adminApi.post<APIResponse<ImageUploadResult>>('/images/upload', formData, {
        headers: {
            'Content-Type': 'multipart/form-data',
        },
        timeout: 300000, // 5 分钟超时，大文件需要更长时间
        onUploadProgress: (progressEvent) => {
            if (onProgress && progressEvent.total) {
                const percent = Math.round((progressEvent.loaded * 100) / progressEvent.total);
                onProgress(percent);
            }
        },
    });

    if (response.data.code !== 200) {
        throw new Error(response.data.msg || '上传失败');
    }

    return response.data.data!;
}
// 删除镜像
export async function deleteImage(id: string): Promise<void> {
    const response = await adminApi.delete<APIResponse<null>>(`/images/${id}`);
    if (response.data.code !== 200) {
        throw new Error(response.data.msg || '删除失败');
    }
}

// Overview Stats

