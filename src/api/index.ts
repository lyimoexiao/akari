import type { ZodType } from 'zod'
import ky from 'ky'
import { z } from 'zod'

const API_ROOT = `${import.meta.env.VITE_API_BASE_URL?.replace(/\/$/, '') ?? ''}/api/v1/`

const apiEnvelopeSchema = z.object({
  code: z.number(),
  success: z.boolean(),
  msg: z.string(),
  data: z.unknown().optional(),
  trace_id: z.string().optional(),
})

interface RequestOptions<T> {
  readonly method: 'GET' | 'POST' | 'PUT' | 'DELETE'
  readonly path: string
  readonly schema: ZodType<T>
  readonly body?: unknown
  readonly token?: string
  readonly searchParams?: Readonly<Record<string, string | number | undefined>>
  readonly signal?: AbortSignal
}

export class ApiRequestError extends Error {
  readonly name = 'ApiRequestError'

  constructor(
    message: string,
    readonly status: number,
    readonly code = -1,
    readonly traceId?: string,
    options?: ErrorOptions,
  ) {
    super(message, options)
  }
}

function createSearchParams(values?: RequestOptions<unknown>['searchParams']): URLSearchParams | undefined {
  if (!values)
    return undefined

  const searchParams = new URLSearchParams()
  for (const [key, value] of Object.entries(values)) {
    if (value !== undefined && value !== '')
      searchParams.set(key, String(value))
  }
  return searchParams
}

async function request<T>(options: RequestOptions<T>): Promise<T> {
  let response: Response
  try {
    response = await ky(`${API_ROOT}${options.path}`, {
      method: options.method,
      headers: options.token ? { Authorization: `Bearer ${options.token}` } : undefined,
      json: options.body,
      searchParams: createSearchParams(options.searchParams),
      signal: options.signal,
      timeout: 15_000,
      retry: 0,
      throwHttpErrors: false,
    })
  }
  catch (error) {
    throw new ApiRequestError('无法连接到服务器', 0, -1, undefined, { cause: error })
  }

  let payload: unknown
  try {
    payload = await response.json()
  }
  catch (error) {
    throw new ApiRequestError('服务器返回了无法解析的响应', response.status, -1, undefined, { cause: error })
  }

  const envelopeResult = apiEnvelopeSchema.safeParse(payload)
  if (!envelopeResult.success) {
    throw new ApiRequestError('服务器响应格式不正确', response.status, -1, undefined, {
      cause: envelopeResult.error,
    })
  }

  const envelope = envelopeResult.data
  if (!response.ok || !envelope.success) {
    throw new ApiRequestError(envelope.msg, response.status, envelope.code, envelope.trace_id)
  }

  const dataResult = options.schema.safeParse(envelope.data ?? envelope.msg)
  if (!dataResult.success) {
    throw new ApiRequestError('服务器数据格式不正确', response.status, envelope.code, envelope.trace_id, {
      cause: dataResult.error,
    })
  }
  return dataResult.data
}

export const http = {
  get<T>(options: Omit<RequestOptions<T>, 'method'>): Promise<T> {
    return request({ ...options, method: 'GET' })
  },
  post<T>(options: Omit<RequestOptions<T>, 'method'>): Promise<T> {
    return request({ ...options, method: 'POST' })
  },
  put<T>(options: Omit<RequestOptions<T>, 'method'>): Promise<T> {
    return request({ ...options, method: 'PUT' })
  },
  delete<T>(options: Omit<RequestOptions<T>, 'method'>): Promise<T> {
    return request({ ...options, method: 'DELETE' })
  },
}
