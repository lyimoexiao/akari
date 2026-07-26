import type { RouteRecordRaw } from 'vue-router'

export const publicRoutes: RouteRecordRaw[] = [
  {
    path: '',
    component: () => import('@/layouts/DefaultLayout.vue'),
    children: [
      {
        path: '',
        name: 'Home',
        component: () => import('@/views/Home.vue'),
        meta: { title: '首页' },
      },
      {
        path: 'skinlib',
        name: 'Skinlib',
        component: () => import('@/views/Skinlib.vue'),
        meta: { title: '皮肤库' },
      },
    ],
  },
]
