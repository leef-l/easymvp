import { contextBridge } from "electron";

contextBridge.exposeInMainWorld("desktopBridge", {
  platform: process.platform,
  version: "0.1.0",
  coreBaseUrl: process.env.EASYMVP_CORE_BASE_URL ?? "http://127.0.0.1:8000",
});
