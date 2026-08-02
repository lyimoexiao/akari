<script setup lang="ts">
import type { DataTableColumns } from 'naive-ui'
import type { UserItem } from '@/types/admin'
import { formatDate, useAsyncState } from '@vueuse/core'
import { computed, h } from 'vue'
import { useRouter } from 'vue-router'
import { ApiRequestError } from '@/api'
import { adminListUsers } from '@/api/admin'
import RoleTag from '@/components/workspace/RoleTag.vue'
import ScrollableTableRegion from '@/components/workspace/ScrollableTableRegion.vue'
import WorkspacePage from '@/components/workspace/WorkspacePage.vue'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()

async function loadDashboard() {
  if (!authStore.token)
    return { total: 0, recent: [] as readonly UserItem[] }
  const response = await adminListUsers(authStore.token, { page: 1, page_size: 5 })
  return { total: response.total, recent: response.items }
}

const { state, isLoading, error } = useAsyncState(loadDashboard, { total: 0, recent: [] as readonly UserItem[] })
const errorMessage = computed(() => error.value instanceof ApiRequestError
  ? error.value.message
  : error.value ? '管理概览加载失败' : null)

const columns: DataTableColumns<UserItem> = [
  { title: '用户名', key: 'username', minWidth: 120 },
  { title: '邮箱', key: 'email', minWidth: 200, ellipsis: { tooltip: true } },
  { title: '角色', key: 'role', width: 128, render: user => h(RoleTag, { role: user.role }) },
  { title: '注册时间', key: 'created_at', width: 164, render: user => formatDate(new Date(user.created_at), 'YYYY-MM-DD HH:mm') },
]
</script>

<template>
  <WorkspacePage title="管理概览" description="查看账户规模、权限范围和最近注册">
    <NAlert v-if="errorMessage" type="error" class="mb-4">
      {{ errorMessage }}
    </NAlert>

    <div class="grid grid-cols-2 gap-4 lg:grid-cols-4">
      <NCard size="small">
        <NFlex align="center" :size="12">
          <div class="rounded-xl bg-[var(--n-primary-color)]/12 p-3">
            <span class="i-ph-users-duotone text-4.5 text-[var(--n-primary-color)]" />
          </div>
          <NStatistic label="总用户数" :value="isLoading ? '-' : state.total" />
        </NFlex>
      </NCard>
      <NCard size="small">
        <NFlex align="center" :size="12">
          <div class="rounded-xl bg-[var(--n-primary-color)]/12 p-3">
            <span class="i-ph-key-duotone text-4.5 text-[var(--n-primary-color)]" />
          </div>
          <NStatistic label="当前权限" :value="authStore.permissions.length" />
        </NFlex>
      </NCard>
      <NCard size="small">
        <NFlex align="center" :size="12">
          <div class="rounded-xl bg-[var(--n-primary-color)]/12 p-3">
            <span class="i-ph-user-gear-duotone text-4.5 text-[var(--n-primary-color)]" />
          </div>
          <NStatistic label="用户管理" :value="authStore.hasPermission('users.read') ? '可用' : '不可用'" />
        </NFlex>
      </NCard>
      <NCard size="small">
        <NFlex align="center" :size="12">
          <div class="rounded-xl bg-[var(--n-primary-color)]/12 p-3">
            <span class="i-ph-shield-check-duotone text-4.5 text-[var(--n-primary-color)]" />
          </div>
          <NStatistic label="角色管理" :value="authStore.hasPermission('roles.read') ? '可用' : '不可用'" />
        </NFlex>
      </NCard>
    </div>

    <NCard title="快捷操作" size="small" class="mt-4">
      <NFlex :wrap="true">
        <NButton @click="router.push({ name: 'AdminUsers' })">
          用户管理
        </NButton>
        <NButton v-if="authStore.hasPermission('roles.read')" @click="router.push({ name: 'AdminRoles' })">
          角色与权限
        </NButton>
        <NButton v-if="authStore.hasPermission('request-logs.read')" @click="router.push({ name: 'AdminRequestLogs' })">
          请求日志
        </NButton>
      </NFlex>
    </NCard>

    <NCard title="最近注册" size="small" class="mt-4">
      <ScrollableTableRegion label="最近注册用户列表，可横向滚动" :min-width="612">
        <NDataTable :columns="columns" :data="state.recent" :loading="isLoading" :bordered="false" />
      </ScrollableTableRegion>
    </NCard>
  </WorkspacePage>
</template>
