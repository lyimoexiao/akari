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
      {
        path: 'closet',
        name: 'UserCloset',
        component: () => import('@/views/user/Closet.vue'),
        meta: { title: '我的衣橱', requiresAuth: true },
      },
      {
        path: 'upload',
        name: 'UserTextureUpload',
        component: () => import('@/views/user/TextureUpload.vue'),
        meta: { title: '上传纹理', requiresAuth: true },
      },
    ],
  },
]
