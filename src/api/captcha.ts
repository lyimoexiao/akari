import { http } from '@/api/index'

export interface CaptchaEnvelope {
  enabled: boolean
  data: CaptchaData
}

export interface CaptchaData {
  captcha_id: string
  type: string
  master_image: string
  thumb_image?: string
  tile_image?: string
  tile_width?: number
  tile_height?: number
  thumb_x?: number
  thumb_y?: number
  angle?: number
}

export function fetchCaptcha(): Promise<CaptchaEnvelope> {
  return http.get<CaptchaEnvelope>('/captcha')
}
