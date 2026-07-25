<script setup lang="ts">
import { NButton, NCard, NDescriptions, NDescriptionsItem, NTag, NText } from 'naive-ui'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const router = useRouter()
</script>

<template>
  <div class="p-6">
    <h2 class="text-lg font-semibold mb-5">
      系统设置
    </h2>

    <NCard title="个人信息" :bordered="true" size="small" class="mb-4">
      <template #header-extra>
        <NTag
          :type="authStore.role === 'super_admin' ? 'warning' : authStore.role === 'staff' ? 'info' : 'default'"
          size="small"
          :bordered="false"
        >
          {{ authStore.role === 'super_admin' ? '超级管理员' : authStore.role === 'staff' ? 'Staff' : '用户' }}
        </NTag>
      </template>
      <NDescriptions label-placement="left" :column="2" size="small">
        <NDescriptionsItem label="用户名">
          {{ authStore.username }}
        </NDescriptionsItem>
        <NDescriptionsItem label="邮箱">
          <div class="flex items-center gap-2">
            <span class="font-mono text-xs">{{ authStore.email }}</span>
            <NTag
              :type="authStore.isEmailVerified ? 'success' : 'warning'"
              size="tiny"
              :bordered="false"
            >
              {{ authStore.isEmailVerified ? '已验证' : '未验证' }}
            </NTag>
          </div>
        </NDescriptionsItem>
        <NDescriptionsItem label="角色">
          {{ authStore.role }}
        </NDescriptionsItem>
        <NDescriptionsItem label="用户 ID">
          <NText code class="text-xs">
            {{ authStore.user?.id }}
          </NText>
        </NDescriptionsItem>
      </NDescriptions>
    </NCard>

    <NCard title="导航" :bordered="true" size="small">
      <NButton size="small" @click="router.push('/')">
        <template #icon>
          <span class="i-ph-house-duotone" />
        </template>
        返回首页
      </NButton>
    </NCard>
  </div>
</template>
