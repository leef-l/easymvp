import { defineConfig } from '@vben/vite-config';

export default defineConfig(async () => {
  const systemProxyTarget =
    process.env.VITE_PROXY_SYSTEM_TARGET ?? 'http://localhost:9000';
  const aiProxyTarget =
    process.env.VITE_PROXY_AI_TARGET ?? 'http://localhost:9001';
  const mvpProxyTarget =
    process.env.VITE_PROXY_MVP_TARGET ?? 'http://localhost:9002';

  return {
    application: {},
    vite: {
      server: {
        proxy: {
          '/api/system': {
            changeOrigin: true,
            target: systemProxyTarget,
            ws: true,
          },
          '/api/ai': {
            changeOrigin: true,
            target: aiProxyTarget,
            ws: true,
          },
          '/api/mvp': {
            changeOrigin: true,
            target: mvpProxyTarget,
            ws: true,
          },
        },
      },
    },
  };
});
