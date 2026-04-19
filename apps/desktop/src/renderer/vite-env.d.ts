/// <reference types="vite/client" />

declare global {
  interface Window {
    desktopBridge?: {
      platform: string;
      version: string;
      coreBaseUrl?: string;
    };
  }
}

export {};
