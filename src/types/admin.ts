import { z } from 'zod'
import { userResponseSchema } from './auth'

export const listUsersResponseSchema = z.object({
  items: z.array(userResponseSchema),
  total: z.number().int().nonnegative(),
  page: z.number().int().positive(),
  page_size: z.number().int().positive(),
  total_pages: z.number().int().nonnegative(),
})

export const roleItemSchema = z.object({
  id: z.number().int().positive(),
  name: z.string().min(1),
  description: z.string(),
  is_default: z.boolean(),
  permissions: z.array(z.string()),
})

export const roleListResponseSchema = z.object({
  items: z.array(roleItemSchema),
})

export const permissionSnapshotSchema = z.object({
  roles: z.array(z.string()),
  rules: z.array(z.object({
    role: z.string(),
    object: z.string(),
    action: z.string(),
  })),
  inheritance: z.array(z.object({
    role: z.string(),
    parent: z.string(),
  })),
})

export const requestLogSchema = z.object({
  id: z.number().int().positive(),
  request_id: z.string(),
  user_id: z.number().int().positive().optional(),
  module: z.string(),
  method: z.string(),
  path: z.string(),
  status: z.number().int(),
  latency_ms: z.number().int().nonnegative(),
  ip: z.string(),
  user_agent: z.string(),
  request_headers: z.string(),
  request_body: z.string(),
  response_headers: z.string(),
  response_body: z.string(),
  created_at: z.string(),
})

export type UserItem = z.infer<typeof userResponseSchema>
export type ListUsersResp = z.infer<typeof listUsersResponseSchema>
export type RoleItem = z.infer<typeof roleItemSchema>
export type PermissionSnapshot = z.infer<typeof permissionSnapshotSchema>
export type RequestLog = z.infer<typeof requestLogSchema>

export interface ListUsersReq {
  readonly [key: string]: string | number | undefined
  readonly page?: number
  readonly page_size?: number
  readonly query?: string
}

export interface SaveRoleReq {
  readonly name: string
  readonly description: string
  readonly permissions: readonly string[]
}
