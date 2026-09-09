/**
 * Wavelet 统一 API 响应信封解析。
 * 与后端 pkg/response.Response 及 axios api-client 约定一致。
 * 全面支持 api-design 标准结构化错误 { error: { code, message, details } } 与兼容字段 error_msg。
 */

export interface ApiEnvelope<T = unknown> {
  data: T | null;
  error?: {
    code: string;
    message: string;
    details?: unknown;
  };
  error_msg?: string;
  meta?: {
    total?: number;
    page?: number;
    per_page?: number;
    total_pages?: number;
  };
}

export class ApiEnvelopeError extends Error {
  readonly status: number;
  readonly code?: string;

  constructor(message: string, status: number, code?: string) {
    super(message);
    this.name = 'ApiEnvelopeError';
    this.status = status;
    this.code = code;
  }
}

function hasEnvelopeShape(value: unknown): value is ApiEnvelope<unknown> {
  if (!value || typeof value !== 'object') {
    return false;
  }
  return 'data' in value || 'error' in value || 'error_msg' in value;
}

/**
 * 解析 fetch 响应体中的 API 信封。
 * - 优先提取 error.message，回退至 error_msg
 * - HTTP 非 2xx：以业务错误或回退文案抛错
 * - HTTP 200 但存在错误体：视为业务失败
 */
export async function readApiEnvelope<T>(
  res: Response,
  fallbackMessage: string,
): Promise<ApiEnvelope<T>> {
  let body: unknown;
  try {
    body = await res.json();
  } catch {
    throw new ApiEnvelopeError(fallbackMessage, res.status);
  }

  if (!hasEnvelopeShape(body)) {
    throw new ApiEnvelopeError(fallbackMessage, res.status);
  }

  const envelope = body as ApiEnvelope<T>;
  const errMsg = envelope.error?.message || envelope.error_msg;
  const errCode = envelope.error?.code;

  if (!res.ok) {
    throw new ApiEnvelopeError(errMsg || fallbackMessage, res.status, errCode);
  }
  if (errMsg) {
    throw new ApiEnvelopeError(errMsg, res.status, errCode);
  }

  return envelope;
}

export async function readApiData<T>(
  res: Response,
  fallbackMessage: string,
): Promise<T> {
  const envelope = await readApiEnvelope<T>(res, fallbackMessage);
  if (envelope.data == null) {
    throw new ApiEnvelopeError(fallbackMessage, res.status);
  }
  return envelope.data;
}
