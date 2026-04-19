import { spawn, type ChildProcessWithoutNullStreams } from "node:child_process";
import fs from "node:fs";
import path from "node:path";

type CoreManagerStatus =
  | "disabled"
  | "idle"
  | "starting"
  | "running"
  | "exited"
  | "failed";

type CoreManagerMode = "external" | "managed";
export type CoreBootstrapPhase =
  | "idle"
  | "probing"
  | "starting"
  | "ready"
  | "failed";

export type CoreStartupIssue = {
  code: string;
  severity: "info" | "warning" | "error";
  summary: string;
  details: string[];
  actions: string[];
};

export type CoreHealthProbe = {
  reachable: boolean;
  status: string;
  httpStatus: number;
  health?: Record<string, unknown>;
  error?: string;
};

export type CoreManagerSnapshot = {
  enabled: boolean;
  mode: CoreManagerMode;
  status: CoreManagerStatus;
  pid: number;
  command: string;
  args: string[];
  cwd: string;
  lastError: string;
  lastExitCode: number;
  logTail: string[];
  startupIssue: CoreStartupIssue | null;
};

export type CoreBootstrapSnapshot = {
  phase: CoreBootstrapPhase;
  pending: boolean;
  issue: CoreStartupIssue | null;
  managed: boolean;
  launchMode: string;
  baseUrl: string;
  startedAt: string;
  updatedAt: string;
  lastProbe: CoreHealthProbe | null;
};

type LaunchSpec = {
  command: string;
  args: string[];
  cwd: string;
};

const MAX_LOG_LINES = 24;
const LOG_LINE_LENGTH_LIMIT = 280;
const MANAGED_CORE_PROBE_ATTEMPTS = 24;
const MANAGED_CORE_PROBE_INTERVAL_MS = 750;

let childProcess: ChildProcessWithoutNullStreams | null = null;
let startPromise: Promise<CoreManagerSnapshot> | null = null;
let bootstrapPromise: Promise<CoreBootstrapSnapshot> | null = null;
const logTail: string[] = [];

const state: CoreManagerSnapshot = {
  enabled: false,
  mode: "external",
  status: "disabled",
  pid: 0,
  command: "",
  args: [],
  cwd: "",
  lastError: "",
  lastExitCode: 0,
  logTail,
  startupIssue: null,
};

const bootstrapState: CoreBootstrapSnapshot = {
  phase: "idle",
  pending: false,
  issue: null,
  managed: false,
  launchMode: "normal",
  baseUrl: "",
  startedAt: "",
  updatedAt: "",
  lastProbe: null,
};

function normalizeBaseUrl(baseUrl: string) {
  return baseUrl.trim().replace(/\/+$/, "");
}

function parsePort(baseUrl: string) {
  try {
    const url = new URL(baseUrl);
    const protocolPort = url.protocol === "https:" ? "443" : "80";
    return Number(url.port || protocolPort);
  } catch {
    return 8000;
  }
}

function pushLog(line: string) {
  const normalized = line.trim();
  if (!normalized) {
    return;
  }
  const clipped =
    normalized.length > LOG_LINE_LENGTH_LIMIT
      ? `${normalized.slice(0, LOG_LINE_LENGTH_LIMIT)}...`
      : normalized;
  logTail.push(clipped);
  while (logTail.length > MAX_LOG_LINES) {
    logTail.shift();
  }
}

function nowIso() {
  return new Date().toISOString();
}

function setBootstrapPhase(
  phase: CoreBootstrapPhase,
  params?: {
    issue?: CoreStartupIssue | null;
    launchMode?: string;
    baseUrl?: string;
    managed?: boolean;
    lastProbe?: CoreHealthProbe | null;
  },
) {
  if (!bootstrapState.startedAt || bootstrapState.phase === "idle") {
    bootstrapState.startedAt = nowIso();
  }
  bootstrapState.phase = phase;
  bootstrapState.pending = phase === "probing" || phase === "starting";
  if (params?.issue !== undefined) {
    bootstrapState.issue = params.issue;
  }
  if (params?.launchMode !== undefined) {
    bootstrapState.launchMode = params.launchMode;
  }
  if (params?.baseUrl !== undefined) {
    bootstrapState.baseUrl = params.baseUrl;
  }
  if (params?.managed !== undefined) {
    bootstrapState.managed = params.managed;
  }
  if (params?.lastProbe !== undefined) {
    bootstrapState.lastProbe = params.lastProbe;
  }
  bootstrapState.updatedAt = nowIso();
}

function attachLogStream(
  stream: NodeJS.ReadableStream,
  prefix: "stdout" | "stderr",
) {
  let buffer = "";
  stream.on("data", (chunk: Buffer | string) => {
    buffer += chunk.toString();
    const lines = buffer.split(/\r?\n/);
    buffer = lines.pop() ?? "";
    for (const line of lines) {
      pushLog(`[${prefix}] ${line}`);
    }
  });
}

function getRepoRoot() {
  return path.resolve(__dirname, "../../../../");
}

function getDevCoreCwd() {
  return path.join(getRepoRoot(), "apps/core");
}

function getPackagedCoreBinaryPath() {
  const fileName =
    process.platform === "win32" ? "easymvp-core.exe" : "easymvp-core";
  return path.join(process.resourcesPath, "bin", fileName);
}

function stripWrappedQuotes(value: string) {
  const trimmed = value.trim();
  if (trimmed.length < 2) {
    return trimmed;
  }
  const first = trimmed[0];
  const last = trimmed[trimmed.length - 1];
  if ((first === `"` && last === `"`) || (first === `'` && last === `'`)) {
    return trimmed.slice(1, -1).trim();
  }
  return trimmed;
}

function splitArgs(argsText?: string) {
  const source = (argsText ?? "").trim();
  if (!source) {
    return [];
  }

  const args: string[] = [];
  let current = "";
  let quote: `"` | `'` | "" = "";

  for (let index = 0; index < source.length; index += 1) {
    const char = source[index];

    if (quote) {
      if (char === quote) {
        quote = "";
      } else if (
        char === "\\"
        && quote === `"`
        && index + 1 < source.length
        && (`"\\`.includes(source[index + 1] ?? ""))
      ) {
        index += 1;
        current += source[index] ?? "";
      } else {
        current += char;
      }
      continue;
    }

    if (char === `"` || char === `'`) {
      quote = char;
      continue;
    }

    if (/\s/.test(char)) {
      if (current) {
        args.push(current);
        current = "";
      }
      continue;
    }

    if (
      char === "\\"
      && index + 1 < source.length
      && (`"'\\`.includes(source[index + 1] ?? ""))
    ) {
      index += 1;
      current += source[index] ?? "";
      continue;
    }

    current += char;
  }

  if (current) {
    args.push(current);
  }

  return args;
}

function summarizeSpawnError(error: string) {
  const normalized = error.toLowerCase();
  if (normalized.includes("enoent")) {
    return {
      code: "core_command_missing",
      severity: "error" as const,
      summary: "Managed core command is missing or not executable.",
      actions: [
        "Verify EASYMVP_CORE_CMD points to a valid binary, or package easymvp-core into resources/bin.",
        "In development mode, enable EASYMVP_MANAGE_CORE=1 only when local Go is available.",
      ],
    };
  }
  return {
    code: "managed_core_spawn_failed",
    severity: "error" as const,
    summary: "Managed core process failed during startup.",
    actions: [
      "Inspect the recent core log tail in Recovery.",
      "Verify the configured command, working directory, and local runtime dependencies.",
    ],
  };
}

function collectStartupHealthDiagnostics(health?: Record<string, unknown>) {
  const startup =
    health && typeof health === "object" && "startup" in health
      ? (health.startup as Record<string, unknown> | undefined)
      : undefined;
  const diagnostics =
    startup && typeof startup === "object" && "diagnostics" in startup
      ? startup.diagnostics
      : undefined;
  return Array.isArray(diagnostics)
    ? diagnostics.filter((item) => item && typeof item === "object")
    : [];
}

function detectClassifiedStartupIssue(text: string) {
  const normalized = text.toLowerCase();
  if (normalized.includes("migration")) {
    return {
      code: "migration_failure",
      severity: "error" as const,
      summary: "Core startup is blocked by a migration failure.",
      actions: [
        "Inspect migration diagnostics and verify the packaged migration path.",
        "Fix the schema or migration directory problem before restarting the managed core.",
      ],
    };
  }
  if (
    normalized.includes("directory is not writable")
    || normalized.includes("permission denied")
    || normalized.includes("read-only file system")
    || normalized.includes("data root")
    || normalized.includes("db_path")
  ) {
    return {
      code: "data_directory_unwritable",
      severity: "error" as const,
      summary: "Core startup cannot write to the configured data directory.",
      actions: [
        "Open the desktop data folder and verify permissions, free space, and path ownership.",
        "Adjust the configured data root or relaunch in safe mode after fixing filesystem access.",
      ],
    };
  }
  if (
    normalized.includes("eaddrinuse")
    || normalized.includes("address already in use")
  ) {
    return {
      code: "core_port_in_use",
      severity: "warning" as const,
      summary: "Core startup could not bind the configured port.",
      actions: [
        "Stop the conflicting process or change the configured core port.",
        "Retry managed core startup after the port is free.",
      ],
    };
  }
  return null;
}

function classifyStartupIssue(params: {
  baseUrl: string;
  health: CoreHealthProbe;
}): CoreStartupIssue {
  const { baseUrl, health } = params;
  const healthDiagnostics = collectStartupHealthDiagnostics(health.health);
  const healthSummary = JSON.stringify(health.health ?? {});
  const classifiedFromHealth = detectClassifiedStartupIssue(healthSummary);
  if (classifiedFromHealth) {
    return {
      code: classifiedFromHealth.code,
      severity: classifiedFromHealth.severity,
      summary: classifiedFromHealth.summary,
      details: [
        `Configured base URL: ${baseUrl}`,
        health.error || "Health snapshot exposed a blocking startup diagnostic.",
      ],
      actions: classifiedFromHealth.actions,
    };
  }

  if (!state.enabled) {
    return {
      code: "external_core_unreachable",
      severity: "warning",
      summary: "Desktop shell could not reach an external core service.",
      details: [
        `Configured base URL: ${baseUrl}`,
        health.error || "Health probe did not return a successful response.",
      ],
      actions: [
        "Start apps/core manually, or configure managed core startup for this desktop shell.",
        "Verify the port and host match the current core instance.",
      ],
    };
  }

  if (state.status === "starting") {
    return {
      code: "managed_core_starting",
      severity: "info",
      summary: "Managed core is starting but has not reported healthy yet.",
      details: [
        `Configured base URL: ${baseUrl}`,
        state.command
          ? `Launch command: ${state.command}`
          : "Launch command is unavailable.",
      ],
      actions: [
        "Wait a moment and retry the health probe.",
        "If startup stalls, restart the managed core and inspect the recent log tail.",
      ],
    };
  }

  if (state.lastError) {
    const classified = detectClassifiedStartupIssue(
      `${state.lastError}\n${logTail.join("\n")}`,
    );
    if (classified) {
      return {
        code: classified.code,
        severity: classified.severity,
        summary: classified.summary,
        details: [
          `Configured base URL: ${baseUrl}`,
          state.command
            ? `Launch command: ${state.command}`
            : "Launch command is unavailable.",
          state.lastError,
        ],
        actions: classified.actions,
      };
    }
    const mapped = summarizeSpawnError(state.lastError);
    return {
      code: mapped.code,
      severity: mapped.severity,
      summary: mapped.summary,
      details: [
        `Configured base URL: ${baseUrl}`,
        state.command
          ? `Launch command: ${state.command}`
          : "Launch command is unavailable.",
        state.lastError,
      ],
      actions: mapped.actions,
    };
  }

  if (health.error?.toLowerCase().includes("econnrefused")) {
    return {
      code: "core_connection_refused",
      severity: "warning",
      summary:
        "Core port is reachable on the network path but refusing connections.",
      details: [`Configured base URL: ${baseUrl}`, health.error],
      actions: [
        "Check whether another process owns the port or the core exited immediately after launch.",
        "Restart the managed core and inspect the recent log tail for port bind or migration errors.",
      ],
    };
  }

  if (healthDiagnostics.length > 0) {
    const item = healthDiagnostics[0] as Record<string, unknown>;
    return {
      code:
        (typeof item.code === "string" && item.code.trim()) ||
        "core_startup_attention",
      severity:
        typeof item.severity === "string" && item.severity.trim() === "error"
          ? "error"
          : "warning",
      summary:
        (typeof item.summary === "string" && item.summary.trim()) ||
        "Core startup reported a structured diagnostic.",
      details: [
        `Configured base URL: ${baseUrl}`,
        typeof item.detail === "string" ? item.detail : "",
      ].filter(Boolean),
      actions: Array.isArray(item.actions)
        ? item.actions
            .map((entry) => (typeof entry === "string" ? entry.trim() : ""))
            .filter(Boolean)
        : [
            "Inspect the startup diagnostics and recent log tail in Recovery.",
            "Retry the health probe after addressing the reported startup warning.",
          ],
    };
  }

  return {
    code: "core_health_unreachable",
    severity: "warning",
    summary: "Core health endpoint is still unreachable.",
    details: [
      `Configured base URL: ${baseUrl}`,
      health.error || "Health probe did not return a successful response.",
    ],
    actions: [
      "Retry the health probe after confirming the current launch mode and base URL.",
      "Use Recovery to restart managed core or switch to safe mode for further triage.",
    ],
  };
}

function resolveLaunchSpec(
  baseUrl: string,
  launchMode: string,
): LaunchSpec | null {
  const commandFromEnv = process.env.EASYMVP_CORE_CMD?.trim();
  if (commandFromEnv) {
    return {
      command: stripWrappedQuotes(commandFromEnv),
      args: splitArgs(process.env.EASYMVP_CORE_ARGS),
      cwd: process.env.EASYMVP_CORE_CWD?.trim() || process.cwd(),
    };
  }

  const packagedBinary = getPackagedCoreBinaryPath();
  if (fs.existsSync(packagedBinary)) {
    const args = [`--port=${parsePort(baseUrl)}`];
    if (launchMode === "safe-mode") {
      args.push("--safe-mode");
    }
    return {
      command: packagedBinary,
      args,
      cwd: process.resourcesPath,
    };
  }

  if ((process.env.EASYMVP_MANAGE_CORE ?? "").trim() !== "1") {
    return null;
  }

  const args = ["run", "main.go", `--port=${parsePort(baseUrl)}`];
  if (launchMode === "safe-mode") {
    args.push("--safe-mode");
  }
  return {
    command: "go",
    args,
    cwd: getDevCoreCwd(),
  };
}

export async function probeCoreHealth(
  baseUrl: string,
): Promise<CoreHealthProbe> {
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), 1500);
  try {
    const response = await fetch(
      `${normalizeBaseUrl(baseUrl)}/api/v3/system/healthz`,
      {
        method: "GET",
        signal: controller.signal,
        headers: {
          Accept: "application/json",
        },
      },
    );
    const text = await response.text();
    let payload: unknown = null;
    if (text.trim() !== "") {
      payload = JSON.parse(text);
    }
    const data =
      payload && typeof payload === "object" && "data" in payload
        ? (payload as { data?: Record<string, unknown> }).data
        : (payload as Record<string, unknown> | null);
    return {
      reachable: response.ok,
      status: response.ok ? "ready" : "degraded",
      httpStatus: response.status,
      health: data ?? undefined,
      error: response.ok
        ? undefined
        : `Core health probe failed with HTTP ${response.status}`,
    };
  } catch (error) {
    return {
      reachable: false,
      status: "unreachable",
      httpStatus: 0,
      health: undefined,
      error:
        error instanceof Error ? error.message : "Core health probe failed",
    };
  } finally {
    clearTimeout(timeout);
  }
}

export function getCoreManagerSnapshot(): CoreManagerSnapshot {
  return {
    ...state,
    args: [...state.args],
    logTail: [...logTail],
    startupIssue: state.startupIssue
      ? {
          ...state.startupIssue,
          details: [...state.startupIssue.details],
          actions: [...state.startupIssue.actions],
        }
      : null,
  };
}

export function getCoreBootstrapSnapshot(): CoreBootstrapSnapshot {
  return {
    ...bootstrapState,
    issue: bootstrapState.issue
      ? {
          ...bootstrapState.issue,
          details: [...bootstrapState.issue.details],
          actions: [...bootstrapState.issue.actions],
        }
      : null,
    lastProbe: bootstrapState.lastProbe
      ? {
          ...bootstrapState.lastProbe,
          health: bootstrapState.lastProbe.health
            ? { ...bootstrapState.lastProbe.health }
            : undefined,
        }
      : null,
  };
}

export async function ensureManagedCore(params: {
  baseUrl: string;
  launchMode: string;
}): Promise<CoreManagerSnapshot> {
  const { baseUrl, launchMode } = params;
  const spec = resolveLaunchSpec(baseUrl, launchMode);
  state.enabled = Boolean(spec);
  state.mode = spec ? "managed" : "external";

  if (!spec) {
    state.status = "disabled";
    state.command = "";
    state.args = [];
    state.cwd = "";
    state.startupIssue = null;
    return getCoreManagerSnapshot();
  }

  if (childProcess && childProcess.exitCode === null) {
    state.status = "running";
    state.pid = childProcess.pid ?? 0;
    return getCoreManagerSnapshot();
  }

  if (startPromise) {
    return startPromise;
  }

  startPromise = (async () => {
    const health = await probeCoreHealth(baseUrl);
    if (health.reachable) {
      state.status = "running";
      state.lastError = "";
      state.startupIssue = null;
      return getCoreManagerSnapshot();
    }

    state.status = "starting";
    state.command = spec.command;
    state.args = [...spec.args];
    state.cwd = spec.cwd;
    state.lastError = health.error ?? "";
    pushLog(
      `Starting managed core: ${spec.command} ${spec.args.join(" ")}`.trim(),
    );

    const child = spawn(spec.command, spec.args, {
      cwd: spec.cwd,
      env: {
        ...process.env,
        EASYMVP_CORE_BASE_URL: normalizeBaseUrl(baseUrl),
      },
      stdio: "pipe",
    });
    childProcess = child;
    state.pid = child.pid ?? 0;

    attachLogStream(child.stdout, "stdout");
    attachLogStream(child.stderr, "stderr");

    child.once("error", (error) => {
      state.status = "failed";
      state.lastError = error.message;
      state.pid = 0;
      pushLog(`[manager] ${error.message}`);
      state.startupIssue = classifyStartupIssue({
        baseUrl,
        health: {
          reachable: false,
          status: "unreachable",
          httpStatus: 0,
          error: error.message,
        },
      });
      childProcess = null;
    });

    child.once("exit", (code, signal) => {
      state.status = code === 0 ? "exited" : "failed";
      state.lastExitCode = code ?? 0;
      state.lastError =
        signal && !state.lastError
          ? `Managed core exited from signal ${signal}`
          : state.lastError;
      state.pid = 0;
      pushLog(
        `[manager] core exited code=${code ?? "null"} signal=${signal ?? "none"}`,
      );
      childProcess = null;
    });

    for (let attempt = 0; attempt < MANAGED_CORE_PROBE_ATTEMPTS; attempt += 1) {
      await new Promise((resolve) =>
        setTimeout(resolve, MANAGED_CORE_PROBE_INTERVAL_MS),
      );
      const probe = await probeCoreHealth(baseUrl);
      if (probe.reachable) {
        state.status = "running";
        state.lastError = "";
        state.startupIssue = null;
        return getCoreManagerSnapshot();
      }
      if (child.exitCode !== null) {
        break;
      }
    }

    state.status = child.exitCode === null ? "starting" : "failed";
    if (!state.lastError) {
      state.lastError =
        health.error || "Managed core did not become healthy in time";
    }
    state.startupIssue = classifyStartupIssue({
      baseUrl,
      health: {
        reachable: false,
        status: child.exitCode === null ? "starting" : "unreachable",
        httpStatus: 0,
        error: state.lastError,
      },
    });
    return getCoreManagerSnapshot();
  })();

  try {
    return await startPromise;
  } finally {
    startPromise = null;
  }
}

async function runCoreBootstrap(params: {
  baseUrl: string;
  launchMode: string;
}): Promise<CoreBootstrapSnapshot> {
  const { baseUrl, launchMode } = params;
  const spec = resolveLaunchSpec(baseUrl, launchMode);
  setBootstrapPhase("probing", {
    issue: null,
    launchMode,
    baseUrl,
    managed: Boolean(spec),
  });

  const initialProbe = await probeCoreHealth(baseUrl);
  setBootstrapPhase(initialProbe.reachable ? "ready" : "probing", {
    lastProbe: initialProbe,
    issue: initialProbe.reachable
      ? null
      : classifyStartupIssue({
          baseUrl,
          health: initialProbe,
        }),
    launchMode,
    baseUrl,
    managed: Boolean(spec),
  });

  if (initialProbe.reachable) {
    state.startupIssue = null;
    return getCoreBootstrapSnapshot();
  }

  if (!spec) {
    const issue = classifyStartupIssue({
      baseUrl,
      health: initialProbe,
    });
    state.startupIssue = issue;
    setBootstrapPhase("failed", {
      issue,
      lastProbe: initialProbe,
      launchMode,
      baseUrl,
      managed: false,
    });
    return getCoreBootstrapSnapshot();
  }

  setBootstrapPhase("starting", {
    issue: classifyStartupIssue({
      baseUrl,
      health: initialProbe,
    }),
    lastProbe: initialProbe,
    launchMode,
    baseUrl,
    managed: true,
  });

  const snapshot = await ensureManagedCore(params);
  let probe = await probeCoreHealth(baseUrl);
  for (let attempt = 0; attempt < 6 && !probe.reachable; attempt += 1) {
    await new Promise((resolve) => setTimeout(resolve, 500));
    probe = await probeCoreHealth(baseUrl);
  }
  const issue = probe.reachable
    ? null
    : (snapshot.startupIssue ??
      classifyStartupIssue({
        baseUrl,
        health: probe,
      }));

  setBootstrapPhase(probe.reachable ? "ready" : "failed", {
    issue,
    lastProbe: probe,
    launchMode,
    baseUrl,
    managed: true,
  });
  return getCoreBootstrapSnapshot();
}

export function startCoreBootstrap(params: {
  baseUrl: string;
  launchMode: string;
}): Promise<CoreBootstrapSnapshot> {
  if (!bootstrapPromise) {
    bootstrapPromise = runCoreBootstrap(params).finally(() => {
      bootstrapPromise = null;
    });
  }
  return bootstrapPromise;
}

export async function waitForCoreBootstrap(params: {
  baseUrl: string;
  launchMode: string;
  timeoutMs?: number;
}): Promise<CoreBootstrapSnapshot> {
  const { timeoutMs = 1800, ...bootstrapParams } = params;
  const bootstrap = startCoreBootstrap(bootstrapParams);
  if (timeoutMs <= 0) {
    return bootstrap;
  }
  const timeout = new Promise<CoreBootstrapSnapshot>((resolve) => {
    setTimeout(() => {
      resolve(getCoreBootstrapSnapshot());
    }, timeoutMs);
  });
  return Promise.race([bootstrap, timeout]);
}

export async function restartManagedCore(params: {
  baseUrl: string;
  launchMode: string;
}): Promise<CoreManagerSnapshot> {
  if (childProcess && childProcess.exitCode === null) {
    childProcess.kill();
    childProcess = null;
    state.pid = 0;
    state.status = "idle";
    state.startupIssue = null;
    pushLog("[manager] stopping existing managed core before restart");
    await new Promise((resolve) => setTimeout(resolve, 250));
  }
  return ensureManagedCore(params);
}

export async function restartCoreBootstrap(params: {
  baseUrl: string;
  launchMode: string;
}): Promise<CoreBootstrapSnapshot> {
  await restartManagedCore(params);
  return runCoreBootstrap(params);
}
