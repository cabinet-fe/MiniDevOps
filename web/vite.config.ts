import { defineConfig } from 'vite'
import Vue from '@vitejs/plugin-vue'
import VueJSX from '@vitejs/plugin-vue-jsx'
import Components from 'unplugin-vue-components/vite'
import {
  UltraUIResolver,
  MetaComponentsResolver,
  defineServerProxy
} from 'vite-helper'

import { dirname } from 'path'
import { fileURLToPath } from 'url'
import UnoCSS from 'unocss/vite'

const __dirname = dirname(fileURLToPath(import.meta.url))

export default defineConfig(() => {
  return {
    plugins: [
      Vue(),
      VueJSX(),
      Components({
        resolvers: [UltraUIResolver, MetaComponentsResolver]
      }),
      UnoCSS()
    ],
    server: {
      port: 3001,
      proxy: defineServerProxy({
        '/api': 'http://localhost:3000/api'
      })
    },
    base: '/',

    resolve: {
      extensions: ['.ts', '.js', '.json', '.tsx'],
      alias: {
        '@': __dirname
      }
    }
  }
})
