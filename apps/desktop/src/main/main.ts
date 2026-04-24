import { app, BrowserWindow, dialog, ipcMain, shell } from "electron";
import { writeFile } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";
import {
  ensureManagedCore,
  getCoreBootstrapSnapshot,
  getCoreManagerSnapshot,
  restartCoreBootstrap,
  restartManagedCore,
  startCoreBootstrap,
  stopManagedCore,
  waitForCoreBootstrap,
} from "./coreManager.js";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const SMOKE_TEST_TIMEOUT_MS = 30000;

function resolveCoreBaseUrl() {
  return (process.env.EASYMVP_CORE_BASE_URL ?? "http://127.0.0.1:8000").replace(
    /\/+$/,
    "",
  );
}

function resolveLaunchMode() {
  return process.argv.includes("--safe-mode") ? "safe-mode" : "normal";
}

function isSmokeTestMode() {
  return process.argv.includes("--smoke-test");
}

function smokeLine(fields: Record<string, string | number | boolean>) {
  const text = Object.entries(fields)
    .map(([key, value]) => `${key}=${String(value)}`)
    .join(" ");
  console.log(`EASYMVP_SMOKE ${text}`);
}

let smokeExitStarted = false;
let quitAfterCoreStop = false;

function configureManagedCoreEnvironment() {
  if (!process.env.EASYMVP_CORE_DATA_ROOT?.trim()) {
    process.env.EASYMVP_CORE_DATA_ROOT = path.join(
      app.getPath("userData"),
      "core",
    );
  }
}

async function finishSmokeTest(code: number, reason: string) {
  if (smokeExitStarted) {
    return;
  }
  smokeExitStarted = true;
  const pid = getCoreManagerSnapshot().pid;
  if (pid > 0) {
    await stopManagedCore(`smoke-${reason}`);
  }
  smokeLine({
    cleanup: pid > 0 ? "stopped" : "none",
    corePid: pid,
    reason,
  });
  app.exit(code);
}

if (isSmokeTestMode()) {
  app.disableHardwareAcceleration();
  app.commandLine.appendSwitch("disable-gpu");
}

function buildRelaunchArgs(mode: "normal" | "safe-mode") {
  const filtered = process.argv.slice(1).filter((arg) => arg !== "--safe-mode");
  if (mode === "safe-mode") {
    filtered.push("--safe-mode");
  }
  return filtered;
}

function relaunchDesktop(mode: "normal" | "safe-mode") {
  app.relaunch({
    args: buildRelaunchArgs(mode),
  });
  void stopManagedCore("relaunch").finally(() => app.exit(0));
}

function logSmokeCoreDiagnostics() {
  const manager = getCoreManagerSnapshot();
  console.error(
    [
      "desktop smoke test core manager diagnostics:",
      `status=${manager.status}`,
      `pid=${manager.pid}`,
      `lastExitCode=${manager.lastExitCode}`,
      `command=${manager.command || "<empty>"}`,
      `cwd=${manager.cwd || "<empty>"}`,
      `lastError=${manager.lastError || "<empty>"}`,
      "logTail:",
      ...manager.logTail.map((line) => `  ${line}`),
    ].join("\n"),
  );
}

function registerDesktopBridgeHandlers() {
  ipcMain.handle("desktop:get-runtime-info", async () => {
    const baseUrl = resolveCoreBaseUrl();
    const launchMode = resolveLaunchMode();
    startCoreBootstrap({
      baseUrl,
      launchMode,
    });
    return {
      platform: process.platform,
      version: app.getVersion(),
      packaged: app.isPackaged,
      launchMode,
      coreBaseUrl: baseUrl,
      dataDirectory: app.getPath("userData"),
      startup: getCoreBootstrapSnapshot(),
      core: getCoreBootstrapSnapshot().lastProbe,
      coreManager: getCoreManagerSnapshot(),
    };
  });
  ipcMain.handle("desktop:select-directory", async () => {
    const targetWindow =
      BrowserWindow.getFocusedWindow() ?? BrowserWindow.getAllWindows()[0];
    const result = await dialog.showOpenDialog(targetWindow ?? undefined, {
      properties: ["openDirectory", "createDirectory"],
    });
    if (result.canceled || result.filePaths.length === 0) {
      return { canceled: true, path: "" };
    }
    return { canceled: false, path: result.filePaths[0] ?? "" };
  });
  ipcMain.handle("desktop:open-path", async (_event, targetPath?: string) => {
    const normalized = (targetPath ?? "").trim();
    if (!normalized) {
      return { ok: false, error: "path is required" };
    }
    const error = await shell.openPath(normalized);
    return {
      ok: error === "",
      error: error || "",
    };
  });
  ipcMain.handle(
    "desktop:save-diagnostic-export",
    async (_event, payload?: { suggestedName?: string; content?: string }) => {
      const targetWindow =
        BrowserWindow.getFocusedWindow() ?? BrowserWindow.getAllWindows()[0];
      const suggestedName =
        payload?.suggestedName?.trim() || "easymvp-diagnostics.json";
      const content = payload?.content ?? "";
      if (!content.trim()) {
        return { ok: false, error: "diagnostic export content is required" };
      }

      const result = await dialog.showSaveDialog(targetWindow ?? undefined, {
        title: "Export EasyMVP diagnostics",
        defaultPath: path.join(app.getPath("documents"), suggestedName),
        filters: [
          { name: "JSON", extensions: ["json"] },
          { name: "All Files", extensions: ["*"] },
        ],
      });
      if (result.canceled || !result.filePath) {
        return { ok: false, canceled: true, error: "diagnostic export canceled" };
      }

      await writeFile(result.filePath, content, "utf8");
      return { ok: true, error: "", path: result.filePath };
    },
  );
  ipcMain.handle(
    "desktop:show-item-in-folder",
    async (_event, targetPath?: string) => {
      const normalized = (targetPath ?? "").trim();
      if (!normalized) {
        return { ok: false, error: "path is required" };
      }
      shell.showItemInFolder(normalized);
      return { ok: true, error: "" };
    },
  );
  ipcMain.handle("desktop:relaunch-safe-mode", async () => {
    relaunchDesktop("safe-mode");
    return { ok: true, error: "" };
  });
  ipcMain.handle("desktop:relaunch-normal-mode", async () => {
    relaunchDesktop("normal");
    return { ok: true, error: "" };
  });
  ipcMain.handle("desktop:start-managed-core", async () => {
    const snapshot = await ensureManagedCore({
      baseUrl: resolveCoreBaseUrl(),
      launchMode: resolveLaunchMode(),
    });
    return {
      ok: snapshot.status === "running" || snapshot.status === "starting",
      error: snapshot.lastError,
    };
  });
  ipcMain.handle("desktop:restart-managed-core", async () => {
    const snapshot = await restartManagedCore({
      baseUrl: resolveCoreBaseUrl(),
      launchMode: resolveLaunchMode(),
    });
    return {
      ok: snapshot.status === "running" || snapshot.status === "starting",
      error: snapshot.lastError,
    };
  });
  ipcMain.handle("desktop:restart-core-bootstrap", async () => {
    const snapshot = await restartCoreBootstrap({
      baseUrl: resolveCoreBaseUrl(),
      launchMode: resolveLaunchMode(),
    });
    return {
      ok: snapshot.phase === "ready" || snapshot.phase === "starting",
      error: snapshot.issue?.summary || "",
    };
  });
}

function createWindow() {
  const smokeTestMode = isSmokeTestMode();
  const window = new BrowserWindow({
    width: 1440,
    height: 960,
    minWidth: 1180,
    minHeight: 760,
    backgroundColor: "#f3f4f6",
    title: "EasyMVP V3",
    show: !smokeTestMode,
    webPreferences: {
      preload: path.join(__dirname, "../preload/preload.js"),
      contextIsolation: true,
      nodeIntegration: false,
    },
  });

  if (smokeTestMode) {
    const timeout = setTimeout(() => {
      console.error("desktop smoke test timed out before renderer finished loading");
      void finishSmokeTest(1, "timeout");
    }, SMOKE_TEST_TIMEOUT_MS);

    window.webContents.once("did-fail-load", (_event, errorCode, errorDescription) => {
      clearTimeout(timeout);
      console.error(`desktop smoke test failed to load renderer: ${errorCode} ${errorDescription}`);
      void finishSmokeTest(1, "renderer-load-failed");
    });

    window.webContents.once("did-finish-load", async () => {
      try {
        let smokeState:
          | {
              pathname?: string;
              rootChildCount?: number;
              readyState?: string;
            }
          | null = null;
        let bootstrap = getCoreBootstrapSnapshot();
        for (let attempt = 0; attempt < 80; attempt += 1) {
          smokeState = await window.webContents.executeJavaScript(`({
            pathname: window.location.pathname,
            readyState: document.readyState,
            rootChildCount: document.getElementById("root")?.childElementCount || 0,
          })`);
          bootstrap = getCoreBootstrapSnapshot();
          if (
            bootstrap.lastProbe?.reachable
            && (smokeState?.rootChildCount || 0) > 0
          ) {
            break;
          }
          await new Promise((resolve) => setTimeout(resolve, 250));
        }
        clearTimeout(timeout);
        const manager = getCoreManagerSnapshot();
        smokeLine({
          managed: bootstrap.managed,
          phase: bootstrap.phase,
          baseUrl: bootstrap.baseUrl,
          corePid: manager.pid,
          managerStatus: manager.status,
        });
        smokeLine({
          healthOk: Boolean(bootstrap.lastProbe?.reachable),
          httpStatus: bootstrap.lastProbe?.httpStatus ?? 0,
          endpoint: "/api/v3/system/healthz",
        });
        smokeLine({
          rendererMounted: (smokeState?.rootChildCount || 0) > 0,
          rootChildCount: smokeState?.rootChildCount || 0,
          readyState: smokeState?.readyState || "unknown",
        });
        if (!bootstrap.lastProbe?.reachable) {
          console.error(`desktop smoke test failed: core not reachable (${bootstrap.lastProbe?.status || "unknown"})`);
          logSmokeCoreDiagnostics();
          void finishSmokeTest(1, "core-health-failed");
          return;
        }
        if ((smokeState?.rootChildCount || 0) <= 0) {
          console.error(
            `desktop smoke test failed: renderer did not mount (pathname=${smokeState?.pathname || "<empty>"}, readyState=${smokeState?.readyState || "unknown"})`,
          );
          void finishSmokeTest(1, "renderer-not-mounted");
          return;
        }
        setTimeout(() => {
          void finishSmokeTest(0, "ok");
        }, 500);
      } catch (error) {
        clearTimeout(timeout);
        console.error("desktop smoke test failed during renderer assertion", error);
        void finishSmokeTest(1, "assertion-error");
      }
    });
  }

  const devUrl = process.env.EASYMVP_DESKTOP_DEV_URL ?? "http://127.0.0.1:5173";

  if (app.isPackaged) {
    void window.loadFile(path.join(__dirname, "../../../dist/index.html"));
  } else {
    void window.loadURL(devUrl);
    window.webContents.openDevTools({ mode: "detach" });
  }
}

app.whenReady().then(async () => {
  configureManagedCoreEnvironment();
  registerDesktopBridgeHandlers();
  await waitForCoreBootstrap({
    baseUrl: resolveCoreBaseUrl(),
    launchMode: resolveLaunchMode(),
    timeoutMs: isSmokeTestMode() ? 22000 : 3000,
  });
  createWindow();

  app.on("activate", () => {
    if (BrowserWindow.getAllWindows().length === 0) {
      createWindow();
    }
  });
});

app.on("window-all-closed", () => {
  if (process.platform !== "darwin") {
    app.quit();
  }
});

app.on("before-quit", (event) => {
  if (quitAfterCoreStop || getCoreManagerSnapshot().pid <= 0) {
    return;
  }
  event.preventDefault();
  quitAfterCoreStop = true;
  void stopManagedCore("app-quit").finally(() => app.quit());
});
