import { defineConfig } from 'vite'
import {
  pluginPresets,
  createServer,
  autoResolveComponent,
  UltraUIResolver
} from '@builder/vite'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import UnoCSS from 'unocss/vite'
import fg from 'fast-glob'

const __dirname = dirname(fileURLToPath(import.meta.url))

export default defineConfig(async () => {
  const optimizeDeps = await fg.glob(['ultra-ui/components/**/style.js'], {
    cwd: resolve(__dirname, 'node_modules')
  })

  return {
    plugins: [
      ...pluginPresets(['vue', 'vue-jsx', 'unplugin-components'], {
        'unplugin-components': {
          dts: true,
          resolvers: [
            UltraUIResolver,
            autoResolveComponent({
              lib: '@/components',
              prefix: 'M',
              sideEffects: () => undefined
            })
          ]
        }
      }),

      UnoCSS()
    ],
    server: createServer({
      port: 3001,
      proxy: {
        '/api/v1': 'http://127.0.0.1:8080/api/v1'
        // '/ws': {
        //   target: 'ws://localhost:8080',
        //   ws: true,
        //   rewriteWsOrigin: true
        // }
      }
    }),
    base: '/',

    css: {
      preprocessorOptions: {
        scss: {
          additionalData: `
            @use 'ultra-ui/styles/mixins' as m;
            @use 'ultra-ui/styles/vars' as vars;
            @use 'ultra-ui/styles/functions' as fn;
          `
        }
      }
    },

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
