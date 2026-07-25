import type { ApiError } from '@/types/auth'

const BASE_URL = import.meta.env.VITE_API_BASE_URL || ''

// Unified API response envelope from the server.
interface ApiEnvelope<T = unknown> {
  code: number
  success: boolean
  msg: string
  data?: T
  trace_id?: string
}

class HttpClient {
  private baseUrl: string

  constructor(baseUrl: string) {
    this.baseUrl = baseUrl
  }

  private async request<T>(
    method: string,
    path: string,
    body?: unknown,
    token?: string,
  ): Promise<T> {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
    }

    if (token) {
      headers.Authorization = `Bearer ${token}`
    }

    const response = await fetch(`${this.baseUrl}${path}`, {
      method,
      headers,
      body: body ? JSON.stringify(body) : undefined,
    })

    const envelope = (await response.json().catch(() => null)) as ApiEnvelope<T> | null

    // If the server returned our envelope format
    if (envelope && typeof envelope.success === 'boolean') {
      if (envelope.success) {
        return (envelope.data ?? envelope.msg) as T
      }
      throw new ApiRequestError(envelope.msg, response.status, envelope.code, envelope.trace_id)
    }

    // Fallback: unexpected response format
    if (!response.ok) {
      const errorData = envelope as unknown as ApiError | null
      throw new ApiRequestError(
        errorData?.error || `Request failed with status ${response.status}`,
        response.status,
        -1,
        undefined,
      )
    }

    // For non-envelope responses (e.g., plain text), return as-is
    return envelope as unknown as T
  }

  get<T>(path: string, token?: string): Promise<T> {
    return this.request<T>('GET', path, undefined, token)
  }

  post<T>(path: string, body?: unknown, token?: string): Promise<T> {
    return this.request<T>('POST', path, body, token)
  }

  put<T>(path: string, body?: unknown, token?: string): Promise<T> {
    return this.request<T>('PUT', path, body, token)
  }

  delete<T>(path: string, token?: string): Promise<T> {
    return this.request<T>('DELETE', path, undefined, token)
  }
}

export class ApiRequestError extends Error {
  status: number
  code: number
  traceId?: string

  constructor(message: string, status: number, code = -1, traceId?: string) {
    super(message)
    this.name = 'ApiRequestError'
    this.status = status
    this.code = code
    this.traceId = traceId
  }
}

export const http = new HttpClient(`${BASE_URL}/api/v1`)
