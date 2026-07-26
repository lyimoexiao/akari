<script setup lang="ts">
import type { YggdrasilStatusResp } from '@/types/auth'

const props = defineProps<{
  status: YggdrasilStatusResp | null
  loading: boolean
  available: boolean
}>()

const emit = defineEmits<{
  setSkin: []
  setCape: []
  clearSkin: []
  clearCape: []
}>()

const previewUrl = (hash: string) => `${import.meta.env.VITE_API_BASE_URL ?? ''}/api/v1/raw/${hash}`
</script>

<template>
  <NCard title="Yggdrasil 档案" size="small">
    <NSpin v-if="loading" size="small" />
    <template v-else-if="status">
      <NDescriptions label-placement="left" :column="1" size="small">
        <NDescriptionsItem label="绑定档案">
          <NTag :type="status.has_profile ? 'success' : 'warning'" size="small" :bordered="false">
            {{ status.has_profile ? '已绑定' : '未绑定' }}
          </NTag>
        </NDescriptionsItem>
        <NDescriptionsItem v-if="status.profile_uuid" label="档案 UUID">
          <NText code class="break-all text-3">
            {{ status.profile_uuid }}
          </NText>
        </NDescriptionsItem>
        <NDescriptionsItem v-if="status.profile_name" label="档案名称">
          {{ status.profile_name }}
        </NDescriptionsItem>
        <NDescriptionsItem v-if="status.last_login_at" label="最后登录">
          {{ status.last_login_at }}
        </NDescriptionsItem>
        <NDescriptionsItem v-if="status.last_login_ip" label="最后登录 IP">
          <NText code>
            {{ status.last_login_ip }}
          </NText>
        </NDescriptionsItem>
      </NDescriptions>

      <template v-if="status.has_profile">
        <NDivider />
        <NFlex :wrap="true" align="center" :gap="12">
          <NButton size="small" @click="emit('setSkin')">
            设置皮肤
          </NButton>
          <NButton v-if="status.texture_skin_id" size="small" @click="emit('clearSkin')">
            清除皮肤
          </NButton>
          <NButton size="small" @click="emit('setCape')">
            设置披风
          </NButton>
          <NButton v-if="status.texture_cape_id" size="small" @click="emit('clearCape')">
            清除披风
          </NButton>
        </NFlex>
      </template>
    </template>
    <NEmpty v-else :description="available ? '暂时无法读取 Yggdrasil 状态' : '当前账户没有读取权限'" />
  </NCard>
</template>
