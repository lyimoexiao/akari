import type { CaptchaChallenge, CaptchaEnvelope } from '@/api/captcha'
import type { CaptchaAnswer } from '@/types/auth'
import { useAsyncState, useCounter, useToggle } from '@vueuse/core'
import { computed, shallowRef } from 'vue'
import { ApiRequestError } from '@/api'
import { fetchCaptcha } from '@/api/captcha'

const DISABLED_CAPTCHA: CaptchaEnvelope = { enabled: false, data: null }

export function useCaptchaChallenge() {
  const result = shallowRef<CaptchaAnswer | null>(null)
  const [isOpen, toggleOpen] = useToggle(false)
  const { count: challengeKey, inc: incrementChallengeKey } = useCounter()
  const {
    state: envelope,
    isReady,
    isLoading,
    execute,
  } = useAsyncState(fetchCaptcha, DISABLED_CAPTCHA, {
    immediate: false,
    resetOnExecute: false,
    throwError: true,
  })

  const challenge = computed<CaptchaChallenge | null>(() => envelope.value.data)
  const requestFields = computed(() => {
    if (!result.value || !challenge.value)
      return {}

    if ('provider' in challenge.value) {
      const token = result.value.turnstile_token
      return typeof token === 'string' ? { captcha_token: token } : {}
    }

    return {
      captcha_id: challenge.value.captcha_id,
      user_answer: result.value,
    }
  })

  async function load(): Promise<void> {
    try {
      await execute(0)
      incrementChallengeKey()
    }
    catch (error) {
      if (error instanceof ApiRequestError && error.status === 404) {
        envelope.value = DISABLED_CAPTCHA
        return
      }
      throw error
    }
  }

  async function allowSubmit(): Promise<boolean> {
    if (!isReady.value)
      await load()

    if (envelope.value.enabled && result.value === null) {
      toggleOpen(true)
      return false
    }
    return true
  }

  function acceptResult(value: CaptchaAnswer): void {
    result.value = value
    toggleOpen(false)
  }

  async function refresh(): Promise<void> {
    result.value = null
    await load()
  }

  return {
    challenge,
    challengeKey,
    isLoading,
    isOpen,
    requestFields,
    acceptResult,
    allowSubmit,
    refresh,
    toggleOpen,
  }
}
