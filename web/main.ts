import { createApp, h } from 'vue'
import { router } from './router'
import { vLoading } from 'ultra-ui/components/loading/directive'
import { loadTheme } from 'ultra-ui'
import 'ultra-ui/styles'
import 'virtual:uno.css'

import { RouterView } from 'vue-router'

loadTheme()

const app = createApp({
  render: () => h(RouterView)
})

app.config.globalProperties.c = console

app.directive('loading', vLoading)

app.use(router)

app.mount('#app')
