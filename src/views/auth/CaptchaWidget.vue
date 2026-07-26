<script setup lang="ts">
import type { CaptchaChallenge } from '@/api/captcha'
import type { CaptchaAnswer } from '@/types/auth'
import { useScriptTag } from '@vueuse/core'
import { onMounted, onUnmounted, useTemplateRef } from 'vue'
import { z } from 'zod'

const props = defineProps<{ captcha: CaptchaChallenge }>()
const emit = defineEmits<{
  result: [value: CaptchaAnswer]
  refresh: []
  close: []
}>()

const slidePointSchema = z.object({ x: z.number(), y: z.number().optional() })
const turnstileContainer = useTemplateRef<HTMLDivElement>('turnstile')
let turnstileWidgetId: string | null = null

const { load: loadTurnstile } = useScriptTag(
  'https://challenges.cloudflare.com/turnstile/v0/api.js?render=explicit',
  undefined,
  { manual: true, defer: true },
)

function confirmClick(dots: unknown): boolean {
  emit('result', { dots })
  return true
}

function confirmRotate(angle: number): boolean {
  emit('result', { angle })
  return true
}

function confirmSlide(point: unknown): boolean {
  const parsed = slidePointSchema.safeParse(point)
  if (!parsed.success) {
    emit('refresh')
    return false
  }
  emit('result', { x: parsed.data.x, y: parsed.data.y ?? 0 })
  return true
}

function renderTurnstile(): void {
  if (!turnstileContainer.value || !window.turnstile || !('provider' in props.captcha))
    return

  turnstileWidgetId = window.turnstile.render(turnstileContainer.value, {
    'sitekey': props.captcha.site_key,
    'callback': token => emit('result', { turnstile_token: token }),
    'expired-callback': () => emit('refresh'),
    'error-callback': () => emit('refresh'),
  })
}

onMounted(async () => {
  if (!('provider' in props.captcha))
    return
  if (!window.turnstile)
    await loadTurnstile()
  renderTurnstile()
})

onUnmounted(() => {
  if (turnstileWidgetId && window.turnstile)
    window.turnstile.remove(turnstileWidgetId)
})
</script>

<template>
  <gocaptcha-click
    v-if="'type' in captcha && captcha.type === 'click'"
    :config="{ width: 300, height: 240, title: '请依次点击', buttonText: '确认' }"
    :data="{ image: captcha.master_image, thumb: captcha.thumb_image }"
    :events="{ confirm: confirmClick, refresh: () => emit('refresh'), close: () => emit('close') }"
  />
  <gocaptcha-rotate
    v-else-if="'type' in captcha && captcha.type === 'rotate'"
    :config="{ width: 280 }"
    :data="{ image: captcha.master_image, thumb: captcha.thumb_image }"
    :events="{ confirm: confirmRotate, refresh: () => emit('refresh'), close: () => emit('close') }"
  />
  <gocaptcha-slide
    v-else-if="'type' in captcha && captcha.type === 'slide'"
    :config="{ width: 300, height: 240 }"
    :data="{
      image: captcha.master_image,
      thumb: captcha.tile_image,
      thumbX: captcha.thumb_x,
      thumbY: captcha.thumb_y,
      thumbWidth: captcha.tile_width,
      thumbHeight: captcha.tile_height,
    }"
    :events="{ confirm: confirmSlide, refresh: () => emit('refresh'), close: () => emit('close') }"
  />
  <div v-else ref="turnstile" class="min-h-16 flex justify-center" />
</template>
