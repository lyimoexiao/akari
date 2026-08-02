import type { ForgotPasswordReq, LoginReq, RegisterReq, ResetPasswordReq } from '@/types/auth'
import { z } from 'zod'
import {
  authResponseSchema,
  permissionListSchema,
  userResponseSchema,
  yggdrasilMetadataSchema,
  yggdrasilStatusSchema,
} from '@/types/auth'
import { http } from './index'

const messageSchema = z.string()

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

/** 获取 Yggdrasil API metadata（公开，无需登录） */
export function getYggdrasilMetadata() {
  return http.get({ path: 'yggdrasil/', schema: yggdrasilMetadataSchema })
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
