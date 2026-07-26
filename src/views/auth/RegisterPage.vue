<script setup lang="ts">
import type { FormInst, FormRules } from 'naive-ui'
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
const form = reactive({ username: '', email: '', password: '', confirmPassword: '' })
const captcha = useCaptchaChallenge()

const rules: FormRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: ['blur', 'input'] },
    { min: 3, max: 64, message: '用户名应为 3–64 个字符', trigger: 'blur' },
  ],
  email: [
    { required: true, message: '请输入邮箱', trigger: ['blur', 'input'] },
    { type: 'email', message: '邮箱格式不正确', trigger: 'blur' },
  ],
  password: [
    { required: true, message: '请输入密码', trigger: ['blur', 'input'] },
    { min: 6, max: 128, message: '密码应为 6–128 个字符', trigger: 'blur' },
  ],
  confirmPassword: [
    { required: true, message: '请再次输入密码', trigger: ['blur', 'input'] },
    { validator: (_rule, value: string) => value === form.password, message: '两次输入的密码不一致', trigger: ['blur', 'input'] },
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
    await authStore.register({
      username: form.username,
      email: form.email,
      password: form.password,
      ...captcha.requestFields.value,
    })
    await router.replace({ name: 'UserHome' })
  }
  catch (error) {
    message.error(error instanceof ApiRequestError ? error.message : '注册失败，请稍后重试')
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
  <AuthShell title="创建账号" description="加入 Akari">
    <NForm ref="form" :model="form" :rules="rules" @submit.prevent="submit">
      <NFormItem label="用户名" path="username">
        <NInput v-model:value="form.username" autocomplete="username" placeholder="你的用户名" :disabled="loading" />
      </NFormItem>
      <NFormItem label="邮箱" path="email">
        <NInput v-model:value="form.email" type="email" autocomplete="email" placeholder="you@example.com" :disabled="loading" />
      </NFormItem>
      <NFormItem label="密码" path="password">
        <NInput v-model:value="form.password" type="password" autocomplete="new-password" placeholder="至少 6 个字符" show-password-on="click" :disabled="loading" />
      </NFormItem>
      <NFormItem label="确认密码" path="confirmPassword">
        <NInput v-model:value="form.confirmPassword" type="password" autocomplete="new-password" placeholder="再次输入密码" show-password-on="click" :disabled="loading" />
      </NFormItem>
      <NButton type="primary" block attr-type="submit" :loading="loading" size="large">
        创建账号
      </NButton>
    </NForm>

    <template #footer>
      <NText depth="3">
        已有账号？
        <NButton text type="primary" @click="router.push({ name: 'AuthLogin' })">
          登录
        </NButton>
      </NText>
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
