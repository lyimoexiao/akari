<script setup lang="ts">
import type { DataTableColumns } from 'naive-ui'
import type { RoleItem, SaveRoleReq } from '@/types/admin'
import { useAsyncState, useToggle } from '@vueuse/core'
import { NButton, NTag, useDialog, useMessage } from 'naive-ui'
import { computed, h, shallowRef } from 'vue'
import { ApiRequestError } from '@/api'
import {
  adminCreateRole,
  adminDeleteRole,
  adminListRoles,
  adminSetDefaultRole,
  adminUpdateRole,
} from '@/api/admin'
import RoleEditorDialog from '@/components/admin/RoleEditorDialog.vue'
import ScrollableTableRegion from '@/components/workspace/ScrollableTableRegion.vue'
import WorkspacePage from '@/components/workspace/WorkspacePage.vue'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const message = useMessage()
const dialog = useDialog()
const selectedRole = shallowRef<RoleItem | null>(null)
const mutationLoading = shallowRef(false)
const [editorOpen, toggleEditor] = useToggle(false)

async function loadRoles(): Promise<readonly RoleItem[]> {
  if (!authStore.token)
    return []
  return (await adminListRoles(authStore.token)).items
}

const { state: roles, isLoading, error, execute: refresh } = useAsyncState(loadRoles, [])
const availablePermissions = computed(() => [...new Set(roles.value.flatMap(role => role.permissions))].sort())
const errorMessage = computed(() => error.value instanceof ApiRequestError
  ? error.value.message
  : error.value ? '获取角色列表失败' : null)

function openCreate(): void {
  selectedRole.value = null
  toggleEditor(true)
}

function openEdit(role: RoleItem): void {
  selectedRole.value = role
  toggleEditor(true)
}

async function saveRole(data: SaveRoleReq): Promise<void> {
  if (!authStore.token)
    return
  mutationLoading.value = true
  try {
    if (selectedRole.value)
      await adminUpdateRole(authStore.token, selectedRole.value.id, data)
    else
      await adminCreateRole(authStore.token, data)
    message.success(selectedRole.value ? '角色已更新' : '角色已创建')
    toggleEditor(false)
    await refresh(0)
  }
  catch (saveError) {
    message.error(saveError instanceof ApiRequestError ? saveError.message : '保存角色失败')
  }
  finally {
    mutationLoading.value = false
  }
}

function setDefault(role: RoleItem): void {
  dialog.info({
    title: '设置默认角色',
    content: `新注册用户将自动获得角色 ${role.name}，是否继续？`,
    positiveText: '设置',
    negativeText: '取消',
    onPositiveClick: async () => {
      if (!authStore.token)
        return
      try {
        await adminSetDefaultRole(authStore.token, role.id)
        message.success('默认注册角色已更新')
        await refresh(0)
      }
      catch (setError) {
        message.error(setError instanceof ApiRequestError ? setError.message : '设置默认角色失败')
      }
    },
  })
}

function removeRole(role: RoleItem): void {
  dialog.warning({
    title: '删除角色',
    content: `确定删除角色 ${role.name} 吗？`,
    positiveText: '删除',
    negativeText: '取消',
    positiveButtonProps: { type: 'error' },
    onPositiveClick: async () => {
      if (!authStore.token)
        return
      try {
        await adminDeleteRole(authStore.token, role.id)
        message.success('角色已删除')
        await refresh(0)
      }
      catch (deleteError) {
        message.error(deleteError instanceof ApiRequestError ? deleteError.message : '删除角色失败')
      }
    },
  })
}

const columns: DataTableColumns<RoleItem> = [
  { title: '名称', key: 'name', minWidth: 132 },
  { title: '说明', key: 'description', minWidth: 180, ellipsis: { tooltip: true } },
  {
    title: '默认',
    key: 'is_default',
    width: 84,
    render: role => h(NTag, { type: role.is_default ? 'success' : 'default', size: 'small', bordered: false }, { default: () => role.is_default ? '是' : '否' }),
  },
  {
    title: '权限',
    key: 'permissions',
    minWidth: 220,
    render: role => h('span', { class: 'text-sm op-fade' }, `${role.permissions.length} 项`),
  },
  {
    title: '操作',
    key: 'actions',
    width: 224,
    render: role => h('div', { class: 'flex gap-1' }, [
      h(NButton, { size: 'tiny', quaternary: true, onClick: () => openEdit(role) }, { default: () => '编辑' }),
      h(NButton, { size: 'tiny', quaternary: true, disabled: role.is_default, onClick: () => setDefault(role) }, { default: () => '设为默认' }),
      h(NButton, { size: 'tiny', quaternary: true, type: 'error', disabled: role.is_default || role.name === 'super_admin', onClick: () => removeRole(role) }, { default: () => '删除' }),
    ]),
  },
]
</script>

<template>
  <WorkspacePage title="角色与权限" description="管理角色、默认注册角色及其权限集合">
    <template #actions>
      <NButton type="primary" @click="openCreate">
        创建角色
      </NButton>
    </template>
    <NAlert v-if="errorMessage" type="error" class="mb-4">
      {{ errorMessage }}
    </NAlert>
    <ScrollableTableRegion label="角色列表，可横向滚动" :min-width="840">
      <NDataTable :columns="columns" :data="roles" :loading="isLoading" />
    </ScrollableTableRegion>
    <RoleEditorDialog
      v-model="editorOpen"
      :role="selectedRole"
      :permissions="availablePermissions"
      :loading="mutationLoading"
      @submit="saveRole"
    />
  </WorkspacePage>
</template>
