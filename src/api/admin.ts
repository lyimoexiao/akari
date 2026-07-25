import type { ListUsersReq, ListUsersResp } from '@/types/admin'
import { http } from '@/api/index'

export function adminListUsers(token: string, params?: ListUsersReq): Promise<ListUsersResp> {
  const search = params
    ? Object.entries(params)
        .filter(([, v]) => v !== undefined && v !== '')
        .map(([k, v]) => `${encodeURIComponent(k)}=${encodeURIComponent(String(v))}`)
        .join('&')
    : ''
  return http.get<ListUsersResp>(`/admin/users${search ? `?${search}` : ''}`, token)
}

export function adminVerifyEmail(token: string, userId: number): Promise<string> {
  return http.post<string>('/admin/users/verify-email', { user_id: userId }, token)
}

export function adminSetRole(token: string, userId: number, role: string): Promise<string> {
  return http.post<string>('/admin/users/set-role', { user_id: userId, role }, token)
}

export function adminResetPassword(token: string, userId: number, newPassword: string): Promise<string> {
  return http.post<string>('/admin/users/reset-password', { user_id: userId, new_password: newPassword }, token)
}

export function adminDeleteUser(token: string, userId: number): Promise<string> {
  return http.delete<string>(`/admin/users/${userId}`, token)
}
