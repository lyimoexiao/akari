<script setup lang="ts">
import type { YggdrasilStatusResp } from '@/types/auth'
import { useAsyncState } from '@vueuse/core'
import { useMessage } from 'naive-ui'
import { computed, shallowRef } from 'vue'
import { useRouter } from 'vue-router'
import { ApiRequestError } from '@/api'
import { clearProfileCape, clearProfileSkin, getYggdrasilStatus, setProfileCape, setProfileSkin } from '@/api/auth'
import AccountProfileCard from '@/components/user/AccountProfileCard.vue'
import TexturePickerDialog from '@/components/user/TexturePickerDialog.vue'
import YggdrasilStatusCard from '@/components/user/YggdrasilStatusCard.vue'
import WorkspacePage from '@/components/workspace/WorkspacePage.vue'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const message = useMessage()
const authStore = useAuthStore()
const canReadYggdrasil = computed(() => authStore.hasPermission('yggdrasil.status.read'))

const pickerType = shallowRef<'skin' | 'cape'>('skin')
const pickerTitle = shallowRef('')
const pickerShow = shallowRef(false)

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

function openSetSkin() {
  pickerType.value = 'skin'
  pickerTitle.value = '选择皮肤'
  pickerShow.value = true
}

function openSetCape() {
  pickerType.value = 'cape'
  pickerTitle.value = '选择披风'
  pickerShow.value = true
}

async function onTextureSelected(tid: number) {
  const token = authStore.token
  if (!token)
    return
  try {
    if (pickerType.value === 'cape') {
      await setProfileCape(token, tid)
      message.success('披风已设置')
    }
    else {
      await setProfileSkin(token, tid)
      message.success('皮肤已设置')
    }
    void refresh(0)
  }
  catch (err) {
    message.error(err instanceof ApiRequestError ? err.message : '设置失败')
  }
}

async function onClearSkin() {
  const token = authStore.token
  if (!token)
    return
  try {
    await clearProfileSkin(token)
    message.success('皮肤已清除')
    void refresh(0)
  }
  catch (err) {
    message.error(err instanceof ApiRequestError ? err.message : '清除失败')
  }
}

async function onClearCape() {
  const token = authStore.token
  if (!token)
    return
  try {
    await clearProfileCape(token)
    message.success('披风已清除')
    void refresh(0)
  }
  catch (err) {
    message.error(err instanceof ApiRequestError ? err.message : '清除失败')
  }
}
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
        @set-skin="openSetSkin"
        @set-cape="openSetCape"
        @clear-skin="onClearSkin"
        @clear-cape="onClearCape"
      />
    </div>

    <TexturePickerDialog
      v-model:show="pickerShow"
      :title="pickerTitle"
      :type="pickerType"
      @select="onTextureSelected"
    />
  </WorkspacePage>
</template>
