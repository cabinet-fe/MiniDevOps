import { defineConfig } from 'vite'
import Vue from '@vitejs/plugin-vue'
import VueJSX from '@vitejs/plugin-vue-jsx'
import Components from 'unplugin-vue-components/vite'
import { autoResolveComponent } from 'vite-helper'
import { existModule } from 'cat-kit/be'
import { resolve, dirname } from 'path'
import { fileURLToPath } from 'url'
const __dirname = dirname(fileURLToPath(import.meta.url))

export default defineConfig(() => {
  return {
    plugins: [
      Vue(),
      VueJSX(),
      Components({
        resolvers: [
          autoResolveComponent({
            prefix: 'U',
            lib: 'ultra-ui',
            sideEffects(kebabName, lib) {
              let moduleId = `${lib}/components/${kebabName}/style.ts`

              while (!existModule(moduleId)) {
                const preKebabName = kebabName
                kebabName = kebabName.replace(/-[a-z]$/, '')
                if (preKebabName === kebabName) return
                moduleId = `${lib}/components/${kebabName}/style.ts`
              }

              return moduleId
            }
          })
        ]
      })
    ],
    server: {
      port: 3000
    },
    base: '/',

    resolve: {
      extensions: ['.ts', '.js', '.json', '.tsx']
    }
  }
})
