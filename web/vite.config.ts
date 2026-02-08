import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      // 代理所有 /api 请求到后端服务器
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        // 不需要 rewrite，因为后端也是 /api 开头
      }
    }
  }
})
