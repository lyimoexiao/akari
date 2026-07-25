export interface UserItem {
  id: number
  username: string
  email: string
  role: 'super_admin' | 'staff' | 'user'
  email_verified_at: string | null
  created_at: string
}

export interface ListUsersResp {
  items: UserItem[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

export interface ListUsersReq {
  page?: number
  page_size?: number
  query?: string
}
