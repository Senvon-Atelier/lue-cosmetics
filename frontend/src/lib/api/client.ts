import Axios, { AxiosError, AxiosRequestConfig } from 'axios';

export const apiClient = Axios.create({
  baseURL: import.meta.env.VITE_API_URL || '/api/v1',
  withCredentials: true,
});

interface ErrorEnvelope {
  error?: { code?: string; message?: string; fields?: Record<string, string> };
}

/** Typed error preserving the backend's httpx error envelope. */
export class ApiError extends Error {
  status: number;
  code: string;
  fields?: Record<string, string>;

  constructor(status: number, code: string, message: string, fields?: Record<string, string>) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.code = code;
    this.fields = fields;
  }
}

apiClient.interceptors.response.use(
  (response) => response,
  (error: AxiosError<ErrorEnvelope>) => {
    if (error.response) {
      const body = error.response.data?.error;
      throw new ApiError(
        error.response.status,
        body?.code ?? 'unknown',
        body?.message ?? `Request failed with status ${error.response.status}`,
        body?.fields,
      );
    }
    if (error.request) {
      throw new ApiError(0, 'network', 'No response from server. Please check your connection.');
    }
    throw new ApiError(0, 'unknown', error.message || 'An unexpected error occurred');
  },
);

/** Orval mutator: every generated call goes through apiClient and unwraps .data. */
export const customInstance = <T>(config: AxiosRequestConfig): Promise<T> =>
  apiClient(config).then((res) => res.data as T);
