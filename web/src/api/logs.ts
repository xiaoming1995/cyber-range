import adminApi from './admin';

export interface APILog {
    id: string;
    trace_id: string;
    method: string;
    path: string;
    status: number;
    latency_ms: number;
    ip: string;
    user_agent: string;
    user_id?: string;
    error_message?: string;
    request_body?: string;
    response_body?: string;
    created_at: string;
}

export interface LogFilter {
    page?: number;
    page_size?: number;
    status?: number;
    status_min?: number; // 状态码范围
    status_max?: number; // 状态码范围
    path?: string;
    method?: string;
    start_time?: string; // RFC3339
    end_time?: string;   // RFC3339
    trace_id?: string;
}

export interface LogStats {
    total_requests: number;
    error_requests: number;
    avg_latency_ms: number;
    today_requests: number;
    today_errors: number;
}

interface APIResponse<T> {
    code: number;
    msg: string;
    data: T;
}

export interface LogListResponse {
    list: APILog[];
    total: number;
    page: number;
    page_size: number;
}

export const getLogs = async (params: LogFilter): Promise<LogListResponse> => {
    const response = await adminApi.get<APIResponse<LogListResponse>>('/logs', { params });
    if (response.data.code !== 200) {
        throw new Error(response.data.msg || '获取日志失败');
    }
    return response.data.data;
};

export const getLogStats = async (): Promise<LogStats> => {
    const response = await adminApi.get<APIResponse<LogStats>>('/logs/stats');
    if (response.data.code !== 200) {
        throw new Error(response.data.msg || '获取统计失败');
    }
    return response.data.data;
};

export default {
    getLogs,
    getLogStats,
};
