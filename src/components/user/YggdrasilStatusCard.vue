<script setup lang="ts">
import type { YggdrasilStatusResp } from '@/types/auth'
import SkinViewer3D from '@/components/skin/SkinViewer3D.vue'

defineProps<{
  status: YggdrasilStatusResp | null
  loading: boolean
  available: boolean
  skinUrl: string | null
  capeUrl: string | null
  model: 'default' | 'slim'
}>()

const emit = defineEmits<{
  setSkin: []
  setCape: []
  clearSkin: []
  clearCape: []
}>()
</script>

<template>
  <NCard size="small" class="h-full">
    <template #header>
      <NFlex align="center" :size="8">
        <span class="i-ph-cube-duotone text-4 text-[var(--n-primary-color)]" />
        <NText strong>
          Yggdrasil 档案
        </NText>
      </NFlex>
    </template>

    <NSpin v-if="loading" size="small" class="w-full py-8" />
    <NEmpty
      v-else-if="!status"
      :description="available ? '暂时无法读取 Yggdrasil 状态' : '当前账户没有读取权限'"
      class="py-8"
    />
    <template v-else>
      <div class="grid gap-4 sm:grid-cols-[180px_1fr]">
        <!-- 3D 角色 -->
        <div class="skin-stage relative aspect-3/4 w-full overflow-hidden rounded-lg border border-[var(--n-border-color)] sm:aspect-auto">
          <SkinViewer3D
            :skin-url="skinUrl"
            :cape-url="capeUrl"
            :model="model"
            :lazy="false"
            :interactive="true"
            :auto-rotate="true"
            :zoom="0.95"
          />
          <NTag
            v-if="status.has_profile"
            size="tiny"
            type="success"
            :bordered="false"
            class="absolute left-2 top-2 backdrop-blur-md"
          >
            已绑定
          </NTag>
        </div>

        <!-- 信息与操作 -->
        <div>
          <NDescriptions label-placement="left" :column="1" size="small">
            <NDescriptionsItem label="绑定状态">
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
            <NDescriptionsItem label="当前皮肤">
              <NText depth="3">
                {{ status.texture_skin_id ? '已装备' : '未设置' }}
              </NText>
            </NDescriptionsItem>
            <NDescriptionsItem label="当前披风">
              <NText depth="3">
                {{ status.texture_cape_id ? '已装备' : '未设置' }}
              </NText>
            </NDescriptionsItem>
          </NDescriptions>

          <template v-if="status.has_profile">
            <NDivider />
            <NFlex :wrap="true" align="center" :size="10">
              <NButton size="small" type="primary" secondary @click="emit('setSkin')">
                <template #icon>
                  <span class="i-ph-user-duotone" />
                </template>
                设置皮肤
              </NButton>
              <NButton v-if="status.texture_skin_id" size="small" @click="emit('clearSkin')">
                清除皮肤
              </NButton>
              <NButton size="small" type="primary" secondary @click="emit('setCape')">
                <template #icon>
                  <span class="i-ph-flag-duotone" />
                </template>
                设置披风
              </NButton>
              <NButton v-if="status.texture_cape_id" size="small" @click="emit('clearCape')">
                清除披风
              </NButton>
            </NFlex>
          </template>
          <NText v-else depth="3" class="mt-3 block text-3">
            绑定 Yggdrasil 档案后即可在游戏内使用皮肤
          </NText>
        </div>
      </div>
    </template>
  </NCard>
</template>
