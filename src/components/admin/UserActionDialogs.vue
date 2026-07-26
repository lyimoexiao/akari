<script setup lang="ts">
import type { UserItem } from '@/types/admin'

defineProps<{
  user: UserItem | null
  roles: readonly string[]
  loading: boolean
}>()
defineEmits<{
  saveRole: []
  savePassword: []
}>()
const roleOpen = defineModel<boolean>('roleOpen', { required: true })
const passwordOpen = defineModel<boolean>('passwordOpen', { required: true })
const role = defineModel<string>('role', { required: true })
const password = defineModel<string>('password', { required: true })
</script>

<template>
  <NModal v-model:show="roleOpen" preset="card" class="w-[calc(100vw-32px)]! max-w-100" title="修改角色">
    <NFlex vertical :size="16">
      <NText>用户：{{ user?.username }}</NText>
      <NSelect v-model:value="role" :options="roles.map(name => ({ label: name, value: name }))" />
      <NButton type="primary" :loading="loading" @click="$emit('saveRole')">
        保存角色
      </NButton>
    </NFlex>
  </NModal>

  <NModal v-model:show="passwordOpen" preset="card" class="w-[calc(100vw-32px)]! max-w-100" title="重置密码">
    <NFlex vertical :size="16">
      <NText>用户：{{ user?.username }}</NText>
      <NInput v-model:value="password" type="password" autocomplete="new-password" placeholder="新密码（至少 6 位）" show-password-on="click" />
      <NButton type="primary" :disabled="password.length < 6" :loading="loading" @click="$emit('savePassword')">
        确认重置
      </NButton>
    </NFlex>
  </NModal>
</template>
