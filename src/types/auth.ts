import { z } from 'zod'

export const BUILT_IN_ROLES = {
  SUPER_ADMIN: 'super_admin',
  STAFF: 'staff',
  USER: 'user',
} as const

export const userResponseSchema = z.object({
  id: z.number().int().positive(),
  username: z.string().min(1),
  email: z.string().min(1),
  role: z.string().min(1),
  email_verified_at: z.string().nullable().optional(),
  created_at: z.string().min(1),
}).transform(user => ({
  ...user,
  email_verified_at: user.email_verified_at ?? null,
}))

export const authResponseSchema = z.object({
  token: z.string().min(1),
  user: userResponseSchema,
})

export const permissionListSchema = z.object({
  permissions: z.array(z.string()),
})

export const yggdrasilStatusSchema = z.object({
  has_profile: z.boolean(),
  profile_uuid: z.string().optional(),
  profile_name: z.string().optional(),
  texture_skin_id: z.number().int().nullable().optional(),
  texture_cape_id: z.number().int().nullable().optional(),
  last_login_at: z.string().optional(),
  last_login_ip: z.string().optional(),
})

export type AuthResp = z.infer<typeof authResponseSchema>
export type UserResp = z.infer<typeof userResponseSchema>
export type YggdrasilStatusResp = z.infer<typeof yggdrasilStatusSchema>
export type BuiltInRole = (typeof BUILT_IN_ROLES)[keyof typeof BUILT_IN_ROLES]

export type CaptchaAnswer = Readonly<Record<string, unknown>>

export interface LoginReq {
  readonly username: string
  readonly password: string
  readonly captcha_id?: string
  readonly user_answer?: CaptchaAnswer
  readonly captcha_token?: string
}

export type RegisterReq = LoginReq & {
  readonly email: string
}

export interface ForgotPasswordReq {
  readonly email: string
  readonly captcha_id?: string
  readonly user_answer?: CaptchaAnswer
  readonly captcha_token?: string
}

export interface ResetPasswordReq {
  readonly token: string
  readonly password: string
}
