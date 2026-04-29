import { getStoredCoreBaseUrl } from "./preferences";

type ApiEnvelope<T> = {
  code?: number;
  message?: string;
  data?: T;
};

export function getCoreBaseUrl() {
  return getStoredCoreBaseUrl();
}

export async function apiGet<T>(path: string): Promise<T> {
  return apiRequest<T>(path, { method: "GET" });
}

export async function apiPost<T>(path: string, body?: unknown): Promise<T> {
  return apiRequest<T>(path, {
    method: "POST",
    body: body === undefined ? undefined : JSON.stringify(body),
  });
}

export async function apiPut<T>(path: string, body?: unknown): Promise<T> {
  return apiRequest<T>(path, {
    method: "PUT",
    body: body === undefined ? undefined : JSON.stringify(body),
  });
}

export async function apiDelete<T>(path: string): Promise<T> {
  return apiRequest<T>(path, { method: "DELETE" });
}

async function apiRequest<T>(
  path: string,
  options: {
    method: "GET" | "POST" | "PUT" | "DELETE";
    body?: string;
  },
): Promise<T> {
  const response = await fetch(`${getCoreBaseUrl()}${path}`, {
    method: options.method,
    headers: {
      Accept: "application/json",
      ...(options.body !== undefined ? { "Content-Type": "application/json" } : {}),
    },
    body: options.body,
  });

  const text = await response.text();
  let payload: unknown = null;
  if (text.trim() !== "") {
    payload = JSON.parse(text);
  }

  if (!response.ok) {
    const message = extractErrorMessage(payload, response.statusText);
    throw new Error(message);
  }

  // GoFrame returns business errors as HTTP 200 with code != 0 and data == null
  if (payload && typeof payload === "object" && "code" in payload) {
    const code = (payload as ApiEnvelope<T>).code;
    if (code !== undefined && code !== 0) {
      const message = extractErrorMessage(payload, response.statusText);
      throw new Error(message);
    }
  }

  return unwrapPayload<T>(payload);
}

function unwrapPayload<T>(payload: unknown): T {
  if (payload && typeof payload === "object" && "data" in payload) {
    return (payload as ApiEnvelope<T>).data as T;
  }
  return payload as T;
}

function extractErrorMessage(payload: unknown, fallback: string) {
  if (payload && typeof payload === "object") {
    if ("message" in payload && typeof payload.message === "string" && payload.message.trim() !== "") {
      return payload.message;
    }
    if ("msg" in payload && typeof (payload as { msg?: string }).msg === "string") {
      return (payload as { msg: string }).msg;
    }
  }
  return fallback || "Request failed";
}
