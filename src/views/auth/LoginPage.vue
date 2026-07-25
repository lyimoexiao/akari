<script setup lang="ts">
import type { FormInst, FormRules } from 'naive-ui'
import type { CaptchaData } from '@/api/captcha'
import { useMessage } from 'naive-ui'
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ApiRequestError } from '@/api'
import { fetchCaptcha } from '@/api/captcha'
import { useAuthStore } from '@/stores/auth'
import CaptchaWidget from './CaptchaWidget.vue'

const router = useRouter()
const authStore = useAuthStore()
const message = useMessage()

const formRef = ref<FormInst | null>(null)

const username = ref('')
const password = ref('')
const loading = ref(false)

// Captcha state
const captchaEnabled = ref(false)
const captchaData = ref<CaptchaData | null>(null)
const captchaResult = ref<Record<string, unknown> | null>(null)
const showCaptcha = ref(false)
const captchaKey = ref(0)

const rules: FormRules = {
  username: [
    { required: true, message: '请输入用户名或邮箱', trigger: 'blur' },
    { min: 3, message: '至少 3 个字符', trigger: 'blur' },
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '至少 6 个字符', trigger: 'blur' },
  ],
}

async function loadCaptcha() {
  try {
    const envelope = await fetchCaptcha()
    if (envelope.enabled && envelope.data) {
      captchaData.value = envelope.data
      captchaEnabled.value = true
      captchaKey.value++
    }
  }
  catch {
    // captcha disabled
  }
  if (captchaData.value === null) {
    captchaData.value = {} as CaptchaData
  }
}

async function openCaptcha() {
  captchaResult.value = null
  showCaptcha.value = true
}

function onCaptchaResult(result: Record<string, unknown>) {
  captchaResult.value = result
  showCaptcha.value = false
  handleLogin()
}

function onCaptchaClose() {
  showCaptcha.value = false
}

function onCaptchaRefresh() {
  loadCaptcha()
}

async function handleLogin() {
  try {
    await formRef.value?.validate()
  }
  catch {
    return
  }

  if (captchaData.value === null) {
    await loadCaptcha()
  }

  if (captchaEnabled.value && captchaResult.value === null) {
    openCaptcha()
    return
  }

  loading.value = true
  try {
    await authStore.login({
      username: username.value,
      password: password.value,
      captcha_id: captchaData.value?.captcha_id,
      user_answer: captchaResult.value ?? undefined,
    })
    router.push('/')
  }
  catch (err) {
    const msg = err instanceof ApiRequestError ? err.message : '发生未知错误'
    message.error(msg)
    loadCaptcha()
    captchaResult.value = null
  }
  finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center" style="background: var(--n-color-body);">
    <n-card style="width: 400px;" :bordered="true" size="large">
      <template #header>
        <div class="text-center">
          <h1 class="text-2xl font-bold mb-1" style="color: var(--n-text-color);">
            登录
          </h1>
          <p class="text-sm" style="color: var(--n-text-color-3);">
            欢迎回到 Akari
          </p>
        </div>
      </template>

      <n-form ref="formRef" :rules="rules" :model="{ username, password }" @submit.prevent="handleLogin">
        <n-form-item label="用户名或邮箱" path="username">
          <n-input
            v-model:value="username"
            placeholder="用户名或邮箱"
            :disabled="loading"
          />
        </n-form-item>

        <n-form-item label="密码" path="password">
          <n-input
            v-model:value="password"
            type="password"
            placeholder="请输入密码"
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
          登录
        </n-button>
      </n-form>

      <p class="text-sm text-center mt-2" style="color: var(--n-text-color-3);">
        <router-link to="/forgot-password" style="color: var(--n-primary-color); text-decoration: none;">
          忘记密码？
        </router-link>
      </p>

      <p class="text-sm text-center mt-4" style="color: var(--n-text-color-3);">
        没有账号？
        <router-link to="/register" style="color: var(--n-primary-color); text-decoration: none;">
          注册
        </router-link>
      </p>
    </n-card>

    <n-modal
      v-model:show="showCaptcha"
      :mask-closable="true"
      preset="card"
      style="width: 360px;"
      title="安全验证"
      :bordered="false"
      :segmented="false"
      @update:show="onCaptchaClose"
    >
      <CaptchaWidget
        v-if="captchaData"
        :key="captchaKey"
        :captcha="captchaData"
        @result="onCaptchaResult"
        @refresh="onCaptchaRefresh"
        @close="onCaptchaClose"
      />
    </n-modal>
  </div>
</template>
