<script setup lang="ts">
import type { MenuOption } from 'naive-ui'
import { breakpointsTailwind, useBreakpoints } from '@vueuse/core'
import { computed, h } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import AppShell from './components/AppShell.vue'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const breakpoints = useBreakpoints(breakpointsTailwind)
const isDesktop = breakpoints.greaterOrEqual('lg')

const menuOptions = computed<MenuOption[]>(() => [
  {
    label: '概览',
    key: 'AdminOverview',
    icon: () => h('span', { class: 'i-ph-gauge-duotone' }),
  },
  {
    label: '用户',
    key: 'AdminUsers',
    icon: () => h('span', { class: 'i-ph-users-duotone' }),
  },
  ...(authStore.hasPermission('roles.read')
    ? [{ label: '角色与权限', key: 'AdminRoles', icon: () => h('span', { class: 'i-ph-shield-check-duotone' }) }]
    : []),
  ...(authStore.hasPermission('permissions.read')
    ? [{ label: '权限快照', key: 'AdminPermissions', icon: () => h('span', { class: 'i-ph-tree-structure-duotone' }) }]
    : []),
  ...(authStore.hasPermission('request-logs.read')
    ? [{ label: '请求日志', key: 'AdminRequestLogs', icon: () => h('span', { class: 'i-ph-file-magnifying-glass-duotone' }) }]
    : []),
  {
    label: '设置',
    key: 'AdminSettings',
    icon: () => h('span', { class: 'i-ph-gear-duotone' }),
  },
])

const activeKey = computed(() => typeof route.name === 'string' ? route.name : 'AdminOverview')

function navigate(name: string): void {
  void router.push({ name })
}
</script>

<template>
  <AppShell>
    <div class="mx-auto max-w-320 lg:flex">
      <aside v-if="isDesktop" class="w-56 shrink-0 border-r border-[var(--n-border-color)] py-6">
        <div class="flex items-center gap-2 px-5">
          <span class="i-ph-gauge-duotone text-4 text-[var(--n-primary-color)]" />
          <NText class="text-3 font-700 tracking-wide">
            管理后台
          </NText>
        </div>
        <div class="divider-glow mx-5 mt-4 w-12" />
        <NMenu class="mt-4" :value="activeKey" :options="menuOptions" @update:value="navigate" />
      </aside>

      <NScrollbar v-else x-scrollable class="border-b border-[var(--n-border-color)]">
        <NMenu :value="activeKey" :options="menuOptions" mode="horizontal" @update:value="navigate" />
      </NScrollbar>

      <div class="min-w-0 flex-1">
        <RouterView />
      </div>
    </div>
  </AppShell>
</template>
