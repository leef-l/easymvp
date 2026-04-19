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
  app.exit(0);
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
      app.exit(1);
    }, SMOKE_TEST_TIMEOUT_MS);

    window.webContents.once("did-fail-load", (_event, errorCode, errorDescription) => {
      clearTimeout(timeout);
      console.error(`desktop smoke test failed to load renderer: ${errorCode} ${errorDescription}`);
      app.exit(1);
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
        if (!bootstrap.lastProbe?.reachable) {
          console.error(`desktop smoke test failed: core not reachable (${bootstrap.lastProbe?.status || "unknown"})`);
          app.exit(1);
          return;
        }
        if ((smokeState?.rootChildCount || 0) <= 0) {
          console.error(
            `desktop smoke test failed: renderer did not mount (pathname=${smokeState?.pathname || "<empty>"}, readyState=${smokeState?.readyState || "unknown"})`,
          );
          app.exit(1);
          return;
        }
        setTimeout(() => app.exit(0), 500);
      } catch (error) {
        clearTimeout(timeout);
        console.error("desktop smoke test failed during renderer assertion", error);
        app.exit(1);
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
