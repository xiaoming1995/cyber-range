import axios from 'axios';

// API Base URL - 后端服务地址
const API_BASE_URL = '/api';

// 创建axios实例
const apiClient = axios.create({
  baseURL: API_BASE_URL,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// 请求拦截器 - 添加认证 token
apiClient.interceptors.request.use(
  (config) => {
    // 从 localStorage 获取 token（如果有认证系统）
    const token = localStorage.getItem('auth_token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// 响应拦截器 - 处理统一响应格式 {code, msg, data}
apiClient.interceptors.response.use(
  (response) => {
    // 后端返回格式: {code: 200, msg: "success", data: {...}}
    const { code, msg, data } = response.data;
    if (code === 200) {
      // 将 data 字段放回 response.data，这样就能保持类型兼容
      response.data = data;
      return response;
    }
    // 处理业务错误（例如配额超限）
    return Promise.reject(new Error(msg || '请求失败'));
  },
  (error) => {
    console.error('API Error:', error);
    // 处理网络错误或HTTP错误
    if (error.response) {
      // 服务器返回了错误状态码
      const message = error.response.data?.msg || error.response.statusText || '服务器错误';
      return Promise.reject(new Error(message));
    } else if (error.request) {
      // 请求发出但没有收到响应
      return Promise.reject(new Error('网络错误，请检查连接'));
    } else {
      // 请求配置出错
      return Promise.reject(error);
    }
  }
);

// ==================== TypeScript 类型定义 ====================

export interface Challenge {
  id: string;
  title: string;
  description: string;
  category: string;
  difficulty: 'Easy' | 'Medium' | 'Hard';
  image: string;
  points: number;
  created_at?: string;
  updated_at?: string;
}

export interface Instance {
  id: string;
  user_id: string;
  challenge_id: string;
  container_id: string;
  port: number;
  status: string;
  expires_at: string;
  created_at: string;
}

export interface SubmissionResult {
  correct: boolean;
  message: string;
}

// ==================== API 函数 ====================

/**
 * 获取所有题目列表
 * GET /api/challenges
 */
export const getChallenges = async (): Promise<Challenge[]> => {
  try {
    const response = await apiClient.get<Challenge[]>('/challenges');
    return response.data;
  } catch (error) {
    console.error('获取题目列表失败:', error);
    throw error;
  }
};

/**
 * 启动挑战实例
 * POST /api/challenges/:id/start
 */
export const startInstance = async (challengeId: string): Promise<Instance> => {
  try {
    const response = await apiClient.post<Instance>(`/challenges/${challengeId}/start`);
    return response.data;
  } catch (error: unknown) {
    // 处理特殊错误（例如配额超限）
    const messageText =
      error instanceof Error ? error.message : typeof error === 'string' ? error : (error as { message?: unknown } | null)?.message;
    const messageString = typeof messageText === 'string' ? messageText : '';

    if (messageString && messageString.includes('quota exceeded')) {
      throw new Error('配额超限：每个用户最多同时运行1个实例');
    }
    console.error('启动实例失败:', error);
    throw new Error(messageString || '启动环境失败，请检查Docker是否运行');
  }
};

/**
 * 停止挑战实例
 * POST /api/challenges/:id/stop
 */
export const stopInstance = async (challengeId: string): Promise<{ status: string }> => {
  try {
    const response = await apiClient.post<{ status: string }>(`/challenges/${challengeId}/stop`);
    return response.data;
  } catch (error) {
    console.error('停止实例失败:', error);
    throw error;
  }
};

/**
 * 提交Flag验证
 * POST /api/submit
 */
export const submitFlag = async (
  challengeId: string,
  flag: string
): Promise<SubmissionResult> => {
  try {
    const response = await apiClient.post<SubmissionResult>('/submit', {
      challenge_id: challengeId,
      flag: flag,
    });
    return response.data;
  } catch (error) {
    console.error('Flag验证失败:', error);
    throw error;
  }
};
