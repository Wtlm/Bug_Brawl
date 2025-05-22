import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

// https://vite.dev/config/
export default defineConfig({
  root: './', 
  build: {
    outDir: 'dist',
  },
  server: {
    host: true,
    port: 5173,
    open: true,
    historyApiFallback: true,
    proxy: {
      '/ws': {
        target: 'ws://localhost:8080',
        ws: true,
      },
    },
  },
  plugins: [
    react(),
    tailwindcss(),
  ],
})
