<script setup lang="ts">
import type { CaptchaData, TurnstileData } from '@/api/captcha'
import { onMounted, onUnmounted, ref } from 'vue'

const props = defineProps<{
  captcha: CaptchaData | TurnstileData
}>()

const emit = defineEmits<{
  (e: 'result', value: Record<string, unknown>): void
  (e: 'refresh'): void
  (e: 'close'): void
}>()

// ── go-captcha ──
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

// ── Turnstile ──
const turnstileContainer = ref<HTMLDivElement | null>(null)
const turnstileWidgetId = ref<string | null>(null)

function loadTurnstileScript(): Promise<void> {
  return new Promise((resolve) => {
    if ((window as any).turnstile) {
      resolve()
      return
    }
    const script = document.createElement('script')
    script.src = 'https://challenges.cloudflare.com/turnstile/v0/api.js?onload=onTurnstileLoad'
    script.async = true
    script.defer = true
    ;(window as any).onTurnstileLoad = () => {
      resolve()
    }
    document.head.appendChild(script)
  })
}

function renderTurnstile() {
  if (!turnstileContainer.value || !(window as any).turnstile)
    return
  const data = props.captcha as TurnstileData
  turnstileWidgetId.value = (window as any).turnstile.render(turnstileContainer.value, {
    'sitekey': data.site_key,
    'callback': (token: string) => {
      emit('result', { turnstile_token: token } as Record<string, unknown>)
    },
    'expired-callback': () => {
      emit('refresh')
    },
    'error-callback': () => {
      emit('refresh')
    },
  })
}

onMounted(async () => {
  if (props.captcha && 'provider' in props.captcha && props.captcha.provider === 'turnstile') {
    await loadTurnstileScript()
    renderTurnstile()
  }
})

onUnmounted(() => {
  if (turnstileWidgetId.value && (window as any).turnstile) {
    (window as any).turnstile.remove(turnstileWidgetId.value)
  }
})
</script>

<template>
  <!-- go-captcha -->
  <gocaptcha-click
    v-if="'type' in captcha && captcha.type === 'click'"
    ref="captchaRef"
    :config="{ width: 300, height: 240, title: '请依次点击', buttonText: '确认' }"
    :data="{ image: captcha.master_image, thumb: captcha.thumb_image }"
    :events="{ confirm: onConfirmClick, refresh: () => emit('refresh'), close: () => emit('close') }"
  />
  <gocaptcha-rotate
    v-else-if="'type' in captcha && captcha.type === 'rotate'"
    ref="captchaRef"
    :config="{ width: 280 }"
    :data="{ image: captcha.master_image, thumb: captcha.thumb_image }"
    :events="{ confirm: onConfirmRotate, refresh: () => emit('refresh'), close: () => emit('close') }"
  />
  <gocaptcha-slide
    v-else-if="'type' in captcha && captcha.type === 'slide'"
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
  <!-- Turnstile -->
  <div
    v-else-if="'provider' in captcha && captcha.provider === 'turnstile'"
    ref="turnstileContainer"
  />
</template>
