import { defineConfig } from "vitest/config";
import react from "@vitejs/plugin-react";
import path from "node:path";

export default defineConfig({
  plugins: [react()],
  test: {
    environment: "jsdom",
    globals: true,
    setupFiles: ["./src/test/setup.ts"],
    include: ["src/**/*.{test,spec}.{ts,tsx}"],
    exclude: ["node_modules", "dist", ".idea", ".git", ".cache"],
  },
  resolve: {
    conditions: ["development"],
    alias: {
      "@": path.resolve(__dirname, "src/renderer"),
      "react/jsx-runtime": path.resolve(__dirname, "node_modules/react/cjs/react-jsx-runtime.development.js"),
      "react/jsx-dev-runtime": path.resolve(__dirname, "node_modules/react/cjs/react-jsx-dev-runtime.development.js"),
    },
  },
});
