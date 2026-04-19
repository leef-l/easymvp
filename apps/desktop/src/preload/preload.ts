import { contextBridge, ipcRenderer } from "electron";

type DesktopRuntimeInfo = {
  platform: string;
  version: string;
  packaged: boolean;
  launchMode: string;
  coreBaseUrl?: string;
  dataDirectory?: string;
  startupDiagnostics?: unknown[];
  startup_diagnostics?: unknown[];
  recoveryIssues?: unknown[];
  recovery_issues?: unknown[];
  coreDiagnostics?: unknown[];
  core_diagnostics?: unknown[];
  core?: {
    reachable: boolean;
    status: string;
    httpStatus: number;
    health?: Record<string, unknown>;
    error?: string;
    startupDiagnostics?: unknown[];
    startup_diagnostics?: unknown[];
    recoveryIssues?: unknown[];
    recovery_issues?: unknown[];
    diagnostics?: unknown[];
  };
  startup?: {
    phase: "idle" | "probing" | "starting" | "ready" | "failed";
    pending: boolean;
    managed: boolean;
    launchMode: string;
    baseUrl: string;
    startedAt: string;
    updatedAt: string;
    issue: {
      code: string;
      severity: "info" | "warning" | "error";
      summary: string;
      details: string[];
      actions: string[];
    } | null;
    lastProbe: {
      reachable: boolean;
      status: string;
      httpStatus: number;
      health?: Record<string, unknown>;
      error?: string;
    } | null;
  };
  coreManager?: {
    enabled: boolean;
    mode: "external" | "managed";
    status: "disabled" | "idle" | "starting" | "running" | "exited" | "failed";
    pid: number;
    command: string;
    args: string[];
    cwd: string;
    lastError: string;
    lastExitCode: number;
    logTail: string[];
  };
};

type DesktopSelectDirectoryResult = {
  canceled: boolean;
  path: string;
};

type DesktopShellResult = {
  ok: boolean;
  error?: string;
  path?: string;
  canceled?: boolean;
};

contextBridge.exposeInMainWorld("desktopBridge", {
  platform: process.platform,
  version: "0.1.0",
  coreBaseUrl: process.env.EASYMVP_CORE_BASE_URL ?? "http://127.0.0.1:8000",
  getRuntimeInfo: () =>
    ipcRenderer.invoke(
      "desktop:get-runtime-info",
    ) as Promise<DesktopRuntimeInfo>,
  selectDirectory: () =>
    ipcRenderer.invoke(
      "desktop:select-directory",
    ) as Promise<DesktopSelectDirectoryResult>,
  openPath: (targetPath: string) =>
    ipcRenderer.invoke(
      "desktop:open-path",
      targetPath,
    ) as Promise<DesktopShellResult>,
  saveDiagnosticExport: (payload: {
    suggestedName?: string;
    content: string;
  }) =>
    ipcRenderer.invoke(
      "desktop:save-diagnostic-export",
      payload,
    ) as Promise<DesktopShellResult>,
  showItemInFolder: (targetPath: string) =>
    ipcRenderer.invoke(
      "desktop:show-item-in-folder",
      targetPath,
    ) as Promise<DesktopShellResult>,
  relaunchSafeMode: () =>
    ipcRenderer.invoke(
      "desktop:relaunch-safe-mode",
    ) as Promise<DesktopShellResult>,
  relaunchNormalMode: () =>
    ipcRenderer.invoke(
      "desktop:relaunch-normal-mode",
    ) as Promise<DesktopShellResult>,
  startManagedCore: () =>
    ipcRenderer.invoke(
      "desktop:start-managed-core",
    ) as Promise<DesktopShellResult>,
  restartManagedCore: () =>
    ipcRenderer.invoke(
      "desktop:restart-managed-core",
    ) as Promise<DesktopShellResult>,
  restartCoreBootstrap: () =>
    ipcRenderer.invoke(
      "desktop:restart-core-bootstrap",
    ) as Promise<DesktopShellResult>,
});
