<script setup lang="ts">
import type { UserResp } from '@/types/auth'
import { useDateFormat } from '@vueuse/core'
import RoleTag from '@/components/workspace/RoleTag.vue'

const props = defineProps<{
  user: UserResp
}>()

defineEmits<{
  verifyEmail: []
}>()

const registeredAt = useDateFormat(() => new Date(props.user.created_at), 'YYYY-MM-DD HH:mm:ss')
</script>

<template>
  <NCard size="small">
    <template #header>
      <NFlex align="center" :size="8">
        <span class="i-ph-user-circle-duotone text-4 text-[var(--n-primary-color)]" />
        <NText strong>
          账户资料
        </NText>
      </NFlex>
    </template>
    <template #header-extra>
      <RoleTag :role="user.role" />
    </template>
    <NDescriptions label-placement="left" :column="1" size="small">
      <NDescriptionsItem label="用户名">
        {{ user.username }}
      </NDescriptionsItem>
      <NDescriptionsItem label="邮箱">
        <NFlex align="center" :size="8">
          <span>{{ user.email }}</span>
          <NTag :type="user.email_verified_at ? 'success' : 'warning'" size="tiny" :bordered="false">
            {{ user.email_verified_at ? '已验证' : '未验证' }}
          </NTag>
        </NFlex>
      </NDescriptionsItem>
      <NDescriptionsItem label="用户 ID">
        <NText code>
          {{ user.id }}
        </NText>
      </NDescriptionsItem>
      <NDescriptionsItem label="注册时间">
        {{ registeredAt }}
      </NDescriptionsItem>
    </NDescriptions>

    <template v-if="!user.email_verified_at" #action>
      <NButton type="primary" size="small" @click="$emit('verifyEmail')">
        验证邮箱
      </NButton>
    </template>
  </NCard>
</template>
