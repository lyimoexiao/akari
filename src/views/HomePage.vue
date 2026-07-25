<script setup lang="ts">
import type { YggdrasilStatusResp } from '@/types/auth'
import { NButton, NCard, NDescriptions, NDescriptionsItem, NSpin, NTag, NText } from 'naive-ui'
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { getYggdrasilStatus } from '@/api/auth'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const router = useRouter()

const yggStatus = ref<YggdrasilStatusResp | null>(null)
const yggLoading = ref(false)

onMounted(async () => {
  if (!authStore.token)
    return

  try {
    await authStore.refreshCurrentUser()
  }
  catch {
    authStore.logout()
    router.push('/login')
    return
  }

  yggLoading.value = true
  try {
    yggStatus.value = await getYggdrasilStatus(authStore.token)
  }
  catch {
    // Yggdrasil not configured — silently ignore
  }
  yggLoading.value = false
})

function handleLogout() {
  authStore.logout().then(() => router.push('/login'))
}
</script>

<template>
  <n-layout class="min-h-screen" position="static">
    <n-layout-header bordered>
      <div class="max-w-4xl mx-auto px-4 py-3 flex items-center justify-between">
        <h1 class="text-lg font-bold">
          Akari
        </h1>
        <div class="flex items-center gap-3">
          <span class="text-sm op-fade">{{ authStore.username }}</span>

          <NTag
            v-if="authStore.role"
            :type="authStore.role === 'super_admin' ? 'warning' : authStore.role === 'staff' ? 'info' : 'default'"
            size="small"
            :bordered="false"
          >
            {{ authStore.role }}
          </NTag>

          <NButton
            v-if="authStore.isStaff"
            size="small"
            secondary
            @click="router.push('/admin')"
          >
            管理后台
          </NButton>

          <NButton size="small" quaternary @click="handleLogout">
            退出登录
          </NButton>
        </div>
      </div>
    </n-layout-header>

    <n-layout-content>
      <div class="max-w-4xl mx-auto px-4 py-8">
        <NCard :title="`欢迎回来，${authStore.username}！`" size="large" :bordered="true">
          <NDescriptions label-placement="left" :column="1">
            <NDescriptionsItem label="邮箱">
              <div class="flex items-center gap-2">
                {{ authStore.email }}
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
              <NText code>
                {{ authStore.user?.id }}
              </NText>
            </NDescriptionsItem>
          </NDescriptions>

          <template #action>
            <NButton
              v-if="!authStore.isEmailVerified"
              type="primary"
              size="small"
              @click="router.push('/verify-email')"
            >
              验证邮箱
            </NButton>
          </template>
        </NCard>

        <NCard title="权限信息" size="large" :bordered="true" class="mt-4">
          <NDescriptions label-placement="left" :column="1">
            <NDescriptionsItem label="超级管理员">
              <NTag
                :type="authStore.isSuperAdmin ? 'success' : 'default'"
                size="small"
                :bordered="false"
              >
                {{ authStore.isSuperAdmin }}
              </NTag>
            </NDescriptionsItem>

            <NDescriptionsItem label="职员">
              <NTag
                :type="authStore.isStaff ? 'success' : 'default'"
                size="small"
                :bordered="false"
              >
                {{ authStore.isStaff }}
              </NTag>
            </NDescriptionsItem>
          </NDescriptions>
        </NCard>

        <NCard title="Yggdrasil" size="large" :bordered="true" class="mt-4">
          <template v-if="yggLoading">
            <NSpin size="small" />
          </template>
          <template v-else-if="yggStatus">
            <NDescriptions label-placement="left" :column="1">
              <NDescriptionsItem label="绑定档案">
                <NTag
                  :type="yggStatus.has_profile ? 'success' : 'warning'"
                  size="small"
                  :bordered="false"
                >
                  {{ yggStatus.has_profile ? '是' : '否' }}
                </NTag>
              </NDescriptionsItem>

              <NDescriptionsItem v-if="yggStatus.profile_uuid" label="档案 UUID">
                <NText code class="text-xs">
                  {{ yggStatus.profile_uuid }}
                </NText>
              </NDescriptionsItem>

              <NDescriptionsItem v-if="yggStatus.profile_name" label="档案名称">
                {{ yggStatus.profile_name }}
              </NDescriptionsItem>

              <NDescriptionsItem v-if="yggStatus.last_login_at" label="最后登录">
                {{ new Date(yggStatus.last_login_at).toLocaleString() }}
              </NDescriptionsItem>

              <NDescriptionsItem v-if="yggStatus.last_login_ip" label="最后登录 IP">
                <NText code>
                  {{ yggStatus.last_login_ip }}
                </NText>
              </NDescriptionsItem>
            </NDescriptions>
          </template>
          <template v-else>
            <span class="text-sm op-fade">不可用</span>
          </template>
        </NCard>
      </div>
    </n-layout-content>
  </n-layout>
</template>
