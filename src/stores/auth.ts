import type { Serializer } from '@vueuse/core'
import type { LoginReq, RegisterReq, UserResp } from '@/types/auth'
import { useStorage } from '@vueuse/core'
import { defineStore } from 'pinia'
import { computed, shallowRef } from 'vue'
import {
  getCurrentPermissions,
  getCurrentUser,
  loginUser,
  logoutUser,
  registerUser,
} from '@/api/auth'
import { BUILT_IN_ROLES, userResponseSchema } from '@/types/auth'

const TOKEN_KEY = 'akari_auth_token'
const USER_KEY = 'akari_auth_user'

const nullableStringSerializer: Serializer<string | null> = {
  read: value => value || null,
  write: value => value ?? '',
}

const userSerializer: Serializer<UserResp | null> = {
  read(value) {
    try {
      const result = userResponseSchema.safeParse(JSON.parse(value))
      return result.success ? result.data : null
    }
    catch (error) {
      if (error instanceof SyntaxError)
        return null
      throw error
    }
  },
  write: value => JSON.stringify(value),
}

export const useAuthStore = defineStore('auth', () => {
  const token = useStorage<string | null>(TOKEN_KEY, null, localStorage, {
    serializer: nullableStringSerializer,
  })
  const user = useStorage<UserResp | null>(USER_KEY, null, localStorage, {
    serializer: userSerializer,
  })
  const permissions = shallowRef<readonly string[]>([])
  const sessionReady = shallowRef(false)

  const isAuthenticated = computed(() => token.value !== null && user.value !== null)
  const role = computed(() => user.value?.role ?? null)
  const isEmailVerified = computed(() => Boolean(user.value?.email_verified_at))
  const username = computed(() => user.value?.username ?? null)
  const email = computed(() => user.value?.email ?? null)
  const isSuperAdmin = computed(() => role.value === BUILT_IN_ROLES.SUPER_ADMIN)
  const canManageUsers = computed(() => permissions.value.includes('users.read'))
  const canManageRoles = computed(() => permissions.value.includes('roles.read'))

  function hasPermission(identifier: string): boolean {
    return permissions.value.includes(identifier)
  }

  async function loadSession(): Promise<void> {
    if (!token.value) {
      clearSession()
      sessionReady.value = true
      return
    }

    const [currentUser, permissionList] = await Promise.all([
      getCurrentUser(token.value),
      getCurrentPermissions(token.value),
    ])
    user.value = currentUser
    permissions.value = permissionList.permissions
    sessionReady.value = true
  }

  async function ensureSession(): Promise<void> {
    if (sessionReady.value)
      return

    try {
      await loadSession()
    }
    catch (error) {
      clearSession()
      sessionReady.value = true
      throw error
    }
  }

  async function login(data: LoginReq) {
    const response = await loginUser(data)
    setAuth(response.token, response.user)
    await loadSession()
    return response
  }

  async function register(data: RegisterReq) {
    const response = await registerUser(data)
    setAuth(response.token, response.user)
    await loadSession()
    return response
  }

  function setAuth(newToken: string, newUser: UserResp): void {
    token.value = newToken
    user.value = newUser
    sessionReady.value = false
  }

  function clearSession(): void {
    token.value = null
    user.value = null
    permissions.value = []
  }

  async function logout(): Promise<void> {
    const currentToken = token.value
    clearSession()
    sessionReady.value = true
    if (currentToken)
      await Promise.allSettled([logoutUser(currentToken)])
  }

  async function refreshCurrentUser() {
    if (!token.value)
      throw new ApiSessionError()

    const currentUser = await getCurrentUser(token.value)
    user.value = currentUser
    return currentUser
  }

  return {
    token,
    user,
    permissions,
    sessionReady,
    isAuthenticated,
    role,
    isEmailVerified,
    username,
    email,
    isSuperAdmin,
    canManageUsers,
    canManageRoles,
    hasPermission,
    ensureSession,
    loadSession,
    login,
    register,
    logout,
    refreshCurrentUser,
  }
})

class ApiSessionError extends Error {
  readonly name = 'ApiSessionError'

  constructor() {
    super('需要先登录')
  }
}
