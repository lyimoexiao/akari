<script setup lang="ts">
import type { CaptchaData } from '@/api/captcha'
import { ref } from 'vue'

defineProps<{
  captcha: CaptchaData
}>()

const emit = defineEmits<{
  (e: 'result', value: Record<string, unknown>): void
  (e: 'refresh'): void
  (e: 'close'): void
}>()

const captchaRef = ref<any>(null)

function onConfirmClick(dots: any, _reset: () => void) {
  emit('result', { dots } as Record<string, unknown>)
  return true
}

function onConfirmRotate(angle: number, _reset: () => void) {
  emit('result', { angle } as Record<string, unknown>)
  return true
}

function onConfirmSlide(point: any, _reset: () => void) {
  emit('result', { x: point?.x ?? 0, y: point?.y ?? 0 } as Record<string, unknown>)
  return true
}
</script>

<template>
  <gocaptcha-click
    v-if="captcha.type === 'click'"
    ref="captchaRef"
    :config="{ width: 300, height: 240, title: '请依次点击', buttonText: '确认' }"
    :data="{ image: captcha.master_image, thumb: captcha.thumb_image }"
    :events="{ confirm: onConfirmClick, refresh: () => emit('refresh'), close: () => emit('close') }"
  />
  <gocaptcha-rotate
    v-else-if="captcha.type === 'rotate'"
    ref="captchaRef"
    :config="{ width: 280 }"
    :data="{ image: captcha.master_image, thumb: captcha.thumb_image }"
    :events="{ confirm: onConfirmRotate, refresh: () => emit('refresh'), close: () => emit('close') }"
  />
  <gocaptcha-slide
    v-else-if="captcha.type === 'slide'"
    ref="captchaRef"
    :config="{ width: 300, height: 240 }"
    :data="{
      image: captcha.master_image,
      thumb: captcha.tile_image ?? captcha.thumb_image ?? '',
      thumbX: captcha.thumb_x ?? 0,
      thumbY: captcha.thumb_y ?? 0,
      thumbWidth: captcha.tile_width ?? 40,
      thumbHeight: captcha.tile_height ?? 40,
    }"
    :events="{ confirm: onConfirmSlide, refresh: () => emit('refresh'), close: () => emit('close') }"
  />
</template>
