import type { ListUsersResp } from '@/types/admin'
import { useAsyncState } from '@vueuse/core'
import { reactive, shallowRef } from 'vue'
import {
  adminDeleteUser,
  adminListUsers,
  adminResetPassword,
  adminSetRole,
  adminVerifyEmail,
} from '@/api/admin'
import { useAuthStore } from '@/stores/auth'

const EMPTY_RESULT: ListUsersResp = {
  items: [],
  total: 0,
  page: 1,
  page_size: 20,
  total_pages: 0,
}

export function useAdminUsers() {
  const authStore = useAuthStore()
  const filters = reactive({ page: 1, pageSize: 20, query: '' })
  const mutationLoading = shallowRef(false)

  async function fetchUsers(): Promise<ListUsersResp> {
    const token = requireToken()
    return adminListUsers(token, {
      page: filters.page,
      page_size: filters.pageSize,
      query: filters.query || undefined,
    })
  }

  const {
    state: result,
    isLoading,
    error,
    execute: refresh,
  } = useAsyncState(fetchUsers, EMPTY_RESULT, { resetOnExecute: false })

  async function verifyEmail(userId: number): Promise<void> {
    await runMutation(() => adminVerifyEmail(requireToken(), userId), true)
  }

  async function setRole(userId: number, role: string): Promise<void> {
    await runMutation(() => adminSetRole(requireToken(), userId, role), true)
  }

  async function resetPassword(userId: number, password: string): Promise<void> {
    await runMutation(() => adminResetPassword(requireToken(), userId, password), false)
  }

  async function deleteUser(userId: number): Promise<void> {
    await runMutation(() => adminDeleteUser(requireToken(), userId), true)
  }

  async function runMutation(action: () => Promise<string>, reload: boolean): Promise<void> {
    mutationLoading.value = true
    try {
      await action()
      if (reload)
        await refresh(0)
    }
    finally {
      mutationLoading.value = false
    }
  }

  function requireToken(): string {
    const token = authStore.token
    if (!token)
      throw new AdminSessionError()
    return token
  }

  return {
    filters,
    result,
    isLoading,
    mutationLoading,
    error,
    refresh,
    verifyEmail,
    setRole,
    resetPassword,
    deleteUser,
  }
}

class AdminSessionError extends Error {
  readonly name = 'AdminSessionError'

  constructor() {
    super('管理会话已失效')
  }
}
