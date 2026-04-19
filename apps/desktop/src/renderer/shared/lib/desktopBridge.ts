import type { DesktopRuntimeInfo } from "@/shared/lib/preferences";

type DesktopBridgeShape = NonNullable<Window["desktopBridge"]> & {
  dataDirectory?: string;
  userDataPath?: string;
};

export type DesktopActionResult = {
  ok: boolean;
  message: string;
  path?: string;
  canceled?: boolean;
  unsupported?: boolean;
};

function getDesktopBridge(): DesktopBridgeShape | undefined {
  if (typeof window === "undefined") {
    return undefined;
  }
  return window.desktopBridge as DesktopBridgeShape | undefined;
}

export function getDesktopBridgeCapabilities() {
  const bridge = getDesktopBridge();
  return {
    bridgeAvailable: Boolean(bridge),
    canSelectDirectory: typeof bridge?.selectDirectory === "function",
    canOpenPath: typeof bridge?.openPath === "function",
    canShowItemInFolder: typeof bridge?.showItemInFolder === "function",
  };
}

export function resolveDesktopDataDirectory(
  runtimeInfo?: Partial<DesktopRuntimeInfo>,
) {
  const bridge = getDesktopBridge();
  return (
    runtimeInfo?.dataDirectory?.trim() ||
    bridge?.dataDirectory?.trim() ||
    bridge?.userDataPath?.trim() ||
    ""
  );
}

export async function selectDirectory(
  label: string,
): Promise<DesktopActionResult> {
  const bridge = getDesktopBridge();
  if (typeof bridge?.selectDirectory !== "function") {
    return {
      ok: false,
      unsupported: true,
      message: `${label} selection is unavailable because the desktop bridge method is not available in this shell.`,
    };
  }

  try {
    const result = await bridge.selectDirectory();
    if (result.canceled || !result.path.trim()) {
      return {
        ok: false,
        canceled: true,
        message: `${label} selection was canceled.`,
      };
    }
    return {
      ok: true,
      path: result.path.trim(),
      message: `${label} set to ${result.path.trim()}.`,
    };
  } catch (error) {
    return {
      ok: false,
      message:
        error instanceof Error
          ? error.message
          : `${label} selection failed in the desktop shell.`,
    };
  }
}

export async function openDesktopPath(
  targetPath: string,
  label: string,
): Promise<DesktopActionResult> {
  const normalized = targetPath.trim();
  const bridge = getDesktopBridge();

  if (!normalized) {
    return {
      ok: false,
      message: `${label} is not available yet because no target path is known in the renderer.`,
    };
  }

  if (typeof bridge?.openPath !== "function") {
    return {
      ok: false,
      unsupported: true,
      path: normalized,
      message: `${label} could not be opened because the desktop bridge openPath method is not available in this shell.`,
    };
  }

  try {
    const result = await bridge.openPath(normalized);
    return {
      ok: Boolean(result.ok),
      path: normalized,
      message: result.ok
        ? `${label} opened in the desktop shell.`
        : result.error?.trim() || `${label} failed to open.`,
    };
  } catch (error) {
    return {
      ok: false,
      path: normalized,
      message:
        error instanceof Error
          ? error.message
          : `${label} failed to open in the desktop shell.`,
    };
  }
}

export async function showDesktopItemInFolder(
  targetPath: string,
  label: string,
): Promise<DesktopActionResult> {
  const normalized = targetPath.trim();
  const bridge = getDesktopBridge();

  if (!normalized) {
    return {
      ok: false,
      message: `${label} is empty, so there is nothing to reveal in the desktop shell.`,
    };
  }

  if (typeof bridge?.showItemInFolder !== "function") {
    return {
      ok: false,
      unsupported: true,
      path: normalized,
      message: `${label} could not be revealed because the desktop bridge showItemInFolder method is not available in this shell.`,
    };
  }

  try {
    const result = await bridge.showItemInFolder(normalized);
    return {
      ok: Boolean(result.ok),
      path: normalized,
      message: result.ok
        ? `${label} revealed in the desktop shell.`
        : result.error?.trim() || `${label} could not be revealed.`,
    };
  } catch (error) {
    return {
      ok: false,
      path: normalized,
      message:
        error instanceof Error
          ? error.message
          : `${label} could not be revealed in the desktop shell.`,
    };
  }
}

export function openCoreUrl(targetUrl: string): DesktopActionResult {
  const normalized = targetUrl.trim();
  if (!normalized) {
    return {
      ok: false,
      message: "Core URL is empty, so there is nothing to open.",
    };
  }

  try {
    window.open(normalized, "_blank", "noopener,noreferrer");
    return {
      ok: true,
      path: normalized,
      message: `Opened ${normalized} in a browser context.`,
    };
  } catch (error) {
    return {
      ok: false,
      path: normalized,
      message:
        error instanceof Error
          ? error.message
          : "Failed to open the current core URL.",
    };
  }
}
