<script setup lang="ts">
import RoleTag from '@/components/workspace/RoleTag.vue'
import WorkspacePage from '@/components/workspace/WorkspacePage.vue'
import { LANGUAGES, THEME_MODES, useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const appStore = useAppStore()

const themeOptions = THEME_MODES.map(value => ({
  value,
  label: value === 'auto' ? '跟随系统' : value === 'light' ? '浅色' : '深色',
}))

const languageOptions = LANGUAGES.map(value => ({
  value,
  label: value === 'zh-CN' ? '简体中文' : 'English',
}))
</script>

<template>
  <WorkspacePage title="管理设置" description="查看当前管理身份并调整本地界面偏好">
    <div class="grid gap-4 lg:grid-cols-2">
      <NCard title="当前身份" size="small">
        <template v-if="authStore.role" #header-extra>
          <RoleTag :role="authStore.role" />
        </template>
        <NDescriptions label-placement="left" :column="1" size="small">
          <NDescriptionsItem label="用户名">
            {{ authStore.username }}
          </NDescriptionsItem>
          <NDescriptionsItem label="邮箱">
            {{ authStore.email }}
          </NDescriptionsItem>
          <NDescriptionsItem label="用户 ID">
            <NText code>
              {{ authStore.user?.id }}
            </NText>
          </NDescriptionsItem>
          <NDescriptionsItem label="有效权限">
            {{ authStore.permissions.length }} 项
          </NDescriptionsItem>
        </NDescriptions>
      </NCard>

      <NCard title="界面偏好" size="small">
        <NForm label-placement="left" label-width="80">
          <NFormItem label="主题">
            <NSelect v-model:value="appStore.themeMode" :options="themeOptions" />
          </NFormItem>
          <NFormItem label="语言">
            <NSelect v-model:value="appStore.lang" :options="languageOptions" />
          </NFormItem>
        </NForm>
      </NCard>
    </div>
  </WorkspacePage>
</template>
