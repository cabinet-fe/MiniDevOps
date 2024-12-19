import { defineConfig } from 'vite'
import { pluginPresets, createServer } from '@builder/vite'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import UnoCSS from 'unocss/vite'
import fg from 'fast-glob'

const __dirname = dirname(fileURLToPath(import.meta.url))

export default defineConfig(async () => {
  const optimizeDeps = await fg.glob(
    ['ultra-ui/components/**/style.js', '@meta/components/**/style.js'],
    {
      cwd: resolve(__dirname, '../node_modules')
    }
  )

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
    },
    optimizeDeps: {
      include: optimizeDeps
    }
  }
})
