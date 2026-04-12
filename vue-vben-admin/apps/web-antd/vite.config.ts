import { resolve } from 'node:path';

import { defineConfig } from '@vben/vite-config';

export default defineConfig(async () => {
  const systemProxyTarget =
    process.env.VITE_PROXY_SYSTEM_TARGET ?? 'http://localhost:41002';
  const aiProxyTarget =
    process.env.VITE_PROXY_AI_TARGET ?? 'http://localhost:41003';
  const mvpProxyTarget =
    process.env.VITE_PROXY_MVP_TARGET ?? 'http://localhost:41004';
  const verifyBuild = process.env.EASYMVP_WEB_ANTD_VERIFY_BUILD === '1';
  const workflowBundle = process.env.EASYMVP_WEB_ANTD_WORKFLOW_BUNDLE === '1';
  const bundleEntry = process.env.EASYMVP_WEB_ANTD_BUNDLE_ENTRY?.trim();
  const bundleOutDir = process.env.EASYMVP_WEB_ANTD_BUNDLE_OUT_DIR?.trim();
  const verifyEntry = workflowBundle
    ? 'src/verify/workflow-bundle.ts'
    : bundleEntry || 'src/main.ts';
  const isVerifyExternal = (id: string) => {
    if (
      id.startsWith('\0')
      || id.startsWith('.')
      || id.startsWith('/')
      || id.startsWith('#')
      || id.startsWith('virtual:')
      || id.startsWith('vite/')
    ) {
      return false;
    }
    return true;
  };

  const application = verifyBuild
    ? {
        archiver: false,
        compress: false,
        devtools: false,
        extraAppConfig: false,
        html: false,
        i18n: false,
        injectAppLoading: false,
        injectGlobalScss: false,
        injectMetadata: false,
        importmap: false,
        license: false,
        nitroMock: false,
        print: false,
        pwa: false,
        vxeTableLazyImport: false,
      }
    : {};
  const verifyBuildOptions = verifyBuild
    ? {
        copyPublicDir: false,
        cssMinify: false,
        emptyOutDir: true,
        minify: false,
        modulePreload: false,
        reportCompressedSize: false,
        rollupOptions: {
          external: isVerifyExternal,
          treeshake: false,
          output: {
            preserveModules: true,
            preserveModulesRoot: process.cwd(),
          },
        },
        target: 'esnext',
      }
    : {};
  const build = verifyBuild
    ? {
        ...verifyBuildOptions,
        lib: {
          entry: resolve(process.cwd(), verifyEntry),
          fileName: workflowBundle
            ? 'workflow-verify'
            : bundleEntry
              ? 'entry-verify'
              : 'app-verify',
          formats: ['es'],
        },
        outDir:
          bundleOutDir
          || (workflowBundle
            ? 'dist-workflow-verify'
            : bundleEntry
              ? 'dist-entry-verify'
              : 'dist-verify'),
      }
    : undefined;

  return {
    application,
    vite: {
      build,
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
