<script setup lang="ts">
import type { YggdrasilStatusResp } from '@/types/auth'
import { useAsyncState } from '@vueuse/core'
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { ApiRequestError } from '@/api'
import { getYggdrasilStatus } from '@/api/auth'
import AccountProfileCard from '@/components/user/AccountProfileCard.vue'
import YggdrasilStatusCard from '@/components/user/YggdrasilStatusCard.vue'
import WorkspacePage from '@/components/workspace/WorkspacePage.vue'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()
const canReadYggdrasil = computed(() => authStore.hasPermission('yggdrasil.status.read'))

async function loadStatus(): Promise<YggdrasilStatusResp | null> {
  await authStore.refreshCurrentUser()
  const token = authStore.token
  if (!token || !canReadYggdrasil.value)
    return null
  return getYggdrasilStatus(token)
}

const {
  state: yggdrasilStatus,
  isLoading,
  error,
  execute: refresh,
} = useAsyncState(loadStatus, null)

const errorMessage = computed(() => error.value instanceof ApiRequestError
  ? error.value.message
  : error.value ? '账户信息加载失败' : null)
</script>

<template>
  <WorkspacePage title="账户概览" :description="`欢迎回来，${authStore.username ?? '用户'}`">
    <NAlert v-if="errorMessage" type="error" class="mb-4">
      {{ errorMessage }}
      <template #action>
        <NButton text type="error" @click="refresh(0)">
          重试
        </NButton>
      </template>
    </NAlert>

    <div class="grid gap-4 md:grid-cols-2">
      <AccountProfileCard
        v-if="authStore.user"
        :user="authStore.user"
        @verify-email="router.push({ name: 'AuthVerifyEmail' })"
      />
      <YggdrasilStatusCard
        :status="yggdrasilStatus"
        :loading="isLoading"
        :available="canReadYggdrasil"
      />
    </div>
  </WorkspacePage>
</template>
