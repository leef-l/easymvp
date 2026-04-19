const PROJECT_ID_KEY = "easymvp.desktop.projectId";
const CORE_BASE_URL_KEY = "easymvp.desktop.coreBaseUrl";

const DEFAULT_PROJECT_ID = "project-demo";
const DEFAULT_CORE_BASE_URL = "http://127.0.0.1:8000";

export function getStoredProjectId() {
  if (typeof window === "undefined") {
    return DEFAULT_PROJECT_ID;
  }
  return window.localStorage.getItem(PROJECT_ID_KEY)?.trim() || DEFAULT_PROJECT_ID;
}

export function setStoredProjectId(projectId: string) {
  if (typeof window === "undefined") {
    return;
  }
  const normalized = projectId.trim() || DEFAULT_PROJECT_ID;
  window.localStorage.setItem(PROJECT_ID_KEY, normalized);
}

export function getStoredCoreBaseUrl() {
  if (typeof window === "undefined") {
    return DEFAULT_CORE_BASE_URL;
  }
  const fromStorage = window.localStorage.getItem(CORE_BASE_URL_KEY)?.trim();
  if (fromStorage) {
    return fromStorage.replace(/\/+$/, "");
  }
  const fromBridge = window.desktopBridge?.coreBaseUrl?.trim();
  if (fromBridge) {
    return fromBridge.replace(/\/+$/, "");
  }
  return DEFAULT_CORE_BASE_URL;
}

export function setStoredCoreBaseUrl(baseUrl: string) {
  if (typeof window === "undefined") {
    return;
  }
  const normalized = baseUrl.trim() || DEFAULT_CORE_BASE_URL;
  window.localStorage.setItem(CORE_BASE_URL_KEY, normalized.replace(/\/+$/, ""));
}
