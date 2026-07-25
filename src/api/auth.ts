import type { AuthResp, ForgotPasswordReq, LoginReq, RegisterReq, ResetPasswordReq, UserResp, YggdrasilStatusResp } from '@/types/auth'
import { http } from '@/api/index'

export function registerUser(data: RegisterReq): Promise<AuthResp> {
  return http.post<AuthResp>('/auth/register', data)
}

export function loginUser(data: LoginReq): Promise<AuthResp> {
  return http.post<AuthResp>('/auth/login', data)
}

export function getCurrentUser(token: string): Promise<UserResp> {
  return http.get<UserResp>('/auth/me', token)
}

export function sendVerificationEmail(token: string): Promise<string> {
  return http.post<string>('/auth/verify-email/send', undefined, token)
}

export function verifyEmail(token: string, verificationToken: string): Promise<string> {
  return http.post<string>('/auth/verify-email', { token: verificationToken }, token)
}

export function logoutUser(token: string): Promise<string> {
  return http.post<string>('/auth/logout', undefined, token)
}

export function getYggdrasilStatus(token: string): Promise<YggdrasilStatusResp> {
  return http.get<YggdrasilStatusResp>('/yggdrasil/user/status', token)
}

export function requestPasswordReset(data: ForgotPasswordReq): Promise<string> {
  return http.post<string>('/auth/password-reset/request', data)
}

export function resetPassword(data: ResetPasswordReq): Promise<string> {
  return http.post<string>('/auth/password-reset', data)
}
