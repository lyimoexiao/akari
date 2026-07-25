<script setup lang="ts">
import type { FormInst, FormRules } from 'naive-ui'
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ApiRequestError } from '@/api'
import { sendVerificationEmail, verifyEmail } from '@/api/auth'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const formRef = ref<FormInst | null>(null)

const step = ref<'send' | 'verify' | 'done'>('send')
const verificationToken = ref('')
const message = ref('')
const error = ref('')
const loading = ref(false)

const rules: FormRules = {
  token: [
    { required: true, message: '请输入验证令牌', trigger: 'blur' },
  ],
}

async function handleSendEmail() {
  error.value = ''
  loading.value = true
  try {
    const res = await sendVerificationEmail(authStore.token!)
    message.value = res
    step.value = 'verify'
  }
  catch (err) {
    if (err instanceof ApiRequestError) {
      error.value = err.message
    }
    else {
      error.value = '发送验证邮件失败'
    }
  }
  finally {
    loading.value = false
  }
}

async function handleVerify() {
  error.value = ''

  try {
    await formRef.value?.validate()
  }
  catch {
    return
  }

  loading.value = true
  try {
    const res = await verifyEmail(authStore.token!, verificationToken.value)
    message.value = res
    step.value = 'done'
    await authStore.refreshCurrentUser()
  }
  catch (err) {
    if (err instanceof ApiRequestError) {
      error.value = err.message
    }
    else {
      error.value = '验证失败'
    }
  }
  finally {
    loading.value = false
  }
}

function goHome() {
  router.push('/')
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center" style="background: var(--n-color-body);">
    <n-card style="width: 400px;" :bordered="true" size="large">
      <template #header>
        <div class="text-center">
          <h1 class="text-2xl font-bold mb-1" style="color: var(--n-text-color);">
            邮箱验证
          </h1>
        </div>
      </template>

      <n-steps :current="step === 'send' ? 0 : step === 'verify' ? 1 : 2" class="mb-6">
        <n-step title="发送" />
        <n-step title="验证" />
        <n-step title="完成" />
      </n-steps>

      <template v-if="step === 'send'">
        <p class="text-sm text-center mb-6" style="color: var(--n-text-color-3);">
          验证邮箱 <strong>{{ authStore.email }}</strong> 以解锁全部功能。
        </p>

        <n-alert v-if="error" type="error" :closable="false" class="mb-4">
          {{ error }}
        </n-alert>

        <n-button
          type="primary"
          block
          size="large"
          :loading="loading"
          @click="handleSendEmail"
        >
          发送验证邮件
        </n-button>
      </template>

      <template v-if="step === 'verify'">
        <p class="text-sm text-center mb-2" style="color: var(--n-text-color-3);">
          验证令牌已发送至你的邮箱。
        </p>
        <p class="text-xs text-center mb-6" style="color: var(--n-text-color-3);">
          输入下方令牌完成验证。
        </p>

        <n-form ref="formRef" :rules="rules" :model="{ token: verificationToken }" @submit.prevent="handleVerify">
          <n-form-item label="验证令牌" path="token">
            <n-input
              v-model:value="verificationToken"
              placeholder="在此粘贴令牌"
              :disabled="loading"
            />
          </n-form-item>

          <n-alert v-if="error" type="error" :closable="false" class="mb-4">
            {{ error }}
          </n-alert>

          <n-button
            type="primary"
            block
            attr-type="submit"
            size="large"
            :loading="loading"
          >
            验证邮箱
          </n-button>
        </n-form>
      </template>

      <template v-if="step === 'done'">
        <div class="text-center py-4">
          <n-icon size="48" color="#18a058">
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor"><path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z" /></svg>
          </n-icon>
          <p class="text-sm mt-4 mb-6" style="color: var(--n-text-color-3);">
            你的邮箱已验证成功！
          </p>

          <n-button
            type="primary"
            block
            size="large"
            @click="goHome"
          >
            返回首页
          </n-button>
        </div>
      </template>
    </n-card>
  </div>
</template>
