import type { RouteRecordRaw } from 'vue-router'
import { adminRoutes } from './admin'
import { authRoutes } from './auth'
import { publicRoutes } from './public'
import { userRoutes } from './user'

const routes: RouteRecordRaw[] = [
  ...publicRoutes,
  ...authRoutes,
  ...userRoutes,
  ...adminRoutes,
]

export default routes
