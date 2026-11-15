import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import path from 'node:path'

export default defineConfig({
  plugins: [
    react(),
    tailwindcss(),
  ],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  // Static images already live in ../web/assets and are served directly by
  // the Go binary. The old ui_source/public copy contained Git LFS pointer
  // stubs and Vite would overwrite the real assets with them on every build.
  publicDir: false,
  build: {
    outDir: '../web',
    emptyOutDir: false,
  },
  server: {
    port: 5173,
    proxy: {
      '/api': 'http://127.0.0.1:18888',
    },
  },
})
