<script setup lang="ts">
import type { FormInst, FormRules } from 'naive-ui'
import { useMessage } from 'naive-ui'
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ApiRequestError } from '@/api'
import { resetPassword } from '@/api/auth'

const router = useRouter()
const message = useMessage()

const formRef = ref<FormInst | null>(null)

const token = ref('')
const password = ref('')
const confirmPassword = ref('')
const loading = ref(false)
const done = ref(false)

const rules: FormRules = {
  token: [
    { required: true, message: '请输入重置令牌', trigger: 'blur' },
  ],
  password: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, message: '至少 6 个字符', trigger: 'blur' },
    { max: 128, message: '最多 128 个字符', trigger: 'blur' },
  ],
  confirmPassword: [
    { required: true, message: '请确认密码', trigger: 'blur' },
    {
      validator: (_rule, value: string) => value === password.value,
      message: '密码不匹配',
      trigger: 'blur',
    },
  ],
}

async function handleReset() {
  try {
    await formRef.value?.validate()
  }
  catch {
    return
  }

  loading.value = true
  try {
    await resetPassword({
      token: token.value,
      password: password.value,
    })
    done.value = true
  }
  catch (err) {
    const msg = err instanceof ApiRequestError ? err.message : '发生未知错误'
    message.error(msg)
  }
  finally {
    loading.value = false
  }
}

function goLogin() {
  router.push('/login')
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center" style="background: var(--n-color-body);">
    <n-card style="width: 400px;" :bordered="true" size="large">
      <template #header>
        <div class="text-center">
          <h1 class="text-2xl font-bold mb-1" style="color: var(--n-text-color);">
            设置新密码
          </h1>
          <p class="text-sm" style="color: var(--n-text-color-3);">
            输入邮件中的重置令牌和新密码
          </p>
        </div>
      </template>

      <template v-if="!done">
        <n-form ref="formRef" :rules="rules" :model="{ token, password, confirmPassword }" @submit.prevent="handleReset">
          <n-form-item label="重置令牌" path="token">
            <n-input
              v-model:value="token"
              placeholder="邮件中的重置令牌"
              :disabled="loading"
            />
          </n-form-item>

          <n-form-item label="新密码" path="password">
            <n-input
              v-model:value="password"
              type="password"
              placeholder="至少 6 个字符"
              show-password-on="click"
              :disabled="loading"
            />
          </n-form-item>

          <n-form-item label="确认密码" path="confirmPassword">
            <n-input
              v-model:value="confirmPassword"
              type="password"
              placeholder="再次输入新密码"
              show-password-on="click"
              :disabled="loading"
            />
          </n-form-item>

          <n-button
            type="primary"
            block
            attr-type="submit"
            :loading="loading"
            size="large"
          >
            重置密码
          </n-button>
        </n-form>
      </template>

      <template v-else>
        <div class="text-center py-4">
          <n-icon size="48" color="#18a058">
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor"><path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z" /></svg>
          </n-icon>
          <p class="text-sm mt-4 mb-6" style="color: var(--n-text-color-3);">
            密码已重置成功！
          </p>

          <n-button
            type="primary"
            block
            size="large"
            @click="goLogin"
          >
            前往登录
          </n-button>
        </div>
      </template>

      <p class="text-sm text-center mt-6" style="color: var(--n-text-color-3);">
        还没有令牌？
        <router-link to="/forgot-password" style="color: var(--n-primary-color); text-decoration: none;">
          重新发送
        </router-link>
      </p>
    </n-card>
  </div>
</template>
