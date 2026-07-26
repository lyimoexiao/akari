import 'vue-router'

export {}

declare module 'vue-router' {
  interface RouteMeta {
    readonly title: string
    readonly publicOnly?: boolean
    readonly requiresAuth?: boolean
    readonly permissions?: readonly string[]
  }
}
