<script setup lang="ts">
import type { DataTableColumns } from 'naive-ui'
import type { UserItem } from '@/types/admin'
import { formatDate } from '@vueuse/core'
import { NButton, NTag } from 'naive-ui'
import { h } from 'vue'
import RoleTag from '@/components/workspace/RoleTag.vue'
import ScrollableTableRegion from '@/components/workspace/ScrollableTableRegion.vue'

const props = defineProps<{
  users: readonly UserItem[]
  loading: boolean
  canAssignRole: boolean
}>()

const emit = defineEmits<{
  verifyEmail: [user: UserItem]
  changeRole: [user: UserItem]
  resetPassword: [user: UserItem]
  deleteUser: [user: UserItem]
}>()

const columns: DataTableColumns<UserItem> = [
  { title: 'ID', key: 'id', width: 52 },
  { title: '用户名', key: 'username', width: 100 },
  { title: '邮箱', key: 'email', width: 160, ellipsis: { tooltip: true } },
  {
    title: '角色',
    key: 'role',
    width: 104,
    render: row => h(RoleTag, { role: row.role }),
  },
  {
    title: '邮箱状态',
    key: 'email_verified_at',
    width: 84,
    render: row => h(NTag, {
      type: row.email_verified_at ? 'success' : 'warning',
      size: 'small',
      bordered: false,
    }, { default: () => row.email_verified_at ? '已验证' : '未验证' }),
  },
  {
    title: '注册时间',
    key: 'created_at',
    width: 132,
    render: row => formatDate(new Date(row.created_at), 'YYYY-MM-DD HH:mm'),
  },
  {
    title: '操作',
    key: 'actions',
    width: 192,
    render(row) {
      const actions = []
      if (!row.email_verified_at) {
        actions.push(h(NButton, { size: 'tiny', quaternary: true, type: 'primary', onClick: () => emit('verifyEmail', row) }, { default: () => '验证邮箱' }))
      }
      if (props.canAssignRole) {
        actions.push(h(NButton, { size: 'tiny', quaternary: true, onClick: () => emit('changeRole', row) }, { default: () => '修改角色' }))
      }
      actions.push(h(NButton, { size: 'tiny', quaternary: true, onClick: () => emit('resetPassword', row) }, { default: () => '重置密码' }))
      actions.push(h(NButton, { size: 'tiny', quaternary: true, type: 'error', onClick: () => emit('deleteUser', row) }, { default: () => '删除' }))
      return h('div', { class: 'flex flex-wrap gap-1' }, actions)
    },
  },
]
</script>

<template>
  <ScrollableTableRegion label="用户列表，可横向滚动" :min-width="824">
    <NDataTable
      :columns="columns"
      :data="users"
      :loading="loading"
      :bordered="true"
      :single-line="false"
      size="small"
      striped
    />
  </ScrollableTableRegion>
</template>
