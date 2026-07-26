import type { AddClosetReq, ListClosetReq, ListTexturesReq, RenameClosetReq, UpdateTextureReq } from '@/types/skinlib'
import { z } from 'zod'
import {
  closetIdsSchema,
  listClosetResponseSchema,
  listTexturesResponseSchema,
  textureDetailSchema,
  textureItemSchema,
} from '@/types/skinlib'
import { ApiRequestError, http } from './index'

const messageSchema = z.string()

const API_ROOT = `${import.meta.env.VITE_API_BASE_URL?.replace(/\/$/, '') ?? ''}/api/v1/`

// ── Skin Library ──

export function listTextures(token: string | null, params?: ListTexturesReq) {
  return http.get({ path: 'skinlib', token: token ?? undefined, searchParams: params as Record<string, string | number | undefined>, schema: listTexturesResponseSchema })
}

export function getTexture(token: string | null, tid: number) {
  return http.get({ path: `skinlib/${tid}`, token: token ?? undefined, schema: textureDetailSchema })
}

export async function uploadTexture(token: string, file: File, name: string, type: string) {
  const formData = new FormData()
  formData.append('file', file)
  formData.append('name', name)
  formData.append('type', type)

  const response = await fetch(`${API_ROOT}skinlib`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}` },
    body: formData,
  })

  const payload: unknown = await response.json()
  const envelopeResult = z.object({
    code: z.number(),
    success: z.boolean(),
    msg: z.string(),
    data: z.unknown().optional(),
  }).safeParse(payload)

  if (!envelopeResult.success)
    throw new ApiRequestError('服务器响应格式不正确', response.status)

  const envelope = envelopeResult.data
  if (!response.ok || !envelope.success)
    throw new ApiRequestError(envelope.msg, response.status, envelope.code)

  const dataResult = textureItemSchema.safeParse(envelope.data)
  if (!dataResult.success)
    throw new ApiRequestError('服务器数据格式不正确', response.status, envelope.code)

  return dataResult.data
}

export function updateTexture(token: string, tid: number, data: UpdateTextureReq) {
  return http.put({ path: `skinlib/${tid}`, token, body: data, schema: messageSchema })
}

export function deleteTexture(token: string, tid: number) {
  return http.delete({ path: `skinlib/${tid}`, token, schema: messageSchema })
}

// ── Closet ──

export function listCloset(token: string, params?: ListClosetReq) {
  return http.get({ path: 'closet', token, searchParams: params as Record<string, string | number | undefined>, schema: listClosetResponseSchema })
}

export function getClosetAllIds(token: string) {
  return http.get({ path: 'closet/all-ids', token, schema: closetIdsSchema })
}

export function addToCloset(token: string, data: AddClosetReq) {
  return http.post({ path: 'closet', token, body: data, schema: messageSchema })
}

export function renameClosetItem(token: string, tid: number, data: RenameClosetReq) {
  return http.put({ path: `closet/${tid}`, token, body: data, schema: messageSchema })
}

export function removeFromCloset(token: string, tid: number) {
  return http.delete({ path: `closet/${tid}`, token, schema: messageSchema })
}
