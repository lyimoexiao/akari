import type { DialogApi, LoadingBarApi, MessageApi, ModalApi, NotificationApi } from 'naive-ui'

declare const __BUILD_VERSION__: string
declare const __BUILD_GIT_HASH__: string

declare global {
  interface TurnstileRenderOptions {
    readonly 'sitekey': string
    readonly 'callback': (token: string) => void
    readonly 'expired-callback': () => void
    readonly 'error-callback': () => void
  }

  interface TurnstileApi {
    render: (element: HTMLElement, options: TurnstileRenderOptions) => string
    remove: (widgetId: string) => void
  }

  interface Window {
    $message?: MessageApi
    $dialog?: DialogApi
    $notification?: NotificationApi
    $loadingBar?: LoadingBarApi
    $modal?: ModalApi
    turnstile?: TurnstileApi
  }
}
