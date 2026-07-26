<script setup lang="ts">
import type { FormInst, FormRules } from 'naive-ui'
import { useMessage } from 'naive-ui'
import { reactive, shallowRef, useTemplateRef } from 'vue'
import { useRouter } from 'vue-router'
import { ApiRequestError } from '@/api'
import { sendVerificationEmail, verifyEmail } from '@/api/auth'
import AuthShell from '@/components/auth/AuthShell.vue'
import { useAuthStore } from '@/stores/auth'

type VerificationStep = 'send' | 'verify' | 'done'

const router = useRouter()
const authStore = useAuthStore()
const message = useMessage()
const formRef = useTemplateRef<FormInst>('form')
const form = reactive({ token: '' })
const step = shallowRef<VerificationStep>('send')
const loading = shallowRef(false)

const rules: FormRules = {
  token: [{ required: true, message: '请输入验证令牌', trigger: ['blur', 'input'] }],
}

async function sendEmail(): Promise<void> {
  const token = authStore.token
  if (!token)
    return

  loading.value = true
  try {
    const responseMessage = await sendVerificationEmail(token)
    message.success(responseMessage)
    step.value = 'verify'
  }
  catch (error) {
    message.error(error instanceof ApiRequestError ? error.message : '发送验证邮件失败')
  }
  finally {
    loading.value = false
  }
}

async function submit(): Promise<void> {
  try {
    await formRef.value?.validate()
  }
  catch (error) {
    if (Array.isArray(error))
      return
    throw error
  }

  const token = authStore.token
  if (!token)
    return

  loading.value = true
  try {
    await verifyEmail(token, form.token)
    await authStore.refreshCurrentUser()
    step.value = 'done'
  }
  catch (error) {
    message.error(error instanceof ApiRequestError ? error.message : '邮箱验证失败')
  }
  finally {
    loading.value = false
  }
}
</script>

<template>
  <AuthShell title="验证邮箱" description="完成验证以保护你的 Akari 账户">
    <NSteps :current="step === 'send' ? 1 : step === 'verify' ? 2 : 3" size="small" class="mb-6">
      <NStep title="发送" />
      <NStep title="验证" />
      <NStep title="完成" />
    </NSteps>

    <NFlex v-if="step === 'send'" vertical :size="16">
      <NText depth="3" class="text-center">
        验证邮件将发送至 <strong>{{ authStore.email }}</strong>
      </NText>
      <NButton type="primary" block size="large" :loading="loading" @click="sendEmail">
        发送验证邮件
      </NButton>
    </NFlex>

    <NForm v-else-if="step === 'verify'" ref="form" :model="form" :rules="rules" @submit.prevent="submit">
      <NAlert type="info" :show-icon="true" class="mb-4">
        验证令牌已发送，请检查收件箱和垃圾邮件。
      </NAlert>
      <NFormItem label="验证令牌" path="token">
        <NInput v-model:value="form.token" placeholder="在此粘贴令牌" :disabled="loading" />
      </NFormItem>
      <NButton type="primary" block attr-type="submit" size="large" :loading="loading">
        验证邮箱
      </NButton>
    </NForm>

    <NResult v-else status="success" title="邮箱已验证" description="你的账户已解锁全部可用功能。">
      <template #footer>
        <NButton type="primary" @click="router.replace({ name: 'UserHome' })">
          返回账户中心
        </NButton>
      </template>
    </NResult>
  </AuthShell>
</template>
