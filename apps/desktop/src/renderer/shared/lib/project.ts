import { useMemo } from "react";
import { useSearchParams } from "react-router-dom";
import { getStoredProjectId, setStoredProjectId } from "./preferences";

type RouteQueryValue = string | number | boolean | null | undefined;

export function useProjectState() {
  const [searchParams, setSearchParams] = useSearchParams();
  const projectId = searchParams.get("project")?.trim() || getStoredProjectId();

  function buildRoute(
    pathname: string,
    extraParams?: Record<string, RouteQueryValue>,
  ) {
    const next = new URLSearchParams();
    next.set("project", projectId);
    for (const [key, value] of Object.entries(extraParams ?? {})) {
      if (value === undefined || value === null || `${value}`.trim() === "") {
        continue;
      }
      next.set(key, String(value));
    }
    return `${pathname}?${next.toString()}`;
  }

  const routes = useMemo(
    () => ({
      workspace: buildRoute("/workspace"),
      plan: buildRoute("/plan"),
      execution: buildRoute("/execution"),
      replay: buildRoute("/replay"),
      acceptance: buildRoute("/acceptance"),
      audit: buildRoute("/audit"),
      diagnostics: buildRoute("/diagnostics"),
      settings: buildRoute("/settings"),
      repairDraft: buildRoute("/repair-draft"),
    }),
    [projectId],
  );

  function updateProjectId(nextProjectId: string) {
    const normalized = nextProjectId.trim() || getStoredProjectId();
    setStoredProjectId(normalized);
    setSearchParams((current) => {
      const next = new URLSearchParams(current);
      next.set("project", normalized);
      return next;
    });
  }

  return {
    searchParams,
    buildRoute,
    projectId,
    updateProjectId,
    routes,
  };
}
