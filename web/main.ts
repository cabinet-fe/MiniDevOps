import { createApp, h } from 'vue'
import App from './app.vue'
import { router } from './router'
import { vLoading } from 'ultra-ui/components/loading/directive'
import { loadTheme } from 'ultra-ui'
import 'ultra-ui/styles'
import 'virtual:uno.css'
import { authHttp, http } from '@meta/utils'

loadTheme()

authHttp.setDefaultConfig({
  baseUrl: '/api'
})

http.setDefaultConfig({
  baseUrl: '/api'
})

const app = createApp({
  render: () => h(App)
})

app.config.globalProperties.c = console

app.directive('loading', vLoading)

app.use(router)

app.mount('#app')
