import type { RouteRecordRaw } from 'vue-router'

export const authRoutes: RouteRecordRaw[] = [
  {
    path: '/auth',
    children: [
      {
        path: 'login',
        name: 'AuthLogin',
        component: () => import('@/views/auth/LoginPage.vue'),
        meta: { title: '登录', publicOnly: true },
      },
      {
        path: 'register',
        name: 'AuthRegister',
        component: () => import('@/views/auth/RegisterPage.vue'),
        meta: { title: '注册', publicOnly: true },
      },
      {
        path: 'verify-email',
        name: 'AuthVerifyEmail',
        component: () => import('@/views/auth/VerifyEmailPage.vue'),
        meta: { title: '验证邮箱', requiresAuth: true },
      },
      {
        path: 'forgot-password',
        name: 'AuthForgotPassword',
        component: () => import('@/views/auth/ForgotPasswordPage.vue'),
        meta: { title: '找回密码', publicOnly: true },
      },
      {
        path: 'reset-password',
        name: 'AuthResetPassword',
        component: () => import('@/views/auth/ResetPasswordPage.vue'),
        meta: { title: '重置密码', publicOnly: true },
      },
    ],
  },
]
