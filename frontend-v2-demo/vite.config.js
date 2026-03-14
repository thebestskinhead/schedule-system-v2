import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  server: {
    port: 8081,
    host: '0.0.0.0',
    strictPort: true,
    allowedHosts: ['.cloudstudio.club', 'localhost', '127.0.0.1'],
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        configure: (proxy, options) => {
          proxy.on('error', (err, req, res) => {
            console.warn('[代理错误]', req.url, '->', options.target, ':', err.message)
          })
        }
      }
    }
  }
})
