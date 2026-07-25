<script setup lang="ts">
import type { UserItem } from '@/types/admin'
import { NButton, NCard, NDataTable, NStatistic, NTag, useMessage } from 'naive-ui'
import { computed, h, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { adminListUsers } from '@/api/admin'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const router = useRouter()
const message = useMessage()

const totalUsers = ref(0)
const recentUsers = ref<UserItem[]>([])
const loading = ref(false)

const verifiedCount = computed(() => recentUsers.value.filter(u => u.email_verified_at).length)
const adminCount = computed(() => recentUsers.value.filter(u => u.role === 'super_admin' || u.role === 'staff').length)

const recentColumns = [
  { title: '用户名', key: 'username', minWidth: 100 },
  { title: '邮箱', key: 'email', ellipsis: { tooltip: true }, minWidth: 160 },
  {
    title: '角色',
    key: 'role',
    width: 100,
    render(row: UserItem) {
      return h(NTag, {
        type: row.role === 'super_admin' ? 'warning' : row.role === 'staff' ? 'info' : 'default',
        size: 'small',
        bordered: false,
      }, { default: () => row.role })
    },
  },
  {
    title: '注册时间',
    key: 'created_at',
    minWidth: 150,
    render(row: UserItem) {
      return new Date(row.created_at).toLocaleString()
    },
  },
]

onMounted(async () => {
  if (!authStore.token)
    return
  loading.value = true
  try {
    const res = await adminListUsers(authStore.token, { page: 1, page_size: 1 })
    totalUsers.value = res.total
    const recent = await adminListUsers(authStore.token, { page: 1, page_size: 5 })
    recentUsers.value = recent.items
  }
  catch (e: any) {
    message.error(e.message || '获取数据失败')
  }
  loading.value = false
})
</script>

<template>
  <div class="p-6">
    <h2 class="text-lg font-semibold mb-5">
      概览
    </h2>

    <!-- Stat cards: 4 cols desktop, 2 tablet, 2 mobile -->
    <div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
      <NCard :bordered="true" size="small">
        <NStatistic label="总用户数" :value="loading ? '-' : totalUsers" tabular-nums>
          <template #prefix>
            <span class="i-ph-users-duotone text-lg op-fade mr-1" />
          </template>
        </NStatistic>
      </NCard>
      <NCard :bordered="true" size="small">
        <NStatistic label="已验证邮箱" :value="loading ? '-' : verifiedCount" tabular-nums>
          <template #prefix>
            <span class="i-ph-check-circle-duotone text-lg op-fade mr-1" />
          </template>
        </NStatistic>
      </NCard>
      <NCard :bordered="true" size="small">
        <NStatistic label="管理员/Staff" :value="loading ? '-' : adminCount" tabular-nums>
          <template #prefix>
            <span class="i-ph-shield-star-duotone text-lg op-fade mr-1" />
          </template>
        </NStatistic>
      </NCard>
      <NCard :bordered="true" size="small">
        <NStatistic label="服务器状态" value="正常" tabular-nums>
          <template #prefix>
            <span class="i-ph-lightning-duotone text-lg color-active mr-1" />
          </template>
        </NStatistic>
      </NCard>
    </div>

    <NCard title="快捷操作" :bordered="true" size="small" class="mt-4">
      <div class="flex flex-wrap gap-3">
        <NButton size="small" @click="router.push('/admin/users')">
          <template #icon>
            <span class="i-ph-users-duotone" />
          </template>
          用户管理
        </NButton>
        <NButton size="small" @click="router.push('/admin/settings')">
          <template #icon>
            <span class="i-ph-gear-duotone" />
          </template>
          系统设置
        </NButton>
        <NButton size="small" @click="router.push('/')">
          <template #icon>
            <span class="i-ph-house-duotone" />
          </template>
          返回首页
        </NButton>
      </div>
    </NCard>

    <NCard title="最近注册" :bordered="true" size="small" class="mt-4">
      <NDataTable
        :columns="recentColumns"
        :data="recentUsers"
        :loading="loading"
        :bordered="false"
        :single-line="true"
        size="small"
        scroll-x="600"
      />
    </NCard>
  </div>
</template>
