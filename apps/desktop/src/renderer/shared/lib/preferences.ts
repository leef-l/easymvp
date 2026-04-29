const PROJECT_ID_KEY = "easymvp.desktop.projectId";
const CORE_BASE_URL_KEY = "easymvp.desktop.coreBaseUrl";

const DEFAULT_PROJECT_ID = "project-demo";
const DEFAULT_CORE_BASE_URL = "";
const DEFAULT_RUNTIME_PLATFORM = "unknown";
const DEFAULT_RUNTIME_VERSION = "dev";
const DEFAULT_LAUNCH_MODE = "unknown";

export type DesktopRuntimeInfo = {
  platform: string;
  version: string;
  packaged: boolean;
  launchMode: string;
  coreBaseUrl: string;
  dataDirectory: string;
  coreStatus: string;
  coreReachable: boolean;
  coreHttpStatus: number;
  coreError: string;
  coreManagerEnabled: boolean;
  coreManagerMode: "external" | "managed";
  coreManagerStatus:
    | "disabled"
    | "idle"
    | "starting"
    | "running"
    | "exited"
    | "failed";
  coreManagerPid: number;
  coreManagerCommand: string;
  coreManagerArgs: string[];
  coreManagerCwd: string;
  coreManagerLastError: string;
  coreManagerLastExitCode: number;
  coreManagerLogTail: string[];
  bridgeAvailable: boolean;
  source: "bridge-runtime" | "bridge-static" | "defaults";
  issue: string;
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
  core?: {
    reachable: boolean;
    status: string;
    httpStatus: number;
    health?: Record<string, unknown>;
    error?: string;
  };
};

export type DesktopRecoverySeverity = "info" | "warning" | "error" | "critical";

export type DesktopRecoveryAction = {
  id: string;
  label: string;
  description: string;
};

export type DesktopRecoveryIssue = {
  code: string;
  category:
    | "migration_failure"
    | "data_directory_unwritable"
    | "core_unavailable"
    | "policy_denied"
    | "verification_conflict"
    | "fault_loop_detected"
    | "runtime_attention";
  severity: DesktopRecoverySeverity;
  summary: string;
  detail: string;
  source: "bridge" | "health-probe" | "core-manager" | "startup-diagnostic";
  mode: "managed" | "external" | "unknown";
  evidence: string[];
  actions: DesktopRecoveryAction[];
};

export type DesktopRuntimeDiagnosis = {
  status: "healthy" | "degraded" | "failed";
  summary: string;
  mode: "managed" | "external" | "unknown";
  managerState:
    | "disabled"
    | "idle"
    | "starting"
    | "running"
    | "exited"
    | "failed";
  issues: DesktopRecoveryIssue[];
};

type DesktopShellResult = {
  ok: boolean;
  error?: string;
  path?: string;
  canceled?: boolean;
};

type DesktopSelectDirectoryResult = {
  canceled: boolean;
  path: string;
};

function normalizeBaseUrl(baseUrl?: string) {
  return baseUrl?.trim().replace(/\/+$/, "") || DEFAULT_CORE_BASE_URL;
}

function getDesktopRuntimeSeed(): DesktopRuntimeInfo {
  const bridge =
    typeof window === "undefined" ? undefined : window.desktopBridge;

  return {
    platform: bridge?.platform?.trim() || DEFAULT_RUNTIME_PLATFORM,
    version: bridge?.version?.trim() || DEFAULT_RUNTIME_VERSION,
    packaged: false,
    launchMode: DEFAULT_LAUNCH_MODE,
    coreBaseUrl: normalizeBaseUrl(bridge?.coreBaseUrl),
    dataDirectory: "",
    coreStatus: "unknown",
    coreReachable: false,
    coreHttpStatus: 0,
    coreError: "",
    coreManagerEnabled: false,
    coreManagerMode: "external",
    coreManagerStatus: "disabled",
    coreManagerPid: 0,
    coreManagerCommand: "",
    coreManagerArgs: [],
    coreManagerCwd: "",
    coreManagerLastError: "",
    coreManagerLastExitCode: 0,
    coreManagerLogTail: [],
    bridgeAvailable: Boolean(bridge),
    source: bridge ? "bridge-static" : "defaults",
    issue: bridge ? "" : "desktop bridge unavailable in renderer",
    startup: undefined,
    core: undefined,
  };
}

export function getStoredProjectId() {
  if (typeof window === "undefined") {
    return "";
  }
  return window.localStorage.getItem(PROJECT_ID_KEY)?.trim() ?? "";
}

export function setStoredProjectId(projectId: string) {
  if (typeof window === "undefined") {
    return;
  }
  const normalized = projectId.trim();
  if (normalized === "") {
    window.localStorage.removeItem(PROJECT_ID_KEY);
  } else {
    window.localStorage.setItem(PROJECT_ID_KEY, normalized);
  }
}

export function getStoredCoreBaseUrl() {
  if (typeof window === "undefined") {
    return DEFAULT_CORE_BASE_URL;
  }
  const fromStorage = window.localStorage.getItem(CORE_BASE_URL_KEY)?.trim();
  if (fromStorage) {
    return normalizeBaseUrl(fromStorage);
  }
  const fromBridge = window.desktopBridge?.coreBaseUrl?.trim();
  if (fromBridge) {
    return normalizeBaseUrl(fromBridge);
  }
  return DEFAULT_CORE_BASE_URL;
}

export function setStoredCoreBaseUrl(baseUrl: string) {
  if (typeof window === "undefined") {
    return;
  }
  const normalized = baseUrl.trim() || DEFAULT_CORE_BASE_URL;
  window.localStorage.setItem(CORE_BASE_URL_KEY, normalizeBaseUrl(normalized));
}

export async function getDesktopRuntimeInfo(): Promise<DesktopRuntimeInfo> {
  const seed = getDesktopRuntimeSeed();
  const runtimeLoader =
    typeof window === "undefined"
      ? undefined
      : window.desktopBridge?.getRuntimeInfo;

  if (!runtimeLoader) {
    return seed;
  }

  try {
    const runtimeInfo = await runtimeLoader();
    return {
      platform: runtimeInfo.platform?.trim() || seed.platform,
      version: runtimeInfo.version?.trim() || seed.version,
      packaged: Boolean(runtimeInfo.packaged),
      launchMode: runtimeInfo.launchMode?.trim() || seed.launchMode,
      coreBaseUrl: normalizeBaseUrl(
        runtimeInfo.coreBaseUrl || seed.coreBaseUrl,
      ),
      dataDirectory: runtimeInfo.dataDirectory?.trim() || seed.dataDirectory,
      coreStatus: runtimeInfo.core?.status?.trim() || seed.coreStatus,
      coreReachable: Boolean(runtimeInfo.core?.reachable),
      coreHttpStatus: Number(runtimeInfo.core?.httpStatus || 0),
      coreError: runtimeInfo.core?.error?.trim() || "",
      coreManagerEnabled: Boolean(runtimeInfo.coreManager?.enabled),
      coreManagerMode: runtimeInfo.coreManager?.mode || seed.coreManagerMode,
      coreManagerStatus:
        runtimeInfo.coreManager?.status || seed.coreManagerStatus,
      coreManagerPid: Number(runtimeInfo.coreManager?.pid || 0),
      coreManagerCommand:
        runtimeInfo.coreManager?.command?.trim() || seed.coreManagerCommand,
      coreManagerArgs: runtimeInfo.coreManager?.args || seed.coreManagerArgs,
      coreManagerCwd:
        runtimeInfo.coreManager?.cwd?.trim() || seed.coreManagerCwd,
      coreManagerLastError:
        runtimeInfo.coreManager?.lastError?.trim() || seed.coreManagerLastError,
      coreManagerLastExitCode: Number(
        runtimeInfo.coreManager?.lastExitCode || 0,
      ),
      coreManagerLogTail:
        runtimeInfo.coreManager?.logTail || seed.coreManagerLogTail,
      bridgeAvailable: true,
      source: "bridge-runtime",
      issue: "",
      startup: runtimeInfo.startup
        ? {
            phase: runtimeInfo.startup.phase,
            pending: Boolean(runtimeInfo.startup.pending),
            managed: Boolean(runtimeInfo.startup.managed),
            launchMode: runtimeInfo.startup.launchMode?.trim() || seed.launchMode,
            baseUrl: normalizeBaseUrl(runtimeInfo.startup.baseUrl),
            startedAt: runtimeInfo.startup.startedAt?.trim() || "",
            updatedAt: runtimeInfo.startup.updatedAt?.trim() || "",
            issue: runtimeInfo.startup.issue
              ? {
                  code: runtimeInfo.startup.issue.code?.trim() || "STARTUP_ISSUE",
                  severity: runtimeInfo.startup.issue.severity,
                  summary: runtimeInfo.startup.issue.summary?.trim() || "Startup issue",
                  details: Array.isArray(runtimeInfo.startup.issue.details)
                    ? runtimeInfo.startup.issue.details
                    : [],
                  actions: Array.isArray(runtimeInfo.startup.issue.actions)
                    ? runtimeInfo.startup.issue.actions
                    : [],
                }
              : null,
            lastProbe: runtimeInfo.startup.lastProbe
              ? {
                  reachable: Boolean(runtimeInfo.startup.lastProbe.reachable),
                  status: runtimeInfo.startup.lastProbe.status?.trim() || seed.coreStatus,
                  httpStatus: Number(runtimeInfo.startup.lastProbe.httpStatus || 0),
                  health: runtimeInfo.startup.lastProbe.health,
                  error: runtimeInfo.startup.lastProbe.error?.trim() || "",
                }
              : null,
          }
        : undefined,
      core: runtimeInfo.core
        ? {
            reachable: Boolean(runtimeInfo.core.reachable),
            status: runtimeInfo.core.status?.trim() || seed.coreStatus,
            httpStatus: Number(runtimeInfo.core.httpStatus || 0),
            health: runtimeInfo.core.health,
            error: runtimeInfo.core.error?.trim() || "",
          }
        : undefined,
    };
  } catch (error) {
    return {
      ...seed,
      issue:
        error instanceof Error
          ? error.message
          : "failed to read desktop runtime info",
    };
  }
}

export function getDesktopRuntimeRecoveryHint(runtimeInfo: DesktopRuntimeInfo) {
  if (runtimeInfo.issue) {
    return "Renderer fell back to static bridge values. Relaunch the desktop shell or open Diagnostics to verify preload and runtime wiring.";
  }
  if (
    runtimeInfo.coreManagerEnabled &&
    runtimeInfo.coreManagerStatus === "failed"
  ) {
    return "Managed core startup failed. Open Recovery to inspect the launch command, recent log tail, and retry the managed core process.";
  }
  if (
    runtimeInfo.coreManagerEnabled &&
    runtimeInfo.coreManagerStatus === "starting"
  ) {
    return "Managed core is starting in the background. Wait for health to turn ready or open Recovery to inspect launch diagnostics.";
  }
  if (!runtimeInfo.coreReachable) {
    return "Desktop bridge is available, but the local core health probe is not reachable. Verify the core service, port, and launch mode before retrying runtime actions.";
  }
  if (!runtimeInfo.packaged) {
    return "Running in a non-packaged shell. Verify the reported core base URL matches your local service before retrying runtime actions.";
  }
  return "Use this snapshot to confirm the packaged shell is pointed at the intended core instance before triaging runtime failures.";
}

export function getDesktopRuntimeDiagnosis(
  runtimeInfo: DesktopRuntimeInfo,
): DesktopRuntimeDiagnosis {
  const issues: DesktopRecoveryIssue[] = [];
  const rawDiagnostics = extractRuntimeDiagnostics(runtimeInfo);

  for (const item of rawDiagnostics) {
    const issue = normalizeExternalDiagnostic(item, runtimeInfo);
    if (issue) {
      issues.push(issue);
    }
  }

  if (runtimeInfo.issue) {
    issues.push({
      code: "DESKTOP_BRIDGE_UNAVAILABLE",
      category: "runtime_attention",
      severity: "error",
      summary: "Renderer could not load full desktop runtime context",
      detail:
        runtimeInfo.issue ||
        "The renderer fell back to static bridge values because runtime info could not be loaded.",
      source: "bridge",
      mode: inferRuntimeMode(runtimeInfo),
      evidence: compactEvidence([
        `source=${runtimeInfo.source}`,
        `bridgeAvailable=${String(runtimeInfo.bridgeAvailable)}`,
        runtimeInfo.issue,
      ]),
      actions: buildActions([
        "retry-health-probe",
        "open-diagnostics",
        "open-settings",
      ]),
    });
  }

  if (!runtimeInfo.coreReachable) {
    issues.push(
      runtimeInfo.coreManagerEnabled
        ? {
            code:
              runtimeInfo.coreManagerStatus === "starting"
                ? "MANAGED_CORE_STARTING"
                : runtimeInfo.coreManagerStatus === "failed"
                  ? "MANAGED_CORE_FAILED"
                  : runtimeInfo.coreManagerStatus === "exited"
                    ? "MANAGED_CORE_EXITED"
                    : "MANAGED_CORE_UNREACHABLE",
            category: "core_unavailable",
            severity:
              runtimeInfo.coreManagerStatus === "starting"
                ? "warning"
                : runtimeInfo.coreManagerStatus === "failed" ||
                    runtimeInfo.coreManagerStatus === "exited"
                  ? "critical"
                  : "error",
            summary:
              runtimeInfo.coreManagerStatus === "starting"
                ? "Managed core is still starting"
                : runtimeInfo.coreManagerStatus === "failed"
                  ? "Managed core failed to start"
                  : runtimeInfo.coreManagerStatus === "exited"
                    ? "Managed core exited before the health probe recovered"
                    : "Managed core is enabled but the local health probe is still unavailable",
            detail:
              runtimeInfo.coreManagerLastError ||
              runtimeInfo.coreError ||
              "The desktop shell is configured to manage the core process, but the health probe has not recovered yet.",
            source: "core-manager",
            mode: "managed",
            evidence: compactEvidence([
              `coreStatus=${runtimeInfo.coreStatus}`,
              `managerStatus=${runtimeInfo.coreManagerStatus}`,
              runtimeInfo.coreManagerCommand
                ? `command=${runtimeInfo.coreManagerCommand}`
                : "",
              runtimeInfo.coreManagerCwd
                ? `cwd=${runtimeInfo.coreManagerCwd}`
                : "",
              runtimeInfo.coreManagerLastExitCode
                ? `exitCode=${String(runtimeInfo.coreManagerLastExitCode)}`
                : "",
              runtimeInfo.coreError,
              runtimeInfo.coreManagerLastError,
            ]),
            actions: buildActions([
              runtimeInfo.coreManagerStatus === "starting"
                ? "retry-health-probe"
                : "restart-managed-core",
              "start-managed-core",
              "open-diagnostics",
              "open-data-folder",
              runtimeInfo.launchMode === "safe-mode"
                ? "relaunch-normal-mode"
                : "relaunch-safe-mode",
            ]),
          }
        : {
            code: "EXTERNAL_CORE_UNREACHABLE",
            category: "core_unavailable",
            severity: "error",
            summary: "External core health probe is unreachable",
            detail:
              runtimeInfo.coreError ||
              "The renderer cannot reach the configured core base URL. This shell appears to depend on an externally managed core.",
            source: "health-probe",
            mode: "external",
            evidence: compactEvidence([
              `coreStatus=${runtimeInfo.coreStatus}`,
              `baseUrl=${runtimeInfo.coreBaseUrl}`,
              runtimeInfo.coreHttpStatus
                ? `httpStatus=${String(runtimeInfo.coreHttpStatus)}`
                : "",
              runtimeInfo.coreError,
            ]),
            actions: buildActions([
              "retry-health-probe",
              "open-settings",
              "relaunch-safe-mode",
              "open-diagnostics",
            ]),
          },
    );
  }

  if (runtimeInfo.coreReachable && issues.length === 0) {
    return {
      status: "healthy",
      summary:
        "Desktop runtime and local core probe look healthy with the currently available renderer fields.",
      mode: inferRuntimeMode(runtimeInfo),
      managerState: runtimeInfo.coreManagerStatus,
      issues: [],
    };
  }

  const deduped = dedupeIssues(issues);
  return {
    status: deduped.some((item) => item.severity === "critical")
      ? "failed"
      : "degraded",
    summary: buildDiagnosisSummary(runtimeInfo, deduped),
    mode: inferRuntimeMode(runtimeInfo),
    managerState: runtimeInfo.coreManagerStatus,
    issues: deduped,
  };
}

export async function selectDesktopDirectory(): Promise<DesktopSelectDirectoryResult> {
  const picker =
    typeof window === "undefined"
      ? undefined
      : window.desktopBridge?.selectDirectory;
  if (!picker) {
    return {
      canceled: true,
      path: "",
    };
  }
  return picker();
}

export async function openDesktopPath(
  targetPath: string,
): Promise<DesktopShellResult> {
  const opener =
    typeof window === "undefined" ? undefined : window.desktopBridge?.openPath;
  if (!opener) {
    return {
      ok: false,
      error: "desktop bridge unavailable in renderer",
    };
  }
  return opener(targetPath);
}

export async function showDesktopItemInFolder(
  targetPath: string,
): Promise<DesktopShellResult> {
  const opener =
    typeof window === "undefined"
      ? undefined
      : window.desktopBridge?.showItemInFolder;
  if (!opener) {
    return {
      ok: false,
      error: "desktop bridge unavailable in renderer",
    };
  }
  return opener(targetPath);
}

export async function relaunchDesktopSafeMode(): Promise<DesktopShellResult> {
  const relaunch =
    typeof window === "undefined"
      ? undefined
      : window.desktopBridge?.relaunchSafeMode;
  if (!relaunch) {
    return {
      ok: false,
      error: "desktop bridge unavailable in renderer",
    };
  }
  return relaunch();
}

export async function relaunchDesktopNormalMode(): Promise<DesktopShellResult> {
  const relaunch =
    typeof window === "undefined"
      ? undefined
      : window.desktopBridge?.relaunchNormalMode;
  if (!relaunch) {
    return {
      ok: false,
      error: "desktop bridge unavailable in renderer",
    };
  }
  return relaunch();
}

export async function startManagedCore(): Promise<DesktopShellResult> {
  const starter =
    typeof window === "undefined"
      ? undefined
      : window.desktopBridge?.startManagedCore;
  if (!starter) {
    return {
      ok: false,
      error: "desktop bridge unavailable in renderer",
    };
  }
  return starter();
}

export async function restartManagedCore(): Promise<DesktopShellResult> {
  const starter =
    typeof window === "undefined"
      ? undefined
      : window.desktopBridge?.restartManagedCore;
  if (!starter) {
    return {
      ok: false,
      error: "desktop bridge unavailable in renderer",
    };
  }
  return starter();
}

export async function exportDesktopDiagnostics(
  payload: Record<string, unknown>,
  suggestedName: string,
): Promise<DesktopShellResult> {
  const exporter =
    typeof window === "undefined"
      ? undefined
      : window.desktopBridge?.saveDiagnosticExport;
  if (!exporter) {
    return {
      ok: false,
      error: "desktop bridge unavailable in renderer",
    };
  }

  try {
    return await exporter({
      suggestedName,
      content: JSON.stringify(payload, null, 2),
    });
  } catch (error) {
    return {
      ok: false,
      error:
        error instanceof Error
          ? error.message
          : "failed to export diagnostics",
    };
  }
}

function inferRuntimeMode(
  runtimeInfo: DesktopRuntimeInfo,
): "managed" | "external" | "unknown" {
  if (runtimeInfo.coreManagerEnabled) {
    return "managed";
  }
  if (runtimeInfo.coreBaseUrl || runtimeInfo.coreStatus !== "unknown") {
    return "external";
  }
  return "unknown";
}

function readStartupHealthDiagnostics(runtimeInfo: DesktopRuntimeInfo) {
  const health = runtimeInfo.core?.health;
  if (!health || typeof health !== "object") {
    return [];
  }
  const startup =
    "startup" in health ? (health.startup as Record<string, unknown>) : undefined;
  if (!startup || typeof startup !== "object") {
    return [];
  }
  const diagnostics = startup.diagnostics;
  if (!Array.isArray(diagnostics)) {
    return [];
  }

  return diagnostics
    .filter((item) => item && typeof item === "object")
    .map((item) => {
      const record = item as Record<string, unknown>;
      const detail = readString(record.detail);
      return {
        code: readString(record.code, record.error_code),
        summary: readString(record.summary, record.message),
        detail,
        severity: readString(record.severity),
        source: "startup-diagnostic",
        mode: runtimeInfo.coreManagerEnabled ? "managed" : "external",
        evidence: compactEvidence([
          readString(record.component),
          readString(record.field),
          detail,
        ]),
        actions: Array.isArray(record.actions)
          ? record.actions.map((action) => {
              const text =
                typeof action === "string" ? action : stringifyValue(action);
              return {
                id: mapRuntimeActionID(text),
                label: text,
                description: text,
              };
            })
          : [],
      };
    });
}

function classifyRecoveryCategory(
  code: string,
  detail: string,
  summary: string,
): DesktopRecoveryIssue["category"] {
  const normalized = `${code} ${detail} ${summary}`.toLowerCase();
  if (normalized.includes("migration")) {
    return "migration_failure";
  }
  if (
    normalized.includes("data root")
    || normalized.includes("db_path")
    || normalized.includes("directory")
    || normalized.includes("writable")
    || normalized.includes("permission denied")
  ) {
    return "data_directory_unwritable";
  }
  if (
    normalized.includes("policy_denied")
    || normalized.includes("permission_denied")
    || normalized.includes("forbidden")
    || normalized.includes("run_denied")
  ) {
    return "policy_denied";
  }
  if (
    normalized.includes("verification_conflict")
    || normalized.includes("missing_evidence")
    || normalized.includes("failed_checks")
    || normalized.includes("verification contract")
  ) {
    return "verification_conflict";
  }
  if (normalized.includes("fault_loop") || normalized.includes("fault loop")) {
    return "fault_loop_detected";
  }
  if (
    normalized.includes("runtime")
    || normalized.includes("core")
    || normalized.includes("econnrefused")
    || normalized.includes("connection refused")
    || normalized.includes("unreachable")
  ) {
    return "core_unavailable";
  }
  return "runtime_attention";
}

function mapRuntimeActionID(value: string) {
  const normalized = value.trim().toLowerCase();
  if (!normalized) {
    return "";
  }
  if (
    normalized.includes("restart managed core")
    || normalized.includes("restart the managed core")
  ) {
    return "restart-managed-core";
  }
  if (
    normalized.includes("start managed core")
    || normalized.includes("configure managed core startup")
  ) {
    return "start-managed-core";
  }
  if (normalized.includes("safe mode")) {
    return "relaunch-safe-mode";
  }
  if (normalized.includes("settings") || normalized.includes("base url")) {
    return "open-settings";
  }
  if (
    normalized.includes("data folder")
    || normalized.includes("data directory")
    || normalized.includes("permissions")
  ) {
    return "open-data-folder";
  }
  if (normalized.includes("diagnostics") || normalized.includes("inspect")) {
    return "open-diagnostics";
  }
  if (normalized.includes("retry") || normalized.includes("wait a moment")) {
    return "retry-health-probe";
  }
  return normalized.replace(/[^a-z0-9]+/g, "-").replace(/^-+|-+$/g, "");
}

function extractRuntimeDiagnostics(runtimeInfo: DesktopRuntimeInfo) {
  const diagnostics: Array<Record<string, unknown>> = [];

  if (runtimeInfo.startup?.issue) {
    diagnostics.push({
      code: runtimeInfo.startup.issue.code,
      summary: runtimeInfo.startup.issue.summary,
      detail: runtimeInfo.startup.issue.details.join(" "),
      severity: runtimeInfo.startup.issue.severity,
      source: "startup-diagnostic",
      mode: runtimeInfo.startup.managed ? "managed" : "external",
      evidence: runtimeInfo.startup.issue.details,
      actions: runtimeInfo.startup.issue.actions.map((action) => ({
        id: mapRuntimeActionID(action),
        label: action,
        description: action,
      })),
    });
  }

  return diagnostics.concat(readStartupHealthDiagnostics(runtimeInfo));
}

function normalizeExternalDiagnostic(
  value: unknown,
  runtimeInfo: DesktopRuntimeInfo,
): DesktopRecoveryIssue | null {
  if (!value || typeof value !== "object") {
    return null;
  }
  const record = value as Record<string, unknown>;
  const code = readString(record.code, record.issueCode, record.error_code);
  const summary = readString(record.summary, record.title, record.message);
  if (!code && !summary) {
    return null;
  }

  const detail = readString(
    record.detail,
    record.description,
    record.error,
    record.hint,
  );
  const source = normalizeSource(readString(record.source, record.scope));
  const severity = normalizeSeverity(
    readString(record.severity, record.level, record.status),
  );
  const mode = normalizeMode(
    readString(record.mode, record.runtimeMode, record.managerMode),
    runtimeInfo,
  );
  const evidence = normalizeEvidence(record);
  const actions = normalizeActions(record.actions);

  return {
    code: code || "STARTUP_DIAGNOSTIC",
    category: classifyRecoveryCategory(code, detail, summary),
    severity,
    summary: summary || "Startup diagnostic reported by runtime",
    detail:
      detail ||
      "A structured startup diagnostic was reported, but it did not include detail text.",
    source,
    mode,
    evidence,
    actions:
      actions.length > 0
        ? actions
        : buildActions([
            "retry-health-probe",
            mode === "managed" ? "restart-managed-core" : "open-settings",
            "open-diagnostics",
          ]),
  };
}

function normalizeSource(
  value: string,
): "bridge" | "health-probe" | "core-manager" | "startup-diagnostic" {
  const normalized = value.trim().toLowerCase();
  if (normalized.includes("bridge")) {
    return "bridge";
  }
  if (normalized.includes("manager")) {
    return "core-manager";
  }
  if (normalized.includes("health")) {
    return "health-probe";
  }
  return "startup-diagnostic";
}

function normalizeSeverity(value: string): DesktopRecoverySeverity {
  const normalized = value.trim().toLowerCase();
  if (normalized === "critical" || normalized === "fatal") {
    return "critical";
  }
  if (normalized === "error" || normalized === "failed") {
    return "error";
  }
  if (normalized === "warning" || normalized === "warn") {
    return "warning";
  }
  return "info";
}

function normalizeMode(
  value: string,
  runtimeInfo: DesktopRuntimeInfo,
): "managed" | "external" | "unknown" {
  const normalized = value.trim().toLowerCase();
  if (normalized === "managed") {
    return "managed";
  }
  if (normalized === "external") {
    return "external";
  }
  return inferRuntimeMode(runtimeInfo);
}

function normalizeEvidence(record: Record<string, unknown>) {
  const raw = record.evidence;
  if (Array.isArray(raw)) {
    return compactEvidence(
      raw.map((item) =>
        typeof item === "string" ? item : stringifyValue(item),
      ),
    );
  }
  return compactEvidence([
    readString(record.detail),
    readString(record.hint),
    readString(record.path),
    readString(record.command),
  ]);
}

function normalizeActions(value: unknown): DesktopRecoveryAction[] {
  if (!Array.isArray(value)) {
    return [];
  }
  const actions: DesktopRecoveryAction[] = [];
  for (const item of value) {
    if (!item || typeof item !== "object") {
      continue;
    }
    const record = item as Record<string, unknown>;
    const id = mapRuntimeActionID(
      readString(record.id, record.code, record.action),
    );
    const label = readString(record.label, record.title);
    const description = readString(
      record.description,
      record.detail,
      record.hint,
    );
    if (!id && !label) {
      continue;
    }
    actions.push({
      id: id || label.toLowerCase().replace(/\s+/g, "-"),
      label: label || id || "Open recovery action",
      description:
        description || "A runtime-provided recovery action is available.",
    });
  }
  return actions;
}

function buildDiagnosisSummary(
  runtimeInfo: DesktopRuntimeInfo,
  issues: DesktopRecoveryIssue[],
) {
  if (issues.length === 0) {
    return getDesktopRuntimeRecoveryHint(runtimeInfo);
  }
  const primary = issues[0];
  const modeLabel =
    primary.mode === "managed" ? "managed core" : "external core";
  return `${primary.summary} ${issues.length > 1 ? `(${issues.length} issues detected across ${modeLabel} startup diagnostics.)` : `(${modeLabel}).`}`;
}

function dedupeIssues(issues: DesktopRecoveryIssue[]) {
  const seen = new Set<string>();
  return issues.filter((item) => {
    const key = `${item.code}:${item.mode}:${item.summary}`;
    if (seen.has(key)) {
      return false;
    }
    seen.add(key);
    return true;
  });
}

function buildActions(actionIds: Array<string | false>) {
  const catalog: Record<string, DesktopRecoveryAction> = {
    "retry-health-probe": {
      id: "retry-health-probe",
      label: "Retry health probe",
      description:
        "Refresh renderer diagnostics and recheck the local core probe.",
    },
    "start-managed-core": {
      id: "start-managed-core",
      label: "Start managed core",
      description: "Launch the managed core command from the desktop shell.",
    },
    "restart-managed-core": {
      id: "restart-managed-core",
      label: "Restart managed core",
      description: "Restart the managed core process and inspect fresh logs.",
    },
    "relaunch-safe-mode": {
      id: "relaunch-safe-mode",
      label: "Relaunch safe mode",
      description:
        "Restart the shell with worker-heavy startup paths disabled.",
    },
    "relaunch-normal-mode": {
      id: "relaunch-normal-mode",
      label: "Return to normal mode",
      description: "Restart the shell with the standard launch path.",
    },
    "open-settings": {
      id: "open-settings",
      label: "Open settings",
      description: "Review project id, base URL, and local path configuration.",
    },
    "open-diagnostics": {
      id: "open-diagnostics",
      label: "Open diagnostics",
      description:
        "Compare renderer, runtime, and persisted diagnostic snapshots.",
    },
    "open-data-folder": {
      id: "open-data-folder",
      label: "Open data folder",
      description: "Inspect the current desktop data directory on disk.",
    },
  };

  return actionIds
    .filter((item): item is string => Boolean(item))
    .map((item) => catalog[item])
    .filter(Boolean);
}

function compactEvidence(items: string[]) {
  return items.map((item) => item.trim()).filter(Boolean);
}

function readString(...values: unknown[]) {
  for (const value of values) {
    if (typeof value === "string" && value.trim() !== "") {
      return value.trim();
    }
  }
  return "";
}

function stringifyValue(value: unknown) {
  if (typeof value === "string") {
    return value;
  }
  try {
    return JSON.stringify(value);
  } catch {
    return String(value);
  }
}
