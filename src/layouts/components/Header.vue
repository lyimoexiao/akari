<script setup lang="ts">
import type { DropdownOption } from 'naive-ui'
import { breakpointsTailwind, useBreakpoints } from '@vueuse/core'
import { computed, h, shallowRef } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const route = useRoute()
const appStore = useAppStore()
const authStore = useAuthStore()

const breakpoints = useBreakpoints(breakpointsTailwind)
const isDesktop = breakpoints.greaterOrEqual('sm')
const drawerOpen = shallowRef(false)

interface NavItem {
  label: string
  path: string
  icon: string
}

const navigation = computed<NavItem[]>(() => [
  { label: '首页', path: '/', icon: 'i-ph-house-duotone' },
  { label: '皮肤库', path: '/skinlib', icon: 'i-ph-t-shirt-duotone' },
  ...(authStore.isAuthenticated
    ? [{ label: '上传', path: '/user/upload', icon: 'i-ph-upload-simple-duotone' }]
    : []),
])

const accountOptions = computed<DropdownOption[]>(() => {
  if (!authStore.isAuthenticated) {
    return [
      { label: '登录', key: 'login', icon: () => h('span', { class: 'i-ph-sign-in-duotone' }) },
      { label: '注册', key: 'register', icon: () => h('span', { class: 'i-ph-user-plus-duotone' }) },
    ]
  }
  return [
    { label: '账户中心', key: 'account', icon: () => h('span', { class: 'i-ph-user-circle-duotone' }) },
    { label: '我的衣橱', key: 'closet', icon: () => h('span', { class: 'i-iconoir-closet' }) },
    ...(authStore.canManageUsers
      ? [{ label: '管理后台', key: 'admin', icon: () => h('span', { class: 'i-ph-gauge-duotone' }) }]
      : []),
    { type: 'divider', key: 'divider' },
    { label: '退出登录', key: 'logout', icon: () => h('span', { class: 'i-ph-sign-out-duotone' }) },
  ]
})

const isActive = (path: string) => route.path === path

const themeIcon = computed(() => {
  switch (appStore.themeMode) {
    case 'light': return 'i-ph-sun-duotone'
    case 'dark': return 'i-ph-moon-duotone'
    default: return 'i-ph-sun-dim-duotone'
  }
})

const themeLabel = computed(() => {
  switch (appStore.themeMode) {
    case 'light': return '明亮模式'
    case 'dark': return '暗黑模式'
    default: return '跟随系统'
  }
})

function navigate(path: string): void {
  drawerOpen.value = false
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
    case 'closet':
      await router.push({ name: 'UserCloset' })
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
    class="sticky top-0 z-10 border-b border-[var(--n-border-color)] transition-all duration-200"
    :class="appStore.isScrolled ? 'backdrop-blur-xl' : 'backdrop-blur-md'"
  >
    <div class="mx-auto flex h-16 max-w-320 items-center justify-between gap-2 px-3 sm:px-6">
      <!-- 品牌 -->
      <button
        class="group flex shrink-0 items-center gap-2 rounded-lg px-2 py-1 transition-colors hover:bg-[var(--n-hover-color)]"
        aria-label="返回首页"
        @click="router.push({ name: 'Home' })"
      >
        <svg viewBox="0 0 32 32" class="h-7 w-7 drop-shadow-sm transition-transform duration-200 group-hover:-translate-y-0.5" aria-hidden="true">
          <polygon points="16,1 31,9.5 16,18 1,9.5" fill="#8BD27C" />
          <polygon points="1,9.5 16,18 16,31 1,22.5" fill="#5CA85C" />
          <polygon points="31,9.5 16,18 16,31 31,22.5" fill="#3F8E4B" />
          <polygon points="16,18 16,31 18.5,29.75 18.5,16.75" fill="#2F7A3D" opacity="0.6" />
        </svg>
        <NBadge type="info" value="beta">
          <NText strong class="font-minecraft text-5 leading-none text-[var(--n-primary-color)]">
            Akari
          </NText>
        </NBadge>
      </button>

      <!-- 桌面导航 -->
      <nav v-if="isDesktop" class="flex min-w-0 flex-1 items-center justify-center gap-1" aria-label="主导航">
        <button
          v-for="item in navigation"
          :key="item.path"
          class="flex items-center gap-1.5 rounded-lg px-3.5 py-1.5 text-sm font-medium transition-colors"
          :class="isActive(item.path)
            ? 'bg-[var(--n-primary-color)]/12 text-[var(--n-primary-color)]'
            : 'text-[var(--n-text-color-3)] hover:bg-[var(--n-hover-color)] hover:text-[var(--n-text-color)]'"
          @click="navigate(item.path)"
        >
          <span :class="item.icon" />
          {{ item.label }}
        </button>
      </nav>

      <div class="flex shrink-0 items-center gap-1">
        <!-- 主题切换 -->
        <NTooltip trigger="hover">
          <template #trigger>
            <button
              class="rounded-lg p-2 text-[var(--n-text-color-3)] transition-colors hover:bg-[var(--n-hover-color)] hover:text-[var(--n-text-color)]"
              :aria-label="`切换主题（当前：${themeLabel}）`"
              @click="appStore.updateThemeMode()"
            >
              <span :class="themeIcon" class="text-4.5" />
            </button>
          </template>
          {{ themeLabel }}（点击切换）
        </NTooltip>

        <!-- 桌面账户菜单 -->
        <NDropdown v-if="isDesktop" :options="accountOptions" trigger="click" @select="selectAccount">
          <NButton quaternary class="shrink-0">
            <template #icon>
              <span class="i-ph-user-circle-duotone" />
            </template>
            <span class="hidden sm:inline">{{ authStore.username ?? '账户' }}</span>
          </NButton>
        </NDropdown>

        <!-- 移动端菜单按钮 -->
        <button
          v-else
          class="rounded-lg p-2 text-[var(--n-text-color-3)] transition-colors hover:bg-[var(--n-hover-color)]"
          aria-label="打开菜单"
          @click="drawerOpen = true"
        >
          <span class="i-ph-list-duotone text-4.5" />
        </button>
      </div>
    </div>

    <!-- 移动端抽屉 -->
    <NDrawer v-model:show="drawerOpen" placement="right" :width="260">
      <div class="flex h-full flex-col px-4 py-6">
        <div class="mb-6 flex items-center gap-2 px-2">
          <svg viewBox="0 0 32 32" class="h-6 w-6" aria-hidden="true">
            <polygon points="16,1 31,9.5 16,18 1,9.5" fill="#8BD27C" />
            <polygon points="1,9.5 16,18 16,31 1,22.5" fill="#5CA85C" />
            <polygon points="31,9.5 16,18 16,31 31,22.5" fill="#3F8E4B" />
          </svg>
          <NText strong class="font-minecraft text-4 text-[var(--n-primary-color)]">
            Akari
          </NText>
        </div>

        <nav class="flex flex-col gap-1" aria-label="移动端导航">
          <button
            v-for="item in navigation"
            :key="item.path"
            class="flex items-center gap-2 rounded-lg px-3 py-2.5 text-sm font-medium transition-colors"
            :class="isActive(item.path)
              ? 'bg-[var(--n-primary-color)]/12 text-[var(--n-primary-color)]'
              : 'text-[var(--n-text-color-3)] hover:bg-[var(--n-hover-color)]'"
            @click="navigate(item.path)"
          >
            <span :class="item.icon" class="text-4" />
            {{ item.label }}
          </button>
        </nav>

        <NDivider class="my-4!" />

        <div class="flex flex-col gap-1">
          <template v-if="authStore.isAuthenticated">
            <button class="flex items-center gap-2 rounded-lg px-3 py-2.5 text-sm font-medium text-[var(--n-text-color-3)] transition-colors hover:bg-[var(--n-hover-color)]" @click="selectAccount('account')">
              <span class="i-ph-user-circle-duotone text-4" />
              账户中心
            </button>
            <button class="flex items-center gap-2 rounded-lg px-3 py-2.5 text-sm font-medium text-[var(--n-text-color-3)] transition-colors hover:bg-[var(--n-hover-color)]" @click="selectAccount('closet')">
              <span class="i-iconoir-closet text-4" />
              我的衣橱
            </button>
            <button
              v-if="authStore.canManageUsers"
              class="flex items-center gap-2 rounded-lg px-3 py-2.5 text-sm font-medium text-[var(--n-text-color-3)] transition-colors hover:bg-[var(--n-hover-color)]"
              @click="selectAccount('admin')"
            >
              <span class="i-ph-gauge-duotone text-4" />
              管理后台
            </button>
            <button class="flex items-center gap-2 rounded-lg px-3 py-2.5 text-sm font-medium text-[var(--n-error-color)] transition-colors hover:bg-[var(--n-hover-color)]" @click="selectAccount('logout')">
              <span class="i-ph-sign-out-duotone text-4" />
              退出登录
            </button>
          </template>
          <template v-else>
            <button class="flex items-center gap-2 rounded-lg px-3 py-2.5 text-sm font-medium text-[var(--n-text-color-3)] transition-colors hover:bg-[var(--n-hover-color)]" @click="selectAccount('login')">
              <span class="i-ph-sign-in-duotone text-4" />
              登录
            </button>
            <button class="flex items-center gap-2 rounded-lg px-3 py-2.5 text-sm font-medium text-[var(--n-text-color-3)] transition-colors hover:bg-[var(--n-hover-color)]" @click="selectAccount('register')">
              <span class="i-ph-user-plus-duotone text-4" />
              注册
            </button>
          </template>
        </div>
      </div>
    </NDrawer>
  </header>
</template>
