<script setup lang="ts">
import type { FormInst, FormRules } from 'naive-ui'
import { useUrlSearchParams } from '@vueuse/core'
import { useMessage } from 'naive-ui'
import { reactive, shallowRef, useTemplateRef } from 'vue'
import { useRouter } from 'vue-router'
import { ApiRequestError } from '@/api'
import { resetPassword } from '@/api/auth'
import AuthShell from '@/components/auth/AuthShell.vue'

const router = useRouter()
const message = useMessage()
const params = useUrlSearchParams<{ token?: string }>('history', { write: false })
const formRef = useTemplateRef<FormInst>('form')
const form = reactive({
  token: typeof params.token === 'string' ? params.token : '',
  password: '',
  confirmPassword: '',
})
const loading = shallowRef(false)
const done = shallowRef(false)

const rules: FormRules = {
  token: [{ required: true, message: '请输入重置令牌', trigger: ['blur', 'input'] }],
  password: [
    { required: true, message: '请输入新密码', trigger: ['blur', 'input'] },
    { min: 6, max: 128, message: '密码应为 6–128 个字符', trigger: 'blur' },
  ],
  confirmPassword: [
    { required: true, message: '请确认新密码', trigger: ['blur', 'input'] },
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

  loading.value = true
  try {
    await resetPassword({ token: form.token, password: form.password })
    done.value = true
  }
  catch (error) {
    message.error(error instanceof ApiRequestError ? error.message : '重置密码失败，请稍后重试')
  }
  finally {
    loading.value = false
  }
}
</script>

<template>
  <AuthShell title="设置新密码" description="输入邮件中的重置令牌和新密码">
    <NForm v-if="!done" ref="form" :model="form" :rules="rules" @submit.prevent="submit">
      <NFormItem label="重置令牌" path="token">
        <NInput v-model:value="form.token" placeholder="邮件中的重置令牌" :disabled="loading" />
      </NFormItem>
      <NFormItem label="新密码" path="password">
        <NInput v-model:value="form.password" type="password" autocomplete="new-password" placeholder="至少 6 个字符" show-password-on="click" :disabled="loading" />
      </NFormItem>
      <NFormItem label="确认密码" path="confirmPassword">
        <NInput v-model:value="form.confirmPassword" type="password" autocomplete="new-password" placeholder="再次输入新密码" show-password-on="click" :disabled="loading" />
      </NFormItem>
      <NButton type="primary" block attr-type="submit" :loading="loading" size="large">
        重置密码
      </NButton>
    </NForm>

    <NResult v-else status="success" title="密码已重置" description="现在可以使用新密码登录。">
      <template #footer>
        <NButton type="primary" @click="router.replace({ name: 'AuthLogin' })">
          前往登录
        </NButton>
      </template>
    </NResult>

    <template #footer>
      <NButton text type="primary" @click="router.push({ name: 'AuthForgotPassword' })">
        重新发送令牌
      </NButton>
    </template>
  </AuthShell>
</template>
