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
    <div class="mx-auto max-w-280 lg:flex">
      <aside v-if="isDesktop" class="w-56 shrink-0 border-r border-[var(--n-border-color)] py-6">
        <NText depth="3" class="px-5 text-3 font-600 tracking-wide">
          管理后台
        </NText>
        <NMenu class="mt-3" :value="activeKey" :options="menuOptions" @update:value="navigate" />
      </aside>

      <NScrollbar v-else x-scrollable class="border-b border-[var(--n-border-color)]">
        <NMenu :value="activeKey" :options="menuOptions" mode="horizontal" @update:value="navigate" />
      </NScrollbar>

      <!-- <div v-else class="flex items-center gap-3 border-b border-[var(--n-border-color)] px-4 py-3">
        <NText class="shrink-0 text-3 font-600">
          管理页面
        </NText>
        <NSelect
          aria-label="选择管理页面"
          :value="activeKey"
          :options="selectOptions"
          @update:value="navigate"
        />
      </div> -->

      <div class="min-w-0 flex-1">
        <RouterView />
      </div>
    </div>
  </AppShell>
</template>
