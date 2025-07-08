import { createApp, h } from 'vue'
import { router } from './router'
import { vLoading } from 'ultra-ui/components/loading/directive'
import 'ultra-ui/styles'
import { loadTheme } from 'ultra-ui'
import 'virtual:uno.css'

import { RouterView } from 'vue-router'

loadTheme()

const app = createApp({
  render: () => h(RouterView)
})

app.directive('loading', vLoading)

app.use(router)

app.mount('#app')
