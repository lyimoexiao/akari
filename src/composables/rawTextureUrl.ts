export function rawTextureUrl(hash: string, url?: string | null): string {
  return url || `${import.meta.env.VITE_API_BASE_URL ?? ''}/api/v1/raw/${hash}`
}
