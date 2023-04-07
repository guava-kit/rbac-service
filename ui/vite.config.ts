import path from 'path'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import Unocss from 'unocss/vite'
import ElementPlus from 'unplugin-element-plus/dist/vite'

function resolve(p: string) {
  return path.resolve(__dirname, p)
}

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [vue(), Unocss(), ElementPlus({})],
  resolve: {
    alias: {
      '@': resolve('./src'),
      _c: resolve('./src/components'),
      'vue-i18n': 'vue-i18n/dist/vue-i18n.cjs.js',
    },
  },
  build: {
    target: 'esnext',
  },
})
