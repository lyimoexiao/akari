<script setup lang="ts">
import type { FormInst, FormRules } from 'naive-ui'
import type { CaptchaData } from '@/api/captcha'
import { useMessage } from 'naive-ui'
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ApiRequestError } from '@/api'
import { requestPasswordReset } from '@/api/auth'
import { fetchCaptcha } from '@/api/captcha'
import CaptchaWidget from './CaptchaWidget.vue'

const router = useRouter()
const message = useMessage()

const formRef = ref<FormInst | null>(null)

const email = ref('')
const loading = ref(false)
const sent = ref(false)

const captchaEnabled = ref(false)
const captchaData = ref<CaptchaData | null>(null)
const captchaResult = ref<Record<string, unknown> | null>(null)
const showCaptcha = ref(false)
const captchaKey = ref(0)

const rules: FormRules = {
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '邮箱格式不正确', trigger: 'blur' },
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

function onCaptchaResult(result: Record<string, unknown>) {
  captchaResult.value = result
  showCaptcha.value = false
  handleRequest()
}

function onCaptchaClose() {
  showCaptcha.value = false
}

function onCaptchaRefresh() {
  loadCaptcha()
}

async function handleRequest() {
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
    showCaptcha.value = true
    return
  }

  loading.value = true
  try {
    await requestPasswordReset({
      email: email.value,
      captcha_id: captchaData.value?.captcha_id,
      user_answer: captchaResult.value ?? undefined,
    })
    sent.value = true
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
            重置密码
          </h1>
          <p class="text-sm" style="color: var(--n-text-color-3);">
            输入你的邮箱以接收重置令牌
          </p>
        </div>
      </template>

      <template v-if="!sent">
        <n-form ref="formRef" :rules="rules" :model="{ email }" @submit.prevent="handleRequest">
          <n-form-item label="邮箱" path="email">
            <n-input
              v-model:value="email"
              type="email"
              placeholder="you@example.com"
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
            发送重置邮件
          </n-button>
        </n-form>
      </template>

      <template v-else>
        <div class="text-center py-4">
          <n-icon size="48" color="#18a058">
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor"><path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z" /></svg>
          </n-icon>
          <p class="text-sm mt-4" style="color: var(--n-text-color-3);">
            如果该邮箱存在，重置邮件已发送。
          </p>
          <p class="text-xs mt-2 mb-6" style="color: var(--n-text-color-3);">
            请检查你的收件箱（及垃圾邮件），并使用邮件中的令牌重置密码。
          </p>

          <n-button
            type="primary"
            block
            size="large"
            @click="router.push('/reset-password')"
          >
            前往重置密码
          </n-button>
        </div>
      </template>

      <p class="text-sm text-center mt-6" style="color: var(--n-text-color-3);">
        想起密码了？
        <router-link to="/login" style="color: var(--n-primary-color); text-decoration: none;">
          返回登录
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
