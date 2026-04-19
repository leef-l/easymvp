import { useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { apiGet, apiPost } from "@/shared/lib/api";
import { useProjectState } from "@/shared/lib/project";
import { useQuery } from "@/shared/lib/query";
import { getStoredCoreBaseUrl, setStoredCoreBaseUrl, setStoredProjectId } from "@/shared/lib/preferences";
import type { CreateProjectPayload, CreateProjectResponse, RuntimeHealthView, SystemHealthView } from "@/shared/lib/types";
import { QueryPanel } from "@/shared/ui/QueryPanel";

export function SettingsPage() {
  const navigate = useNavigate();
  const { projectId, updateProjectId, buildRoute, routes } = useProjectState();
  const [draftProjectId, setDraftProjectId] = useState(projectId);
  const [draftBaseUrl, setDraftBaseUrl] = useState(getStoredCoreBaseUrl());
  const [saveTick, setSaveTick] = useState(0);
  const [createBusy, setCreateBusy] = useState(false);
  const [createError, setCreateError] = useState("");
  const [createMessage, setCreateMessage] = useState("");
  const [createPayload, setCreatePayload] = useState<CreateProjectPayload>({
    name: "",
    project_category: "web_app",
    goal_summary: "",
    workspace_root: "",
    repo_root: "",
  });

  const systemState = useQuery(() => apiGet<SystemHealthView>("/api/v3/system/healthz"), [saveTick]);
  const runtimeState = useQuery(() => apiGet<RuntimeHealthView>("/api/v3/runtime/healthz"), [saveTick]);

  const saveSummary = useMemo(() => {
    return `Stored project ${projectId} and core ${getStoredCoreBaseUrl()}`;
  }, [projectId, saveTick]);

  function handleSave() {
    setStoredCoreBaseUrl(draftBaseUrl);
    setStoredProjectId(draftProjectId);
    updateProjectId(draftProjectId);
    setSaveTick((value) => value + 1);
  }

  async function handleCreateProject() {
    setCreateBusy(true);
    setCreateError("");
    setCreateMessage("");
    try {
      const result = await apiPost<CreateProjectResponse>("/api/v3/projects", createPayload);
      const nextProjectId = result.resource_id.trim();
      if (nextProjectId) {
        setDraftProjectId(nextProjectId);
        updateProjectId(nextProjectId);
        setStoredProjectId(nextProjectId);
        navigate(buildRoute("/workspace", { project: nextProjectId }));
      }
      setCreateMessage(`${result.next_action} · ${nextProjectId || "project created"}`);
      setSaveTick((value) => value + 1);
    } catch (error) {
      setCreateError(error instanceof Error ? error.message : "Create project failed");
    } finally {
      setCreateBusy(false);
    }
  }

  return (
    <QueryPanel
      loading={systemState.loading || runtimeState.loading}
      error={systemState.error || runtimeState.error}
      title="Settings"
      onRetry={() => setSaveTick((value) => value + 1)}
      secondaryActionLabel="Open Diagnostics"
      onSecondaryAction={() => navigate(routes.diagnostics)}
      recoveryMessage="If health checks fail, verify the core base URL and use diagnostics to inspect runtime connectivity."
    >
      {systemState.data && runtimeState.data ? (
        <section className="dashboard-page">
          <div className="dashboard-intro">
            <div>
              <p className="placeholder-section">Settings</p>
              <h3 className="placeholder-title">Local workbench preferences</h3>
              <p className="placeholder-description">{saveSummary}</p>
            </div>
            <button className="primary-button" onClick={handleSave}>
              Save and refresh
            </button>
          </div>

          <div className="content-grid">
            <section className="data-panel">
              <div className="panel-header">
                <h3>Workbench Defaults</h3>
              </div>
              <div className="form-grid">
                <label className="form-field">
                  <span>Project ID</span>
                  <input className="project-input" value={draftProjectId} onChange={(event) => setDraftProjectId(event.target.value)} />
                </label>
                <label className="form-field">
                  <span>Core Base URL</span>
                  <input className="project-input" value={draftBaseUrl} onChange={(event) => setDraftBaseUrl(event.target.value)} />
                </label>
              </div>
            </section>

            <section className="data-panel">
              <div className="panel-header">
                <h3>Create Project</h3>
              </div>
              <div className="form-grid">
                <label className="form-field">
                  <span>Name</span>
                  <input
                    className="project-input"
                    value={createPayload.name}
                    onChange={(event) => setCreatePayload((current) => ({ ...current, name: event.target.value }))}
                  />
                </label>
                <label className="form-field">
                  <span>Project Category</span>
                  <input
                    className="project-input"
                    value={createPayload.project_category}
                    onChange={(event) => setCreatePayload((current) => ({ ...current, project_category: event.target.value }))}
                  />
                </label>
                <label className="form-field">
                  <span>Goal Summary</span>
                  <textarea
                    className="project-input project-textarea"
                    value={createPayload.goal_summary}
                    onChange={(event) => setCreatePayload((current) => ({ ...current, goal_summary: event.target.value }))}
                  />
                </label>
                <label className="form-field">
                  <span>Workspace Root</span>
                  <input
                    className="project-input"
                    value={createPayload.workspace_root}
                    onChange={(event) => setCreatePayload((current) => ({ ...current, workspace_root: event.target.value }))}
                  />
                </label>
                <label className="form-field">
                  <span>Repo Root</span>
                  <input
                    className="project-input"
                    value={createPayload.repo_root}
                    onChange={(event) => setCreatePayload((current) => ({ ...current, repo_root: event.target.value }))}
                  />
                </label>
                <div className="action-row">
                  <button className="primary-button" disabled={createBusy} onClick={() => void handleCreateProject()}>
                    {createBusy ? "Creating..." : "Create Project"}
                  </button>
                </div>
                {createError ? <p className="error-copy">{createError}</p> : null}
                {createMessage ? <p className="muted-copy">{createMessage}</p> : null}
              </div>
            </section>

            <section className="data-panel">
              <div className="panel-header">
                <h3>Health Checks</h3>
              </div>
              <div className="stack-list">
                <article className="list-card">
                  <div className="list-card-head">
                    <strong>System</strong>
                    <span className="status-pill">{systemState.data.status}</span>
                  </div>
                  <p>{systemState.data.service} · {systemState.data.version} · {systemState.data.timestamp}</p>
                </article>
                <article className="list-card">
                  <div className="list-card-head">
                    <strong>Runtime</strong>
                    <span className="status-pill">{runtimeState.data.status}</span>
                  </div>
                  <p>{runtimeState.data.base_url}</p>
                </article>
              </div>
            </section>
          </div>
        </section>
      ) : null}
    </QueryPanel>
  );
}
