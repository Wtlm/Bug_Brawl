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
    port: 8080,
    open: true,
    historyApiFallback: true, 
  },
  plugins: [
    react(),
    tailwindcss(),
  ],
})
