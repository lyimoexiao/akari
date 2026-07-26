<script setup lang="ts">
import type { UserItem } from '@/types/admin'
import { useAsyncState, useDebounceFn, useToggle } from '@vueuse/core'
import { useDialog, useMessage } from 'naive-ui'
import { computed, shallowRef } from 'vue'
import { ApiRequestError } from '@/api'
import { adminListRoles } from '@/api/admin'
import AdminUsersTable from '@/components/admin/AdminUsersTable.vue'
import UserActionDialogs from '@/components/admin/UserActionDialogs.vue'
import WorkspacePage from '@/components/workspace/WorkspacePage.vue'
import { useAdminUsers } from '@/composables/useAdminUsers'
import { useAuthStore } from '@/stores/auth'

const message = useMessage()
const dialog = useDialog()
const authStore = useAuthStore()
const users = useAdminUsers()
const selectedUser = shallowRef<UserItem | null>(null)
const selectedRole = shallowRef('')
const newPassword = shallowRef('')
const [roleOpen, toggleRole] = useToggle(false)
const [passwordOpen, togglePassword] = useToggle(false)

const { state: roles } = useAsyncState(async () => {
  if (!authStore.token || !authStore.hasPermission('roles.read'))
    return []
  const response = await adminListRoles(authStore.token)
  return response.items.map(role => role.name)
}, [])

const errorMessage = computed(() => users.error.value instanceof ApiRequestError
  ? users.error.value.message
  : users.error.value ? '获取用户列表失败' : null)

const search = useDebounceFn(() => {
  users.filters.page = 1
  void users.refresh(0)
}, 300)

function openRole(user: UserItem): void {
  selectedUser.value = user
  selectedRole.value = user.role
  toggleRole(true)
}

function openPassword(user: UserItem): void {
  selectedUser.value = user
  newPassword.value = ''
  togglePassword(true)
}

async function saveRole(): Promise<void> {
  if (!selectedUser.value)
    return
  try {
    await users.setRole(selectedUser.value.id, selectedRole.value)
    message.success('角色已更新')
    toggleRole(false)
  }
  catch (error) {
    message.error(error instanceof ApiRequestError ? error.message : '角色更新失败')
  }
}

async function savePassword(): Promise<void> {
  if (!selectedUser.value)
    return
  try {
    await users.resetPassword(selectedUser.value.id, newPassword.value)
    message.success('密码已重置')
    togglePassword(false)
  }
  catch (error) {
    message.error(error instanceof ApiRequestError ? error.message : '密码重置失败')
  }
}

function verifyEmail(user: UserItem): void {
  dialog.warning({
    title: '确认验证邮箱',
    content: `确定将 ${user.username} 的邮箱标记为已验证吗？`,
    positiveText: '确认',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await users.verifyEmail(user.id)
        message.success('邮箱已验证')
      }
      catch (error) {
        message.error(error instanceof ApiRequestError ? error.message : '邮箱验证失败')
      }
    },
  })
}

function deleteUser(user: UserItem): void {
  dialog.warning({
    title: '确认删除用户',
    content: `确定删除 ${user.username}（${user.email}）吗？此操作不可撤销。`,
    positiveText: '删除',
    negativeText: '取消',
    positiveButtonProps: { type: 'error' },
    onPositiveClick: async () => {
      try {
        await users.deleteUser(user.id)
        message.success('用户已删除')
      }
      catch (error) {
        message.error(error instanceof ApiRequestError ? error.message : '删除用户失败')
      }
    },
  })
}
</script>

<template>
  <WorkspacePage title="用户管理" description="查询账户并执行权限范围内的管理操作">
    <template #actions>
      <NText depth="3">
        共 {{ users.result.value.total }} 人
      </NText>
    </template>

    <NAlert v-if="errorMessage" type="error" class="mb-4">
      {{ errorMessage }}
    </NAlert>

    <NFlex class="mb-4" :wrap="true">
      <NInput
        v-model:value="users.filters.query"
        class="w-full sm:max-w-75"
        clearable
        placeholder="搜索用户名或邮箱"
        aria-label="搜索用户名或邮箱"
        @update:value="search"
        @keyup.enter="users.refresh(0)"
      />
      <NButton type="primary" @click="users.refresh(0)">
        搜索
      </NButton>
      <NButton @click="users.filters.query = ''; search()">
        重置
      </NButton>
    </NFlex>

    <AdminUsersTable
      :users="users.result.value.items"
      :loading="users.isLoading.value"
      :can-assign-role="authStore.hasPermission('roles.assign')"
      @verify-email="verifyEmail"
      @change-role="openRole"
      @reset-password="openPassword"
      @delete-user="deleteUser"
    />

    <NFlex class="mt-4" justify="end">
      <NPagination
        v-model:page="users.filters.page"
        :page-size="users.filters.pageSize"
        :item-count="users.result.value.total"
        @update:page="users.refresh(0)"
      />
    </NFlex>

    <UserActionDialogs
      v-model:role-open="roleOpen"
      v-model:password-open="passwordOpen"
      v-model:role="selectedRole"
      v-model:password="newPassword"
      :user="selectedUser"
      :roles="roles"
      :loading="users.mutationLoading.value"
      @save-role="saveRole"
      @save-password="savePassword"
    />
  </WorkspacePage>
</template>
