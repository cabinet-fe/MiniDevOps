import { createApp, h } from 'vue'
import App from './app.vue'
import { router } from './router'
import 'ultra-ui/styles/theme.css'
import 'ultra-ui/styles/index'
import { vLoading } from 'ultra-ui/components/loading/directive'
import 'ultra-ui/components/loading/style.css'
import 'ultra-ui/components/context-menu/style'

const app = createApp({
  render: () => h(App)
})

app.config.globalProperties.c = console

app.directive('loading', vLoading)

app.use(router)

app.mount('#app')
