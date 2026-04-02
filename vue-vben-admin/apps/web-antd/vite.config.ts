import { defineConfig } from '@vben/vite-config';

export default defineConfig(async () => {
  return {
    application: {},
    vite: {
      server: {
        proxy: {
          '/api/system': {
            changeOrigin: true,
            target: 'http://localhost:9000',
            ws: true,
          },
          '/api/ai': {
            changeOrigin: true,
            target: 'http://localhost:9001',
            ws: true,
          },
          '/api/mvp': {
            changeOrigin: true,
            target: 'http://localhost:9002',
            ws: true,
          },
        },
      },
    },
  };
});
