/**
 * 结构化错误明细 (api-design 规范)
 */
export interface ApiErrorDetail {
  field?: string;
  issue: string;
}

/**
 * 结构化错误体 (api-design 规范)
 */
export interface ApiErrorBody {
  code: string;
  message: string;
  details?: ApiErrorDetail[];
}

/**
 * API 响应通用结构
 * @template T - 数据类型
 */
export interface ApiResponse<T = unknown> {
  /** 响应数据 */
  data: T;
  /** 集合分页与元数据 (api-design 规范) */
  meta?: {
    total?: number;
    page?: number;
    per_page?: number;
    total_pages?: number;
  };
  /** 标准结构化错误体 (api-design 规范) */
  error?: ApiErrorBody;
  /** 响应消息 */
  message?: string;
  /** 兼容字段：旧版错误消息 */
  error_msg?: string;
  /** 兼容字段：旧版错误代码 */
  error_code?: string;
  /** 响应状态码 */
  code?: number;
}

/**
 * API 错误响应结构
 */
export interface ApiError {
  /** 标准结构化错误体 (api-design 规范) */
  error?: ApiErrorBody;
  /** 兼容字段：错误消息 */
  error_msg?: string;
  /** 兼容字段：错误代码 */
  error_code?: string;
  /** 兼容字段：错误详情 */
  details?: unknown;
}

/**
 * 分页请求参数
 */
export interface PaginationParams {
  /** 页码，从 1 开始 */
  page: number;
  /** 每页数量 */
  page_size: number;
}

/**
 * 分页响应数据
 * @template T - 列表项类型
 */
export interface PaginationResponse<T> {
  /** 数据列表 */
  items: T[];
  /** 总数 */
  total: number;
  /** 当前页码 */
  page: number;
  /** 每页数量 */
  page_size: number;
  /** 总页数 */
  total_pages: number;
}

/**
 * HTTP 请求配置
 */
export interface RequestConfig {
  /** 请求超时时间（毫秒） */
  timeout?: number;
  /** 请求头 */
  headers?: Record<string, string>;
  /** 是否携带凭证 */
  withCredentials?: boolean;
}
