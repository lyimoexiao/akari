import { z } from 'zod'

export const textureItemSchema = z.object({
  tid: z.number().int().positive(),
  name: z.string().min(1),
  type: z.enum(['steve', 'alex', 'cape']),
  hash: z.string().min(1),
  url: z.string().optional(),
  size: z.number().int().nonnegative(),
  uploader: z.number().int().nonnegative(),
  public: z.boolean(),
  likes: z.number().int().nonnegative(),
  upload_at: z.string().min(1),
})

export const textureDetailSchema = textureItemSchema.extend({
  uploader_name: z.string().optional(),
})

export const listTexturesResponseSchema = z.object({
  items: z.array(textureItemSchema),
  total: z.number().int().nonnegative(),
  page: z.number().int().positive(),
  page_size: z.number().int().positive(),
  total_pages: z.number().int().nonnegative(),
})

export const closetItemSchema = z.object({
  texture_tid: z.number().int().positive(),
  item_name: z.string().min(1),
  created_at: z.string().min(1),
  texture: z.object({
    name: z.string().min(1),
    type: z.enum(['steve', 'alex', 'cape']),
    hash: z.string().min(1),
    url: z.string().optional(),
    size: z.number().int().nonnegative(),
    public: z.boolean(),
    likes: z.number().int().nonnegative(),
    uploader: z.number().int().nonnegative(),
  }).optional(),
})

export const listClosetResponseSchema = z.object({
  items: z.array(closetItemSchema),
  total: z.number().int().nonnegative(),
  page: z.number().int().positive(),
  page_size: z.number().int().positive(),
  total_pages: z.number().int().nonnegative(),
})

export const closetIdsSchema = z.object({
  ids: z.array(z.number().int()),
})

export type TextureItem = z.infer<typeof textureItemSchema>
export type TextureDetail = z.infer<typeof textureDetailSchema>
export type ListTexturesResp = z.infer<typeof listTexturesResponseSchema>
export type ClosetItem = z.infer<typeof closetItemSchema>
export type ListClosetResp = z.infer<typeof listClosetResponseSchema>
export type ClosetIds = z.infer<typeof closetIdsSchema>

export interface ListTexturesReq {
  readonly page?: number
  readonly page_size?: number
  readonly type?: string
  readonly search?: string
  readonly order?: string
}

export interface ListClosetReq {
  readonly page?: number
  readonly page_size?: number
  readonly type?: string
  readonly search?: string
}

export interface AddClosetReq {
  readonly tid: number
  readonly name: string
}

export interface RenameClosetReq {
  readonly name: string
}

export interface UpdateTextureReq {
  readonly name: string
  readonly public: boolean
}
