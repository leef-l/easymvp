import { defineConfig } from '@vben/vite-config';

export default defineConfig(async () => {
  return {
    application: {},
    vite: {
      server: {
        proxy: {
          '/api': {
            changeOrigin: true,
            // 后端 GoFrame 服务地址，不 rewrite，直接转发
            target: 'https://pw.easytestdev.online',
            ws: true,
          },
        },
      },
    },
  };
});
