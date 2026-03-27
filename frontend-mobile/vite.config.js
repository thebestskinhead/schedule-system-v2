import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import Components from 'unplugin-vue-components/vite'
import { VantResolver } from '@vant/auto-import-resolver'

export default defineConfig({
  base: '/mobile/',
  plugins: [
    vue(),
    Components({
      resolvers: [VantResolver()],
    }),
  ],
  build: {
    outDir: 'dist',
    assetsDir: 'assets'
  },
  server: {
    port: 5174,
    host: '0.0.0.0',
    hmr: true,
    allowedHosts: ['ba0ef1631b5b4535bd920b7a7ea3db0b--5174.ap-shanghai2.cloudstudio.club'],
    proxy: {
      '/api': {
        target: 'https://ba0ef1631b5b4535bd920b7a7ea3db0b--8080.ap-shanghai2.cloudstudio.club',
        secure: false,
        changeOrigin: true
      }
    }
  }
})
