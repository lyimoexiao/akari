<script setup lang="ts">
import type { DropdownOption, MenuOption } from 'naive-ui'
import { computed, h } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const route = useRoute()
const appStore = useAppStore()
const authStore = useAuthStore()

const navigation: MenuOption[] = [
  { label: '首页', key: '/', icon: () => h('span', { class: 'i-ph-house-duotone' }) },
  { label: '皮肤库', key: '/skinlib', icon: () => h('span', { class: 'i-ph-t-shirt-duotone' }) },
]

const accountOptions = computed<DropdownOption[]>(() => {
  if (!authStore.isAuthenticated) {
    return [
      { label: '登录', key: 'login', icon: () => h('span', { class: 'i-ph-sign-in-duotone' }) },
      { label: '注册', key: 'register', icon: () => h('span', { class: 'i-ph-user-plus-duotone' }) },
    ]
  }

  return [
    { label: '账户中心', key: 'account', icon: () => h('span', { class: 'i-ph-user-circle-duotone' }) },
    ...(authStore.canManageUsers
      ? [{ label: '管理后台', key: 'admin', icon: () => h('span', { class: 'i-ph-gauge-duotone' }) }]
      : []),
    { type: 'divider', key: 'divider' },
    { label: '退出登录', key: 'logout', icon: () => h('span', { class: 'i-ph-sign-out-duotone' }) },
  ]
})

function navigate(path: string): void {
  void router.push(path)
}

async function selectAccount(key: string): Promise<void> {
  switch (key) {
    case 'login':
      await router.push({ name: 'AuthLogin' })
      return
    case 'register':
      await router.push({ name: 'AuthRegister' })
      return
    case 'account':
      await router.push({ name: 'UserHome' })
      return
    case 'admin':
      await router.push({ name: 'AdminOverview' })
      return
    case 'logout':
      await authStore.logout()
      await router.replace({ name: 'Home' })
  }
}
</script>

<template>
  <header
    class="sticky top-0 z-10 transition-all duration-200"
    :class="appStore.isScrolled ? 'bg-[var(--n-color)]/88 shadow-sm backdrop-blur-md' : 'bg-[var(--n-color)]'"
  >
    <div class="mx-auto h-16 flex items-center justify-between px-2 sm:px-4">
      <NButton text class="shrink-0 px-2" aria-label="返回首页" @click="router.push({ name: 'Home' })">
        <NBadge type="info" value="beta">
          <NText strong class="text-5 mx-2 my-2 font-minecraft">
            Akari
          </NText>
        </NBadge>
      </NButton>

      <NMenu class="min-w-0 flex-1 justify-center" :value="route.path" :options="navigation" mode="horizontal" responsive @update:value="navigate" />

      <NDropdown :options="accountOptions" trigger="click" @select="selectAccount">
        <NButton quaternary class="shrink-0">
          <template #icon>
            <span class="i-ph-user-circle-duotone" />
          </template>
          <span class="hidden sm:inline">{{ authStore.username ?? '账户' }}</span>
        </NButton>
      </NDropdown>
    </div>
  </header>
</template>
