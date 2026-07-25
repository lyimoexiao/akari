<script setup lang="ts">
import type { UserItem } from '@/types/admin'
import { NButton, NTag, useDialog, useMessage } from 'naive-ui'
import { computed, h, onMounted, ref } from 'vue'
import { adminDeleteUser, adminListUsers, adminResetPassword, adminSetRole, adminVerifyEmail } from '@/api/admin'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const message = useMessage()
const dialog = useDialog()

const users = ref<UserItem[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const totalPages = ref(0)
const query = ref('')
const loading = ref(false)

const passwordModal = ref(false)
const passwordUserId = ref(0)
const newPassword = ref('')
const passwordLoading = ref(false)

const roleModal = ref(false)
const roleUserId = ref(0)
const roleUserName = ref('')
const newRole = ref<'super_admin' | 'staff' | 'user'>('user')
const roleLoading = ref(false)

async function fetchUsers() {
  loading.value = true
  try {
    const res = await adminListUsers(authStore.token!, {
      page: page.value,
      page_size: pageSize.value,
      query: query.value || undefined,
    })
    users.value = res.items
    total.value = res.total
    page.value = res.page
    totalPages.value = res.total_pages
  }
  catch (e: any) {
    message.error(e.message || '获取用户列表失败')
  }
  loading.value = false
}

function onSearch() {
  page.value = 1
  fetchUsers()
}

function handleVerifyEmail(user: UserItem) {
  if (!authStore.token)
    return
  dialog.warning({
    title: '确认',
    content: `手动验证用户 ${user.username} 的邮箱？`,
    positiveText: '确认',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await adminVerifyEmail(authStore.token!, user.id)
        message.success('邮箱已验证')
        fetchUsers()
      }
      catch (e: any) {
        message.error(e.message || '验证失败')
      }
    },
  })
}

function openRoleModal(user: UserItem) {
  roleUserId.value = user.id
  roleUserName.value = user.username
  newRole.value = user.role
  roleModal.value = true
}

async function handleSetRole() {
  if (!authStore.token)
    return
  roleLoading.value = true
  try {
    await adminSetRole(authStore.token!, roleUserId.value, newRole.value)
    message.success('角色已更新')
    roleModal.value = false
    fetchUsers()
  }
  catch (e: any) {
    message.error(e.message || '设置角色失败')
  }
  roleLoading.value = false
}

function openPasswordModal(user: UserItem) {
  passwordUserId.value = user.id
  newPassword.value = ''
  passwordModal.value = true
}

async function handleResetPassword() {
  if (!authStore.token)
    return
  passwordLoading.value = true
  try {
    await adminResetPassword(authStore.token!, passwordUserId.value, newPassword.value)
    message.success('密码已重置')
    passwordModal.value = false
  }
  catch (e: any) {
    message.error(e.message || '重置密码失败')
  }
  passwordLoading.value = false
}

function handleDeleteUser(user: UserItem) {
  if (!authStore.token)
    return
  dialog.warning({
    title: '确认删除',
    content: `确定要删除用户 ${user.username}（${user.email}）？此操作不可撤销。`,
    positiveText: '删除',
    negativeText: '取消',
    positiveButtonProps: { type: 'error' as const },
    onPositiveClick: async () => {
      try {
        await adminDeleteUser(authStore.token!, user.id)
        message.success('用户已删除')
        fetchUsers()
      }
      catch (e: any) {
        message.error(e.message || '删除失败')
      }
    },
  })
}

const columns = computed(() => [
  { title: 'ID', key: 'id', width: 70 },
  { title: '用户名', key: 'username', width: 140 },
  { title: '邮箱', key: 'email', width: 240 },
  {
    title: '角色',
    key: 'role',
    width: 120,
    render(row: UserItem) {
      return h(NTag, {
        type: row.role === 'super_admin' ? 'warning' : row.role === 'staff' ? 'info' : 'default',
        size: 'small',
        bordered: false,
      }, { default: () => row.role })
    },
  },
  {
    title: '邮箱验证',
    key: 'email_verified_at',
    width: 100,
    render(row: UserItem) {
      return h(NTag, {
        type: row.email_verified_at ? 'success' : 'warning',
        size: 'small',
        bordered: false,
      }, { default: () => row.email_verified_at ? '已验证' : '未验证' })
    },
  },
  {
    title: '注册时间',
    key: 'created_at',
    width: 180,
    render(row: UserItem) {
      return new Date(row.created_at).toLocaleString()
    },
  },
  {
    title: '操作',
    key: 'actions',
    width: 320,
    render(row: UserItem) {
      const btns: ReturnType<typeof h>[] = []
      if (!row.email_verified_at) {
        btns.push(h(NButton, { size: 'tiny', quaternary: true, type: 'primary', onClick: () => handleVerifyEmail(row) }, { default: () => '验证邮箱' }))
      }
      if (authStore.isSuperAdmin) {
        btns.push(h(NButton, { size: 'tiny', quaternary: true, type: 'primary', onClick: () => openRoleModal(row) }, { default: () => '角色' }))
      }
      btns.push(h(NButton, { size: 'tiny', quaternary: true, type: 'primary', onClick: () => openPasswordModal(row) }, { default: () => '重置密码' }))
      btns.push(h(NButton, { size: 'tiny', quaternary: true, type: 'error', onClick: () => handleDeleteUser(row) }, { default: () => '删除' }))
      return h('div', { style: 'display: flex; gap: 4px; flex-wrap: wrap' }, btns)
    },
  },
])

onMounted(fetchUsers)
</script>

<template>
  <div class="p-6">
    <div class="flex items-center justify-between mb-4">
      <h1 class="text-lg font-semibold">
        用户管理
      </h1>
      <div class="flex items-center gap-2 text-sm op-fade">
        <span>共 {{ total }} 人</span>
      </div>
    </div>

    <div class="flex items-center gap-3 mb-4">
      <n-input v-model:value="query" placeholder="搜索用户名或邮箱" clearable style="max-width: 300px" @keyup.enter="onSearch" />
      <NButton size="small" type="primary" @click="onSearch">
        <template #icon>
          <span class="i-ph-magnifying-glass-duotone" />
        </template>
        搜索
      </NButton>
      <NButton size="small" @click="query = ''; onSearch()">
        <template #icon>
          <span class="i-ph-arrow-counter-clockwise-duotone" />
        </template>
        重置
      </NButton>
    </div>

    <n-data-table
      :columns="columns"
      :data="users"
      :loading="loading"
      :bordered="true"
      :single-line="false"
      size="small"
      striped
    />

    <div class="flex justify-end mt-4">
      <n-pagination
        v-model:page="page"
        :page-size="pageSize"
        :item-count="total"
        :page-slot="7"
        @update:page="fetchUsers"
      />
    </div>

    <n-modal v-model:show="roleModal" title="设置角色" preset="card" style="width: 400px">
      <n-space vertical>
        <n-text>用户：{{ roleUserName }}</n-text>
        <n-radio-group v-model:value="newRole" name="role">
          <n-space vertical>
            <n-radio value="user">
              用户
            </n-radio>
            <n-radio value="staff">
              Staff
            </n-radio>
            <n-radio value="super_admin">
              超级管理员
            </n-radio>
          </n-space>
        </n-radio-group>
        <NButton type="primary" :loading="roleLoading" @click="handleSetRole">
          确认
        </NButton>
      </n-space>
    </n-modal>

    <n-modal v-model:show="passwordModal" title="重置密码" preset="card" style="width: 400px">
      <n-space vertical>
        <n-input v-model:value="newPassword" type="password" placeholder="新密码（至少6位）" show-password-on="click" />
        <NButton type="primary" :loading="passwordLoading" @click="handleResetPassword">
          确认重置
        </NButton>
      </n-space>
    </n-modal>
  </div>
</template>
