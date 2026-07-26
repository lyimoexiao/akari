import type { ListUsersReq, SaveRoleReq } from '@/types/admin'
import { z } from 'zod'
import {
  listUsersResponseSchema,
  permissionSnapshotSchema,
  requestLogSchema,
  roleItemSchema,
  roleListResponseSchema,
} from '@/types/admin'
import { http } from './index'

const messageSchema = z.string()

export function adminListUsers(token: string, params: ListUsersReq = {}) {
  return http.get({ path: 'users', token, searchParams: params, schema: listUsersResponseSchema })
}

export function adminVerifyEmail(token: string, userId: number) {
  return http.post({ path: 'users/verify-email', token, body: { user_id: userId }, schema: messageSchema })
}

export function adminSetRole(token: string, userId: number, role: string) {
  return http.post({ path: 'roles/assign', token, body: { user_id: userId, role }, schema: messageSchema })
}

export function adminResetPassword(token: string, userId: number, newPassword: string) {
  return http.post({ path: 'users/reset-password', token, body: { user_id: userId, new_password: newPassword }, schema: messageSchema })
}

export function adminDeleteUser(token: string, userId: number) {
  return http.delete({ path: `users/${userId}`, token, schema: messageSchema })
}

export function adminListRoles(token: string) {
  return http.get({ path: 'roles', token, schema: roleListResponseSchema })
}

export function adminCreateRole(token: string, data: SaveRoleReq) {
  return http.post({ path: 'roles', token, body: data, schema: roleItemSchema })
}

export function adminUpdateRole(token: string, roleId: number, data: SaveRoleReq) {
  return http.put({ path: `roles/${roleId}`, token, body: data, schema: roleItemSchema })
}

export function adminDeleteRole(token: string, roleId: number) {
  return http.delete({ path: `roles/${roleId}`, token, schema: messageSchema })
}

export function adminSetDefaultRole(token: string, roleId: number) {
  return http.put({ path: `roles/${roleId}/default`, token, schema: messageSchema })
}

export function adminGetPermissionSnapshot(token: string) {
  return http.get({ path: 'permissions', token, schema: permissionSnapshotSchema })
}

export function adminGetRequestLog(token: string, requestId: string) {
  return http.get({ path: `request-logs/${encodeURIComponent(requestId)}`, token, schema: requestLogSchema })
}
