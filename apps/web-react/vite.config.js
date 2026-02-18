import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

const proxyTarget = process.env.VITE_BACKEND_PROXY_TARGET || 'http://localhost:8080';

export default defineConfig({
  plugins: [react()],
  server: {
    host: '0.0.0.0',
    port: 5173,
    proxy: {
      '/query': proxyTarget,
      '/healthz': proxyTarget
    }
  }
});
