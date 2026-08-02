import { rawTextureUrl } from './rawTextureUrl'

/**
 * 根据纹理信息生成预览 URL。
 * 优先使用后端返回的 url，否则拼 /api/v1/raw/{hash}。
 */
export function texturePreviewUrl(texture: { hash: string, url?: string | null }): string {
  return rawTextureUrl(texture.hash, texture.url)
}
