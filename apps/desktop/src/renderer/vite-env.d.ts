/// <reference types="vite/client" />

declare global {
  interface Window {
    desktopBridge?: {
      platform: string;
      version: string;
      coreBaseUrl?: string;
      getRuntimeInfo?: () => Promise<{
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
          status:
            | "disabled"
            | "idle"
            | "starting"
            | "running"
            | "exited"
            | "failed";
          pid: number;
          command: string;
          args: string[];
          cwd: string;
          lastError: string;
          lastExitCode: number;
          logTail: string[];
        };
      }>;
      dataDirectory?: string;
      userDataPath?: string;
      selectDirectory?: () => Promise<{
        canceled: boolean;
        path: string;
      }>;
      openPath?: (targetPath: string) => Promise<{
        ok: boolean;
        error?: string;
        path?: string;
        canceled?: boolean;
      }>;
      saveDiagnosticExport?: (payload: {
        suggestedName?: string;
        content: string;
      }) => Promise<{
        ok: boolean;
        error?: string;
        path?: string;
        canceled?: boolean;
      }>;
      showItemInFolder?: (targetPath: string) => Promise<{
        ok: boolean;
        error?: string;
        path?: string;
        canceled?: boolean;
      }>;
      relaunchSafeMode?: () => Promise<{
        ok: boolean;
        error?: string;
      }>;
      relaunchNormalMode?: () => Promise<{
        ok: boolean;
        error?: string;
      }>;
      startManagedCore?: () => Promise<{
        ok: boolean;
        error?: string;
      }>;
      restartManagedCore?: () => Promise<{
        ok: boolean;
        error?: string;
      }>;
      restartCoreBootstrap?: () => Promise<{
        ok: boolean;
        error?: string;
        path?: string;
        canceled?: boolean;
      }>;
    };
  }
}

export {};
