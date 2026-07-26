import { createRouter, createWebHistory } from 'vue-router'
import { ApiRequestError } from '@/api'
import { useAuthStore } from '@/stores/auth'
import routes from './routes'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
  scrollBehavior: () => ({ top: 0 }),
})

router.beforeEach(async (to) => {
  window.$loadingBar?.start()
  const authStore = useAuthStore()

  try {
    await authStore.ensureSession()
  }
  catch (error) {
    if (!(error instanceof ApiRequestError))
      throw error
  }

  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    return {
      name: 'AuthLogin',
      query: { redirect: to.fullPath },
    }
  }

  if (to.meta.publicOnly && authStore.isAuthenticated)
    return { name: 'UserHome' }

  const requiredPermissions = to.meta.permissions ?? []
  if (requiredPermissions.some(permission => !authStore.hasPermission(permission)))
    return { name: 'UserHome' }
})

router.afterEach(() => {
  window.$loadingBar?.finish()
})

router.onError(() => {
  window.$loadingBar?.error()
})

export default router
