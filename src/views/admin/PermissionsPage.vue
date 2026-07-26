<script setup lang="ts">
import type { DataTableColumns } from 'naive-ui'
import type { PermissionSnapshot } from '@/types/admin'
import { useAsyncState } from '@vueuse/core'
import { computed } from 'vue'
import { ApiRequestError } from '@/api'
import { adminGetPermissionSnapshot } from '@/api/admin'
import ScrollableTableRegion from '@/components/workspace/ScrollableTableRegion.vue'
import WorkspacePage from '@/components/workspace/WorkspacePage.vue'
import { useAuthStore } from '@/stores/auth'

type PermissionRule = PermissionSnapshot['rules'][number]
type RoleInheritance = PermissionSnapshot['inheritance'][number]

const authStore = useAuthStore()

async function loadSnapshot(): Promise<PermissionSnapshot> {
  if (!authStore.token)
    return { roles: [], rules: [], inheritance: [] }
  return adminGetPermissionSnapshot(authStore.token)
}

const { state, isLoading, error } = useAsyncState(loadSnapshot, { roles: [], rules: [], inheritance: [] })
const errorMessage = computed(() => error.value instanceof ApiRequestError
  ? error.value.message
  : error.value ? '获取权限快照失败' : null)

const ruleColumns: DataTableColumns<PermissionRule> = [
  { title: '角色', key: 'role', width: 148 },
  { title: '方法', key: 'action', width: 96 },
  { title: '资源', key: 'object', minWidth: 280 },
]

const inheritanceColumns: DataTableColumns<RoleInheritance> = [
  { title: '角色', key: 'role', minWidth: 148 },
  { title: '继承自', key: 'parent', minWidth: 148 },
]
</script>

<template>
  <WorkspacePage title="权限快照" description="查看后端当前生效的角色继承与路由规则">
    <NAlert v-if="errorMessage" type="error" class="mb-4">
      {{ errorMessage }}
    </NAlert>

    <div class="grid gap-4 lg:grid-cols-[minmax(0,2fr)_minmax(240px,1fr)]">
      <NCard title="访问规则" size="small">
        <ScrollableTableRegion label="访问规则列表，可横向滚动" :min-width="524">
          <NDataTable :columns="ruleColumns" :data="state.rules" :loading="isLoading" :bordered="false" />
        </ScrollableTableRegion>
      </NCard>
      <NCard title="角色继承" size="small">
        <NDataTable :columns="inheritanceColumns" :data="state.inheritance" :loading="isLoading" :bordered="false" />
        <NDivider />
        <NFlex :wrap="true">
          <NTag v-for="role in state.roles" :key="role" size="small" :bordered="false">
            {{ role }}
          </NTag>
        </NFlex>
      </NCard>
    </div>
  </WorkspacePage>
</template>
