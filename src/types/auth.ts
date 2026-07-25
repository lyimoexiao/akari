export interface AuthResp {
  token: string
  user: UserResp
}

export interface UserResp {
  id: number
  username: string
  email: string
  role: 'super_admin' | 'staff' | 'user'
  email_verified_at: string | null
  created_at: string
}

export interface LoginReq {
  username: string
  password: string
  captcha_id?: string
  user_answer?: Record<string, unknown>
}

export interface RegisterReq {
  username: string
  email: string
  password: string
  captcha_id?: string
  user_answer?: Record<string, unknown>
}

export interface YggdrasilStatusResp {
  has_profile: boolean
  profile_uuid?: string
  profile_name?: string
  last_login_at?: string
  last_login_ip?: string
}

export interface ApiError {
  error: string
}

export interface ForgotPasswordReq {
  email: string
  captcha_id?: string
  user_answer?: Record<string, unknown>
}

export interface ResetPasswordReq {
  token: string
  password: string
}

export const ROLES = {
  SUPER_ADMIN: 'super_admin' as const,
  STAFF: 'staff' as const,
  USER: 'user' as const,
}

export type Role = (typeof ROLES)[keyof typeof ROLES]
