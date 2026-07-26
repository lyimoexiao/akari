import type { ForgotPasswordReq, LoginReq, RegisterReq, ResetPasswordReq } from '@/types/auth'
import { z } from 'zod'
import {
  authResponseSchema,
  permissionListSchema,
  userResponseSchema,
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
  return http.get({ path: 'yggdrasil/user/status', token, schema: yggdrasilStatusSchema })
}

export function requestPasswordReset(data: ForgotPasswordReq) {
  return http.post({ path: 'auth/password-reset/request', body: data, schema: messageSchema })
}

export function resetPassword(data: ResetPasswordReq) {
  return http.post({ path: 'auth/password-reset', body: data, schema: messageSchema })
}
