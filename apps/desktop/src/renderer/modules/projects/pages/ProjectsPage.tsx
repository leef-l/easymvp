import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import { apiDelete, apiGet, apiPost, apiPut } from "@/shared/lib/api";
import { selectDesktopDirectory, setStoredProjectId } from "@/shared/lib/preferences";
import { useProjectState } from "@/shared/lib/project";
import { useQuery } from "@/shared/lib/query";
import type { CreateProjectPayload, CreateProjectResponse, UpdateProjectPayload, WorkspaceHomeView } from "@/shared/lib/types";
import { QueryPanel } from "@/shared/ui/QueryPanel";

export function ProjectsPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { buildRoute } = useProjectState();
  const [refreshTick, setRefreshTick] = useState(0);
  const [editingProject, setEditingProject] = useState<{
    project_id: string;
    name: string;
    goal_summary?: string;
  } | null>(null);
  const [editPayload, setEditPayload] = useState<UpdateProjectPayload>({});
  const [editBusy, setEditBusy] = useState(false);
  const [editError, setEditError] = useState("");
  const [deletingProject, setDeletingProject] = useState<{
    project_id: string;
    name: string;
  } | null>(null);
  const [deleteBusy, setDeleteBusy] = useState(false);
  const [deleteError, setDeleteError] = useState("");
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [createPayload, setCreatePayload] = useState<CreateProjectPayload>({
    name: "",
    project_category: "software_dev",
    goal_summary: "",
    workspace_root: "",
    repo_root: "",
  });
  const [createBusy, setCreateBusy] = useState(false);
  const [createError, setCreateError] = useState("");

  const state = useQuery<WorkspaceHomeView>(
    () => apiGet<WorkspaceHomeView>("/api/v3/workspace/home-view"),
    [refreshTick],
  );

  function selectProject(projectId: string) {
    setStoredProjectId(projectId);
    navigate(`/workspace?project=${encodeURIComponent(projectId)}`);
  }

  function openEdit(project: { project_id: string; name: string }) {
    setEditingProject(project);
    setEditPayload({ name: project.name });
    setEditError("");
  }

  function closeEdit() {
    setEditingProject(null);
    setEditPayload({});
    setEditBusy(false);
    setEditError("");
  }

  async function handleSaveEdit() {
    if (!editingProject) return;
    setEditBusy(true);
    setEditError("");
    try {
      await apiPut(`/api/v3/projects/${encodeURIComponent(editingProject.project_id)}`, editPayload);
      closeEdit();
      setRefreshTick((v) => v + 1);
    } catch (error) {
      setEditError(error instanceof Error ? error.message : t("projects.editFailed"));
    } finally {
      setEditBusy(false);
    }
  }

  function openDelete(project: { project_id: string; name: string }) {
    setDeletingProject(project);
    setDeleteError("");
  }

  function closeDelete() {
    setDeletingProject(null);
    setDeleteBusy(false);
    setDeleteError("");
  }

  async function handleConfirmDelete() {
    if (!deletingProject) return;
    setDeleteBusy(true);
    setDeleteError("");
    try {
      await apiDelete(`/api/v3/projects/${encodeURIComponent(deletingProject.project_id)}`);
      closeDelete();
      setRefreshTick((v) => v + 1);
    } catch (error) {
      setDeleteError(error instanceof Error ? error.message : t("projects.deleteFailed"));
    } finally {
      setDeleteBusy(false);
    }
  }

  function openCreateModal() {
    setShowCreateModal(true);
    setCreateError("");
  }

  function closeCreateModal() {
    setShowCreateModal(false);
    setCreatePayload({
      name: "",
      project_category: "software_dev",
      goal_summary: "",
      workspace_root: "",
      repo_root: "",
    });
    setCreateBusy(false);
    setCreateError("");
  }

  async function handlePickDirectory(target: "workspace_root" | "repo_root") {
    const result = await selectDesktopDirectory();
    if (result.canceled || !result.path.trim()) {
      return;
    }
    setCreatePayload((current) => ({
      ...current,
      [target]: result.path,
    }));
  }

  const createFormValid = createPayload.name.trim() !== "" && createPayload.goal_summary.trim() !== "";

  async function handleCreateProject() {
    if (!createFormValid) return;
    setCreateBusy(true);
    setCreateError("");
    try {
      const result = await apiPost<CreateProjectResponse>("/api/v3/projects", createPayload);
      const nextProjectId = result.resource_id.trim();
      if (nextProjectId) {
        setStoredProjectId(nextProjectId);
        navigate(buildRoute("/workspace", { project: nextProjectId }));
      }
      closeCreateModal();
      setRefreshTick((v) => v + 1);
    } catch (error) {
      setCreateError(error instanceof Error ? error.message : t("settings.createFailed"));
    } finally {
      setCreateBusy(false);
    }
  }

  const summary = state.data?.summary;

  return (
    <QueryPanel
      loading={state.loading}
      refreshing={state.refreshing}
      stale={state.stale}
      error={state.error}
      title={t("projects.title")}
      onRetry={() => setRefreshTick((v) => v + 1)}
      recoveryMessage={t("projects.recovery")}
    >
      {state.data ? (
        <section className="dashboard-page">
          <div className="dashboard-intro">
            <div>
              <p className="placeholder-section">{t("projects.section")}</p>
              <h3 className="placeholder-title">{t("projects.title")}</h3>
              <p className="placeholder-description">
                {t("projects.description")}
              </p>
            </div>
            <div className="action-row">
              <button
                className="primary-button"
                onClick={openCreateModal}
              >
                + {t("projects.createProject")}
              </button>
              <button
                className="secondary-button"
                onClick={() => setRefreshTick((v) => v + 1)}
              >
                {t("projects.refresh")}
              </button>
            </div>
          </div>

          {summary ? (
            <div className="metrics-grid">
              <MetricCard
                label={t("projects.total")}
                value={String(summary.total_projects)}
                tone="calm"
              />
              <MetricCard
                label={t("projects.active")}
                value={String(summary.active_projects)}
                tone="neutral"
              />
              <MetricCard
                label={t("projects.blocked")}
                value={String(summary.blocked_projects)}
                tone={summary.blocked_projects > 0 ? "warn" : "neutral"}
              />
              <MetricCard
                label={t("projects.pendingActions")}
                value={String(summary.pending_actions)}
                tone={summary.pending_actions > 0 ? "warn" : "neutral"}
              />
            </div>
          ) : null}

          <section className="data-panel">
            <div className="panel-header">
              <h3>{t("projects.activeProjects")}</h3>
              <span className="status-pill">
                {state.data.active_projects.length}
              </span>
            </div>
            <div className="project-grid">
              {state.data.active_projects.map((item) => (
                <article
                  key={item.project_id}
                  className="project-card"
                  onClick={() => selectProject(item.project_id)}
                >
                  <div className="project-card-head">
                    <strong>{item.name}</strong>
                    <div className="action-row">
                      <span className="status-pill">{item.current_stage}</span>
                      <span
                        className={`status-pill ${
                          item.production_status === "ready"
                            ? "pill-success"
                            : item.production_status === "pending"
                              ? "pill-advisory"
                              : ""
                        }`}
                      >
                        {item.production_status}
                      </span>
                    </div>
                  </div>
                  <p>
                    {item.project_category} · {t("projects.progress")}{" "}
                    {item.progress_percent}%
                  </p>
                  <div className="project-progress-bar">
                    <div
                      className="project-progress-fill"
                      style={{ width: `${item.progress_percent}%` }}
                    />
                  </div>
                  <div className="action-row project-card-actions">
                    <button
                      className="secondary-button"
                      onClick={(e) => {
                        e.stopPropagation();
                        selectProject(item.project_id);
                      }}
                    >
                      {t("projects.openWorkspace")}
                    </button>
                    <button
                      className="secondary-button"
                      onClick={(e) => {
                        e.stopPropagation();
                        openEdit({ project_id: item.project_id, name: item.name });
                      }}
                    >
                      {t("projects.edit")}
                    </button>
                    <button
                      className="secondary-button"
                      style={{ color: "#c0392b" }}
                      onClick={(e) => {
                        e.stopPropagation();
                        openDelete({ project_id: item.project_id, name: item.name });
                      }}
                    >
                      {t("projects.delete")}
                    </button>
                  </div>
                </article>
              ))}
              {state.data.active_projects.length === 0 ? (
                <article className="project-card project-card-empty">
                  <div className="project-card-head">
                    <strong>{t("projects.noProjects")}</strong>
                  </div>
                  <p>{t("projects.noProjectsDescription")}</p>
                  <div className="action-row">
                    <button
                      className="primary-button"
                      onClick={openCreateModal}
                    >
                      + {t("projects.createProject")}
                    </button>
                  </div>
                </article>
              ) : null}
            </div>
          </section>

          {state.data.need_attention.length > 0 ? (
            <section className="data-panel">
              <div className="panel-header">
                <h3>{t("projects.needAttention")}</h3>
                <span className="status-pill pill-advisory">
                  {state.data.need_attention.length}
                </span>
              </div>
              <div className="stack-list">
                {state.data.need_attention.map((item) => (
                  <button
                    key={item.item_id}
                    className="action-card"
                    onClick={() => selectProject(item.project_id)}
                  >
                    <div className="action-card-head">
                      <strong>{item.title}</strong>
                      <span
                        className={`severity-badge severity-${item.severity}`}
                      >
                        {item.is_blocking
                          ? t("blocking")
                          : item.severity}
                      </span>
                    </div>
                    <p>{item.project_id}</p>
                  </button>
                ))}
              </div>
            </section>
          ) : null}

          <div className="content-grid">
            <section className="data-panel">
              <div className="panel-header">
                <h3>{t("projects.releaseReadiness")}</h3>
              </div>
              <div className="stack-list">
                {state.data.release_readiness.map((item) => (
                  <article key={item.project_id} className="list-card">
                    <div className="list-card-head">
                      <strong>{item.name}</strong>
                      <span
                        className={`status-pill ${
                          item.production_status === "ready"
                            ? "pill-success"
                            : ""
                        }`}
                      >
                        {item.production_status}
                      </span>
                    </div>
                    {item.missing_items > 0 ? (
                      <p>
                        {item.missing_items} {t("projects.missingItems")}
                      </p>
                    ) : null}
                  </article>
                ))}
                {state.data.release_readiness.length === 0 ? (
                  <article className="list-card">
                    <p>{t("projects.noReleaseItems")}</p>
                  </article>
                ) : null}
              </div>
            </section>

            <section className="data-panel">
              <div className="panel-header">
                <h3>{t("projects.recentActivity")}</h3>
              </div>
              <div className="stack-list">
                {state.data.recent_activity.slice(0, 5).map((item) => (
                  <article key={item.event_id} className="list-card">
                    <div className="list-card-head">
                      <strong>{item.title}</strong>
                      <span
                        className={
                          item.needs_attention
                            ? "severity-badge severity-warning"
                            : "status-pill"
                        }
                      >
                        {item.event_type}
                      </span>
                    </div>
                    <p>
                      {item.source_brain} · {item.occurred_at}
                    </p>
                  </article>
                ))}
                {state.data.recent_activity.length === 0 ? (
                  <article className="list-card">
                    <p>{t("projects.noRecentActivity")}</p>
                  </article>
                ) : null}
              </div>
            </section>
          </div>
        </section>
      ) : null}

      {editingProject ? (
        <div className="modal-overlay" onClick={closeEdit}>
          <div className="modal-panel" onClick={(e) => e.stopPropagation()}>
            <div className="panel-header">
              <h3>{t("projects.editProject")}</h3>
            </div>
            <div className="form-grid">
              <label className="form-field">
                <span>{t("settings.name")}</span>
                <input
                  className="project-input"
                  value={editPayload.name || ""}
                  onChange={(e) =>
                    setEditPayload((p) => ({ ...p, name: e.target.value }))
                  }
                />
              </label>
            </div>
            {editError ? (
              <p className="error-copy">
                {editError}
              </p>
            ) : null}
            <div className="action-row">
              <button
                className="primary-button"
                disabled={editBusy || !(editPayload.name || "").trim()}
                onClick={handleSaveEdit}
              >
                {editBusy ? t("projects.saving") : t("projects.save")}
              </button>
              <button className="secondary-button" onClick={closeEdit}>
                {t("projects.cancel")}
              </button>
            </div>
          </div>
        </div>
      ) : null}

      {deletingProject ? (
        <div className="modal-overlay" onClick={closeDelete}>
          <div className="modal-panel" onClick={(e) => e.stopPropagation()}>
            <div className="panel-header">
              <h3>{t("projects.deleteProject")}</h3>
            </div>
            <p>
              {t("projects.deleteConfirm", { name: deletingProject.name })}
            </p>
            {deleteError ? (
              <p className="error-copy">
                {deleteError}
              </p>
            ) : null}
            <div className="action-row">
              <button
                className="primary-button"
                disabled={deleteBusy}
                onClick={handleConfirmDelete}
                style={{ background: "#c0392b", color: "#fff" }}
              >
                {deleteBusy ? t("projects.deleting") : t("projects.delete")}
              </button>
              <button className="secondary-button" onClick={closeDelete}>
                {t("projects.cancel")}
              </button>
            </div>
          </div>
        </div>
      ) : null}

      {showCreateModal ? (
        <div className="modal-overlay" onClick={closeCreateModal}>
          <div className="modal-panel" onClick={(e) => e.stopPropagation()}>
            <div className="panel-header">
              <h3>{t("projects.createNewProject")}</h3>
            </div>
            <div className="form-grid">
              <label className="form-field">
                <span>{t("settings.name")}</span>
                <input
                  className="project-input"
                  disabled={createBusy}
                  value={createPayload.name}
                  onChange={(e) =>
                    setCreatePayload((current) => ({
                      ...current,
                      name: e.target.value,
                    }))
                  }
                />
              </label>
              <label className="form-field">
                <span>{t("settings.projectCategory")}</span>
                <select
                  className="project-input"
                  disabled={createBusy}
                  value={createPayload.project_category}
                  onChange={(e) =>
                    setCreatePayload((current) => ({
                      ...current,
                      project_category: e.target.value,
                    }))
                  }
                >
                  <option value="software_dev">{t("category.softwareDev")}</option>
                  <option value="game_dev">{t("category.gameDev")}</option>
                  <option value="novel_writing">{t("category.novelWriting")}</option>
                  <option value="anime_creation">{t("category.animeCreation")}</option>
                  <option value="short_drama">{t("category.shortDrama")}</option>
                  <option value="short_video">{t("category.shortVideo")}</option>
                  <option value="film_production">{t("category.filmProduction")}</option>
                  <option value="music_production">{t("category.musicProduction")}</option>
                  <option value="social_media_ops">{t("category.socialMediaOps")}</option>
                  <option value="marketing_campaign">{t("category.marketingCampaign")}</option>
                  <option value="ecommerce_launch">{t("category.ecommerceLaunch")}</option>
                  <option value="product_ops">{t("category.productOps")}</option>
                  <option value="website_ops">{t("category.websiteOps")}</option>
                  <option value="web_app">{t("category.webApp")}</option>
                  <option value="mobile_app">{t("category.mobileApp")}</option>
                  <option value="api_service">{t("category.apiService")}</option>
                  <option value="data_pipeline">{t("category.dataPipeline")}</option>
                </select>
              </label>
              <label className="form-field">
                <span>{t("settings.goalSummary")}</span>
                <textarea
                  className="project-input project-textarea"
                  disabled={createBusy}
                  value={createPayload.goal_summary}
                  onChange={(e) =>
                    setCreatePayload((current) => ({
                      ...current,
                      goal_summary: e.target.value,
                    }))
                  }
                />
              </label>
              <label className="form-field">
                <span>{t("settings.workspaceRoot")}</span>
                <input
                  className="project-input"
                  disabled={createBusy}
                  value={createPayload.workspace_root}
                  onChange={(e) =>
                    setCreatePayload((current) => ({
                      ...current,
                      workspace_root: e.target.value,
                    }))
                  }
                />
                <div className="action-row">
                  <button
                    className="secondary-button"
                    disabled={createBusy}
                    onClick={() => void handlePickDirectory("workspace_root")}
                  >
                    {t("settings.chooseFolder")}
                  </button>
                </div>
              </label>
              <label className="form-field">
                <span>{t("settings.repoRoot")}</span>
                <input
                  className="project-input"
                  disabled={createBusy}
                  value={createPayload.repo_root}
                  onChange={(e) =>
                    setCreatePayload((current) => ({
                      ...current,
                      repo_root: e.target.value,
                    }))
                  }
                />
                <div className="action-row">
                  <button
                    className="secondary-button"
                    disabled={createBusy}
                    onClick={() => void handlePickDirectory("repo_root")}
                  >
                    {t("settings.chooseFolder")}
                  </button>
                </div>
              </label>
            </div>
            {createError ? (
              <p className="error-copy">
                {createError}
              </p>
            ) : null}
            <div className="action-row">
              <button
                className="primary-button"
                disabled={createBusy || !createFormValid}
                onClick={() => void handleCreateProject()}
              >
                {createBusy ? t("settings.creating") : t("projects.create")}
              </button>
              <button className="secondary-button" onClick={closeCreateModal} disabled={createBusy}>
                {t("projects.cancel")}
              </button>
            </div>
          </div>
        </div>
      ) : null}
    </QueryPanel>
  );
}

function MetricCard(props: {
  label: string;
  value: string;
  tone: "calm" | "warn" | "neutral";
}) {
  return (
    <article className={`metric-card metric-${props.tone}`}>
      <span>{props.label}</span>
      <strong>{props.value}</strong>
    </article>
  );
}
