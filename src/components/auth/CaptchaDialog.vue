<script setup lang="ts">
import type { CaptchaChallenge } from '@/api/captcha'
import type { CaptchaAnswer } from '@/types/auth'
import CaptchaWidget from '@/views/auth/CaptchaWidget.vue'

defineProps<{
  challenge: CaptchaChallenge | null
  challengeKey: number
}>()

defineEmits<{
  result: [value: CaptchaAnswer]
  refresh: []
}>()

const show = defineModel<boolean>({ required: true })
</script>

<template>
  <NModal
    v-model:show="show"
    preset="card"
    class="w-[calc(100vw-32px)]! max-w-90"
    title="安全验证"
    :bordered="false"
    :segmented="false"
  >
    <CaptchaWidget
      v-if="challenge"
      :key="challengeKey"
      :captcha="challenge"
      @result="$emit('result', $event)"
      @refresh="$emit('refresh')"
      @close="show = false"
    />
  </NModal>
</template>
