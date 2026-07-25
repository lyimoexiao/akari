<script setup lang="ts">
import type { MenuOption } from 'naive-ui'
import { computed, h } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const router = useRouter()
const route = useRoute()

function renderIcon(icon: string) {
  return () => h('span', { class: icon })
}

const menuOptions: MenuOption[] = [
  {
    label: '概览',
    key: 'overview',
    icon: renderIcon('i-ph-gauge-duotone'),
  },
  {
    label: '用户管理',
    key: 'admin-users',
    icon: renderIcon('i-ph-users-duotone'),
  },
  {
    label: '系统设置',
    key: 'admin-settings',
    icon: renderIcon('i-ph-gear-duotone'),
  },
]

const activeKey = computed(() => route.name as string)

function handleMenuUpdate(key: string) {
  router.push({ name: key })
}

function handleLogout() {
  authStore.logout().then(() => router.push('/login'))
}
</script>

<template>
  <n-layout position="absolute" has-sider>
    <n-layout-sider
      bordered
      inverted
      collapse-mode="width"
      :collapsed-width="56"
      :width="224"
      :native-scrollbar="false"
      show-trigger="bar"
    >
      <div class="h-12 flex items-center px-4 gap-2 border-b border-#8882">
        <span class="i-ph-lightning-duotone text-lg text-blue-300 shrink-0" />
        <span class="text-sm font-semibold truncate text-white">Akari Admin</span>
      </div>

      <n-menu
        :value="activeKey"
        :options="menuOptions"
        :collapsed-width="56"
        :collapsed-icon-size="20"
        class="mt-1"
        @update:value="handleMenuUpdate"
      />
    </n-layout-sider>

    <n-layout>
      <n-layout-header bordered class="h-12 flex items-center justify-between px-4">
        <div class="flex items-center gap-3">
          <span class="text-sm op-fade">
            {{ menuOptions.find(m => m.key === activeKey)?.label ?? '管理后台' }}
          </span>
        </div>

        <div class="flex items-center gap-3">
          <n-tag
            v-if="authStore.role"
            :type="authStore.role === 'super_admin' ? 'warning' : 'info'"
            size="small"
            :bordered="false"
          >
            {{ authStore.role === 'super_admin' ? '超级管理员' : authStore.role === 'staff' ? 'Staff' : '用户' }}
          </n-tag>
          <span class="text-sm op-fade hidden sm:block">{{ authStore.username }}</span>
          <n-button size="tiny" quaternary @click="handleLogout">
            退出
          </n-button>
        </div>
      </n-layout-header>

      <n-layout-content>
        <router-view />
      </n-layout-content>
    </n-layout>
  </n-layout>
</template>
