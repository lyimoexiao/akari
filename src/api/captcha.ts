import { z } from 'zod'
import { http } from './index'

const clickCaptchaSchema = z.object({
  captcha_id: z.string(),
  type: z.literal('click'),
  master_image: z.string(),
  thumb_image: z.string(),
})

const rotateCaptchaSchema = z.object({
  captcha_id: z.string(),
  type: z.literal('rotate'),
  master_image: z.string(),
  thumb_image: z.string(),
  angle: z.number(),
})

const slideCaptchaSchema = z.object({
  captcha_id: z.string(),
  type: z.literal('slide'),
  master_image: z.string(),
  tile_image: z.string(),
  tile_width: z.number(),
  tile_height: z.number(),
  thumb_x: z.number(),
  thumb_y: z.number(),
})

const turnstileDataSchema = z.object({
  provider: z.literal('turnstile'),
  site_key: z.string(),
})

export const captchaChallengeSchema = z.union([
  clickCaptchaSchema,
  rotateCaptchaSchema,
  slideCaptchaSchema,
  turnstileDataSchema,
])

const captchaEnvelopeSchema = z.union([
  z.object({ enabled: z.literal(false) }).transform(() => ({ enabled: false as const, data: null })),
  z.object({ enabled: z.literal(true), data: captchaChallengeSchema }),
])

export type CaptchaData = z.infer<typeof clickCaptchaSchema>
  | z.infer<typeof rotateCaptchaSchema>
  | z.infer<typeof slideCaptchaSchema>
export type TurnstileData = z.infer<typeof turnstileDataSchema>
export type CaptchaChallenge = z.infer<typeof captchaChallengeSchema>
export type CaptchaEnvelope = z.infer<typeof captchaEnvelopeSchema>

export function fetchCaptcha() {
  return http.get({ path: 'captcha', schema: captchaEnvelopeSchema })
}
