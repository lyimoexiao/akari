<script setup lang="ts">
import type { FormInst, FormRules } from 'naive-ui'
import { useMessage } from 'naive-ui'
import { reactive, shallowRef, useTemplateRef } from 'vue'
import { useRouter } from 'vue-router'
import { ApiRequestError } from '@/api'
import { requestPasswordReset } from '@/api/auth'
import AuthShell from '@/components/auth/AuthShell.vue'
import CaptchaDialog from '@/components/auth/CaptchaDialog.vue'
import { useCaptchaChallenge } from '@/composables/useCaptchaChallenge'

const router = useRouter()
const message = useMessage()
const formRef = useTemplateRef<FormInst>('form')
const form = reactive({ email: '' })
const loading = shallowRef(false)
const sent = shallowRef(false)
const captcha = useCaptchaChallenge()

const rules: FormRules = {
  email: [
    { required: true, message: '请输入邮箱', trigger: ['blur', 'input'] },
    { type: 'email', message: '邮箱格式不正确', trigger: 'blur' },
  ],
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

  if (!await captcha.allowSubmit())
    return

  loading.value = true
  try {
    await requestPasswordReset({ email: form.email, ...captcha.requestFields.value })
    sent.value = true
  }
  catch (error) {
    message.error(error instanceof ApiRequestError ? error.message : '发送失败，请稍后重试')
    await captcha.refresh()
  }
  finally {
    loading.value = false
  }
}

function acceptCaptcha(value: Readonly<Record<string, unknown>>): void {
  captcha.acceptResult(value)
  void submit()
}
</script>

<template>
  <AuthShell title="找回密码" description="我们会向你的邮箱发送一次性重置令牌">
    <NForm v-if="!sent" ref="form" :model="form" :rules="rules" @submit.prevent="submit">
      <NFormItem label="邮箱" path="email">
        <NInput v-model:value="form.email" type="email" autocomplete="email" placeholder="you@example.com" :disabled="loading" />
      </NFormItem>
      <NButton type="primary" block attr-type="submit" :loading="loading" size="large">
        发送重置邮件
      </NButton>
    </NForm>

    <NResult v-else status="success" title="邮件已发送" description="如果该邮箱存在，你会收到一封包含重置令牌的邮件。">
      <template #footer>
        <NButton type="primary" @click="router.push({ name: 'AuthResetPassword' })">
          输入重置令牌
        </NButton>
      </template>
    </NResult>

    <template #footer>
      <NButton text type="primary" @click="router.push({ name: 'AuthLogin' })">
        返回登录
      </NButton>
    </template>

    <CaptchaDialog
      v-model="captcha.isOpen.value"
      :challenge="captcha.challenge.value"
      :challenge-key="captcha.challengeKey.value"
      @result="acceptCaptcha"
      @refresh="captcha.refresh"
    />
  </AuthShell>
</template>
