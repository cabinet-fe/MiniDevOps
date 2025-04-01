import { defineConfig } from 'vite'
import { pluginPresets, createServer } from '@builder/vite'
import { dirname } from 'path'
import { fileURLToPath } from 'url'
import UnoCSS from 'unocss/vite'

const __dirname = dirname(fileURLToPath(import.meta.url))

export default defineConfig(() => {
  return {
    plugins: [
      ...pluginPresets(['vue', 'vue-jsx', 'unplugin-components'], {
        'unplugin-components': {
          dts: './components.d.ts'
        }
      }),

      UnoCSS()
    ],
    server: createServer({
      port: 3001,
      proxy: {
        '/api': 'http://localhost:8080',
        '/ws': {
          target: 'ws://localhost:8080',
          ws: true,
          rewriteWsOrigin: true
        }
      }
    }),
    base: '/',

    resolve: {
      extensions: ['.ts', '.js', '.json', '.tsx'],
      alias: {
        '@': __dirname
      }
    }
  }
})
