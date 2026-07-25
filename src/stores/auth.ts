import type { LoginReq, RegisterReq, UserResp } from '@/types/auth'
import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { getCurrentUser, loginUser, logoutUser, registerUser } from '@/api/auth'

const TOKEN_KEY = 'akari_auth_token'
const USER_KEY = 'akari_auth_user'

export const useAuthStore = defineStore('auth', () => {
  // State
  const token = ref<string | null>(localStorage.getItem(TOKEN_KEY))
  const user = ref<UserResp | null>(loadUserFromStorage())

  // Getters
  const isAuthenticated = computed(() => !!token.value && !!user.value)
  const role = computed(() => user.value?.role ?? null)
  const isSuperAdmin = computed(() => role.value === 'super_admin')
  const isStaff = computed(() => role.value === 'staff' || role.value === 'super_admin')
  const isEmailVerified = computed(() => !!user.value?.email_verified_at)
  const username = computed(() => user.value?.username ?? '')
  const email = computed(() => user.value?.email ?? '')

  // Actions
  async function login(data: LoginReq) {
    const response = await loginUser(data)
    setAuth(response.token, response.user)
    return response
  }

  async function register(data: RegisterReq) {
    const response = await registerUser(data)
    setAuth(response.token, response.user)
    return response
  }

  function setAuth(newToken: string, newUser: UserResp) {
    token.value = newToken
    user.value = newUser
    localStorage.setItem(TOKEN_KEY, newToken)
    localStorage.setItem(USER_KEY, JSON.stringify(newUser))
  }

  async function logout() {
    const currentToken = token.value
    try {
      if (currentToken) {
        await logoutUser(currentToken)
      }
    }
    catch {
      // Ignore error (token might be invalid, still clear local state)
    }
    token.value = null
    user.value = null
    localStorage.removeItem(TOKEN_KEY)
    localStorage.removeItem(USER_KEY)
  }

  async function refreshCurrentUser() {
    if (!token.value) {
      throw new Error('Not authenticated')
    }
    const currentUser = await getCurrentUser(token.value)
    user.value = currentUser
    localStorage.setItem(USER_KEY, JSON.stringify(currentUser))
    return currentUser
  }

  function loadUserFromStorage(): UserResp | null {
    try {
      const stored = localStorage.getItem(USER_KEY)
      if (stored) {
        return JSON.parse(stored) as UserResp
      }
    }
    catch {
      // Invalid stored data
      localStorage.removeItem(USER_KEY)
    }
    return null
  }

  return {
    // State
    token,
    user,
    // Getters
    isAuthenticated,
    role,
    isSuperAdmin,
    isStaff,
    isEmailVerified,
    username,
    email,
    // Actions
    login,
    register,
    setAuth,
    logout,
    refreshCurrentUser,
  }
})
