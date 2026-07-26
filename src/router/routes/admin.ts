import type { RouteRecordRaw } from 'vue-router'

export const adminRoutes: RouteRecordRaw[] = [
  {
    path: '/admin',
    component: () => import('@/layouts/AdminLayout.vue'),
    meta: { title: '管理后台', requiresAuth: true, permissions: ['users.read'] },
    children: [
      {
        path: '',
        name: 'AdminOverview',
        component: () => import('@/views/admin/DashboardPage.vue'),
        meta: { title: '管理概览', requiresAuth: true, permissions: ['users.read'] },
      },
      {
        path: 'users',
        name: 'AdminUsers',
        component: () => import('@/views/admin/UsersPage.vue'),
        meta: { title: '用户管理', requiresAuth: true, permissions: ['users.read'] },
      },
      {
        path: 'roles',
        name: 'AdminRoles',
        component: () => import('@/views/admin/RolesPage.vue'),
        meta: { title: '角色与权限', requiresAuth: true, permissions: ['roles.read'] },
      },
      {
        path: 'permissions',
        name: 'AdminPermissions',
        component: () => import('@/views/admin/PermissionsPage.vue'),
        meta: { title: '权限快照', requiresAuth: true, permissions: ['permissions.read'] },
      },
      {
        path: 'request-logs',
        name: 'AdminRequestLogs',
        component: () => import('@/views/admin/RequestLogsPage.vue'),
        meta: { title: '请求日志', requiresAuth: true, permissions: ['request-logs.read'] },
      },
      {
        path: 'settings',
        name: 'AdminSettings',
        component: () => import('@/views/admin/SettingsPage.vue'),
        meta: { title: '管理设置', requiresAuth: true, permissions: ['users.read'] },
      },
    ],
  },
]
