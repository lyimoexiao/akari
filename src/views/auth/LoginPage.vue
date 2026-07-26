<script setup lang="ts">
import type { FormInst, FormRules } from 'naive-ui'
import { useUrlSearchParams } from '@vueuse/core'
import { useMessage } from 'naive-ui'
import { reactive, shallowRef, useTemplateRef } from 'vue'
import { useRouter } from 'vue-router'
import { ApiRequestError } from '@/api'
import AuthShell from '@/components/auth/AuthShell.vue'
import CaptchaDialog from '@/components/auth/CaptchaDialog.vue'
import { useCaptchaChallenge } from '@/composables/useCaptchaChallenge'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()
const message = useMessage()
const formRef = useTemplateRef<FormInst>('form')
const loading = shallowRef(false)
const form = reactive({ username: '', password: '' })
const params = useUrlSearchParams<{ redirect?: string }>('history', { write: false })
const captcha = useCaptchaChallenge()

const rules: FormRules = {
  username: [
    { required: true, message: '请输入用户名或邮箱', trigger: ['blur', 'input'] },
    { min: 3, message: '至少 3 个字符', trigger: 'blur' },
  ],
  password: [
    { required: true, message: '请输入密码', trigger: ['blur', 'input'] },
    { min: 6, message: '至少 6 个字符', trigger: 'blur' },
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
    await authStore.login({
      username: form.username,
      password: form.password,
      ...captcha.requestFields.value,
    })
    const redirect = typeof params.redirect === 'string' && params.redirect.startsWith('/')
      ? params.redirect
      : '/user'
    await router.replace(redirect)
  }
  catch (error) {
    message.error(error instanceof ApiRequestError ? error.message : '登录失败，请稍后重试')
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
  <AuthShell title="登录" description="欢迎回到 Akari！">
    <NForm ref="form" :model="form" :rules="rules" @submit.prevent="submit">
      <NFormItem label="用户名或邮箱" path="username">
        <NInput v-model:value="form.username" autocomplete="username" placeholder="请输入用户名或邮箱" :disabled="loading" />
      </NFormItem>
      <NFormItem label="密码" path="password">
        <NInput v-model:value="form.password" type="password" autocomplete="current-password" placeholder="请输入密码" show-password-on="click" :disabled="loading" />
      </NFormItem>
      <NButton type="primary" block attr-type="submit" :loading="loading" size="large">
        登录
      </NButton>
      <NFlex class="pt-4" justify="space-between">
        <NButton text type="primary" @click="router.push({ name: 'AuthForgotPassword' })">
          忘记密码？
        </NButton>
        <NButton text type="primary" @click="router.push({ name: 'AuthRegister' })">
          注册新账号
        </NButton>
      </NFlex>
    </NForm>

    <CaptchaDialog
      v-model="captcha.isOpen.value"
      :challenge="captcha.challenge.value"
      :challenge-key="captcha.challengeKey.value"
      @result="acceptCaptcha"
      @refresh="captcha.refresh"
    />
  </AuthShell>
</template>
