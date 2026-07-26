import type { RouteRecordRaw } from 'vue-router'

export const userRoutes: RouteRecordRaw[] = [
  {
    path: '/user',
    component: () => import('@/layouts/UserLayout.vue'),
    meta: { title: '账户', requiresAuth: true },
    children: [
      {
        path: '',
        name: 'UserHome',
        component: () => import('@/views/user/Home.vue'),
        meta: { title: '账户概览', requiresAuth: true },
      },
    ],
  },
]
