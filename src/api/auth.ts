import type { ForgotPasswordReq, LoginReq, RegisterReq, ResetPasswordReq, YggdrasilMetadataResp } from '@/types/auth'
import { z } from 'zod'
import {
  authResponseSchema,
  permissionListSchema,
  userResponseSchema,
  yggdrasilMetadataSchema,
  yggdrasilStatusSchema,
} from '@/types/auth'
import { ApiRequestError, http } from './index'

const messageSchema = z.string()

// Yggdrasil 协议端点为标准裸 JSON（无 Akari 信封），需要原生 fetch
const API_ROOT = `${import.meta.env.VITE_API_BASE_URL?.replace(/\/$/, '') ?? ''}/api/v1/`

export function registerUser(data: RegisterReq) {
  return http.post({ path: 'auth/register', body: data, schema: authResponseSchema })
}

export function loginUser(data: LoginReq) {
  return http.post({ path: 'auth/login', body: data, schema: authResponseSchema })
}

export function getCurrentUser(token: string) {
  return http.get({ path: 'auth/me', token, schema: userResponseSchema })
}

export function getCurrentPermissions(token: string) {
  return http.get({ path: 'auth/permission', token, schema: permissionListSchema })
}

export function sendVerificationEmail(token: string) {
  return http.post({ path: 'auth/verify-email/send', token, schema: messageSchema })
}

export function verifyEmail(token: string, verificationToken: string) {
  return http.post({ path: 'auth/verify-email', token, body: { token: verificationToken }, schema: messageSchema })
}

export function logoutUser(token: string) {
  return http.post({ path: 'auth/logout', token, schema: messageSchema })
}

export function getYggdrasilStatus(token: string) {
  return http.get({ path: 'yggdrasil/profile/status', token, schema: yggdrasilStatusSchema })
}

/** 获取 Yggdrasil API metadata（公开协议端点，裸 JSON，无需登录） */
export async function getYggdrasilMetadata(): Promise<YggdrasilMetadataResp> {
  let response: Response
  try {
    response = await fetch(`${API_ROOT}yggdrasil/`)
  }
  catch {
    throw new ApiRequestError('无法连接到服务器', 0)
  }
  if (!response.ok)
    throw new ApiRequestError('获取 Yggdrasil 元数据失败', response.status)

  let payload: unknown
  try {
    payload = await response.json()
  }
  catch {
    throw new ApiRequestError('服务器返回了无法解析的响应', response.status)
  }

  const result = yggdrasilMetadataSchema.safeParse(payload)
  if (!result.success)
    throw new ApiRequestError('服务器数据格式不正确', response.status)
  return result.data
}

export function setProfileSkin(token: string, textureTID: number) {
  return http.put({ path: 'yggdrasil/profile/skin', token, body: { texture_tid: textureTID }, schema: messageSchema })
}

export function setProfileCape(token: string, textureTID: number) {
  return http.put({ path: 'yggdrasil/profile/cape', token, body: { texture_tid: textureTID }, schema: messageSchema })
}

export function clearProfileSkin(token: string) {
  return http.delete({ path: 'yggdrasil/profile/skin', token, schema: messageSchema })
}

export function clearProfileCape(token: string) {
  return http.delete({ path: 'yggdrasil/profile/cape', token, schema: messageSchema })
}

export function requestPasswordReset(data: ForgotPasswordReq) {
  return http.post({ path: 'auth/password-reset/request', body: data, schema: messageSchema })
}

export function resetPassword(data: ResetPasswordReq) {
  return http.post({ path: 'auth/password-reset', body: data, schema: messageSchema })
}
