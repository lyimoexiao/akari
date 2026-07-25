import GoCaptcha from 'go-captcha-vue'
import NaiveUI from 'naive-ui'

import { createPinia } from 'pinia'

import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import 'virtual:uno.css'
import 'go-captcha-vue/dist/style.css'

const app = createApp(App)

app.use(createPinia())
app.use(router)
app.use(GoCaptcha)
app.use(NaiveUI)

app.mount('#app')
