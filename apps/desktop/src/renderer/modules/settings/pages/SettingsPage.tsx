import { useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { apiGet, apiPost } from "@/shared/lib/api";
import { useProjectState } from "@/shared/lib/project";
import { useQuery } from "@/shared/lib/query";
import {
  getDesktopRuntimeInfo,
  getDesktopRuntimeRecoveryHint,
  getStoredCoreBaseUrl,
  openDesktopPath,
  selectDesktopDirectory,
  setStoredCoreBaseUrl,
  setStoredProjectId,
  showDesktopItemInFolder,
} from "@/shared/lib/preferences";
import type {
  CreateProjectPayload,
  CreateProjectResponse,
  RuntimeHealthView,
  SystemHealthView,
} from "@/shared/lib/types";
import { QueryPanel } from "@/shared/ui/QueryPanel";

export function SettingsPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { projectId, updateProjectId, buildRoute, routes } = useProjectState();
  const [draftProjectId, setDraftProjectId] = useState(projectId);
  const [draftBaseUrl, setDraftBaseUrl] = useState(getStoredCoreBaseUrl());
  const [saveTick, setSaveTick] = useState(0);
  const [createBusy, setCreateBusy] = useState(false);
  const [createError, setCreateError] = useState("");
  const [createMessage, setCreateMessage] = useState("");
  const [createdProjectId, setCreatedProjectId] = useState("");
  const [createPayload, setCreatePayload] = useState<CreateProjectPayload>({
    name: "",
    project_category: "software_dev",
    goal_summary: "",
    workspace_root: "",
    repo_root: "",
  });

  const systemState = useQuery(
    () => apiGet<SystemHealthView>("/api/v3/system/healthz"),
    [saveTick],
  );
  const runtimeState = useQuery(
    () => apiGet<RuntimeHealthView>("/api/v3/runtime/healthz"),
    [saveTick],
  );
  const desktopRuntimeState = useQuery(() => getDesktopRuntimeInfo(), []);

  const saveSummary = useMemo(() => {
    return t("settings.runtimeNote", { url: getStoredCoreBaseUrl() });
  }, [projectId, saveTick, t]);

  function handleSave() {
    setStoredCoreBaseUrl(draftBaseUrl);
    setStoredProjectId(draftProjectId);
    updateProjectId(draftProjectId);
    setSaveTick((value) => value + 1);
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

  async function handleOpenDataDirectory(mode: "open" | "show") {
    const targetPath = desktopRuntimeState.data?.dataDirectory?.trim() || "";
    if (!targetPath) {
      setCreateError(t("settings.dataDirUnavailable"));
      return;
    }
    const result =
      mode === "show"
        ? await showDesktopItemInFolder(targetPath)
        : await openDesktopPath(targetPath);
    if (!result.ok) {
      setCreateError(result.error || t("settings.pathActionFailed"));
      return;
    }
    setCreateError("");
  }

  const createFormValid = createPayload.name.trim() !== "" && createPayload.goal_summary.trim() !== "";

  async function handleCreateProject() {
    if (!createFormValid) {
      return;
    }
    setCreateBusy(true);
    setCreateError("");
    setCreateMessage("");
    setCreatedProjectId("");
    try {
      const result = await apiPost<CreateProjectResponse>(
        "/api/v3/projects",
        createPayload,
      );
      const nextProjectId = result.resource_id.trim();
      if (nextProjectId) {
        setCreatedProjectId(nextProjectId);
        setDraftProjectId(nextProjectId);
        updateProjectId(nextProjectId);
        setStoredProjectId(nextProjectId);
        navigate(buildRoute("/workspace", { project: nextProjectId }));
      }
      setCreateMessage(
        `${result.next_action} · ${nextProjectId || t("settings.projectCreated")}`,
      );
      setCreatePayload({
        name: "",
        project_category: createPayload.project_category,
        goal_summary: "",
        workspace_root: createPayload.workspace_root,
        repo_root: createPayload.repo_root,
      });
      setSaveTick((value) => value + 1);
    } catch (error) {
      setCreateError(
        error instanceof Error ? error.message : t("settings.createFailed"),
      );
    } finally {
      setCreateBusy(false);
    }
  }

  return (
    <QueryPanel
      loading={systemState.loading || runtimeState.loading}
      error={systemState.error || runtimeState.error}
      title={t("settings.title")}
      onRetry={() => setSaveTick((value) => value + 1)}
      secondaryActionLabel={t("settings.openDiagnostics")}
      onSecondaryAction={() => navigate(routes.diagnostics)}
      recoveryMessage={t("settings.recovery")}
    >
      {systemState.data && runtimeState.data ? (
        <section className="dashboard-page">
          <div className="dashboard-intro">
            <div>
              <p className="placeholder-section">{t("settings.section")}</p>
              <h3 className="placeholder-title">{t("settings.subtitle")}</h3>
              <p className="placeholder-description">{saveSummary}</p>
            </div>
            <button className="primary-button" onClick={handleSave}>
              {t("settings.save")}
            </button>
          </div>

          <div className="content-grid">
            <section className="data-panel">
              <div className="panel-header">
                <h3>{t("settings.defaults")}</h3>
              </div>
              <div className="form-grid">
                <label className="form-field">
                  <span>{t("settings.projectId")}</span>
                  <input
                    className="project-input"
                    value={draftProjectId}
                    onChange={(event) => setDraftProjectId(event.target.value)}
                  />
                </label>
                <label className="form-field">
                  <span>{t("settings.coreUrl")}</span>
                  <input
                    className="project-input"
                    value={draftBaseUrl}
                    onChange={(event) => setDraftBaseUrl(event.target.value)}
                  />
                </label>
              </div>
              {desktopRuntimeState.data ? (
                <div className="runtime-note">
                  <div className="runtime-note-header">
                    <strong>{t("settings.runtime")}</strong>
                    <span className="status-pill">
                      {desktopRuntimeState.data.source}
                    </span>
                  </div>
                  <div className="runtime-grid">
                    <div className="runtime-field">
                      <span>{t("settings.launchMode")}</span>
                      <strong>{desktopRuntimeState.data.launchMode}</strong>
                    </div>
                    <div className="runtime-field">
                      <span>{t("settings.packaged")}</span>
                      <strong>
                        {desktopRuntimeState.data.packaged ? t("yes") : t("no")}
                      </strong>
                    </div>
                    <div className="runtime-field runtime-field-wide">
                      <span>{t("settings.runtimeCoreUrl")}</span>
                      <strong>{desktopRuntimeState.data.coreBaseUrl}</strong>
                    </div>
                    <div className="runtime-field">
                      <span>{t("settings.coreProbe")}</span>
                      <strong>{desktopRuntimeState.data.coreStatus}</strong>
                    </div>
                    <div className="runtime-field">
                      <span>{t("settings.httpStatus")}</span>
                      <strong>
                        {desktopRuntimeState.data.coreHttpStatus || "n/a"}
                      </strong>
                    </div>
                    <div className="runtime-field">
                      <span>{t("settings.platform")}</span>
                      <strong>{desktopRuntimeState.data.platform}</strong>
                    </div>
                    <div className="runtime-field">
                      <span>{t("settings.version")}</span>
                      <strong>{desktopRuntimeState.data.version}</strong>
                    </div>
                    <div className="runtime-field">
                      <span>{t("settings.coreManager")}</span>
                      <strong>
                        {desktopRuntimeState.data.coreManagerStatus}
                      </strong>
                    </div>
                  </div>
                  <p className="muted-copy">
                    {t("settings.runtimeNote", { url: getStoredCoreBaseUrl() })}
                  </p>
                  {desktopRuntimeState.data.coreError ? (
                    <p className="error-copy">
                      {desktopRuntimeState.data.coreError}
                    </p>
                  ) : null}
                  <p
                    className={
                      desktopRuntimeState.data.issue
                        ? "error-copy"
                        : "muted-copy"
                    }
                  >
                    {desktopRuntimeState.data.issue ||
                      getDesktopRuntimeRecoveryHint(desktopRuntimeState.data)}
                  </p>
                </div>
              ) : null}
            </section>

            <section className="data-panel">
              <div className="panel-header">
                <h3>{t("settings.createProject")}</h3>
                <span className="status-pill">
                  {createBusy
                    ? t("settings.creating")
                    : createdProjectId
                      ? t("settings.created")
                      : t("settings.ready")}
                </span>
              </div>
              <div className="form-grid">
                <label className="form-field">
                  <span>{t("settings.name")}</span>
                  <input
                    className="project-input"
                    disabled={createBusy}
                    value={createPayload.name}
                    onChange={(event) =>
                      setCreatePayload((current) => ({
                        ...current,
                        name: event.target.value,
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
                    onChange={(event) =>
                      setCreatePayload((current) => ({
                        ...current,
                        project_category: event.target.value,
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
                    onChange={(event) =>
                      setCreatePayload((current) => ({
                        ...current,
                        goal_summary: event.target.value,
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
                    onChange={(event) =>
                      setCreatePayload((current) => ({
                        ...current,
                        workspace_root: event.target.value,
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
                    onChange={(event) =>
                      setCreatePayload((current) => ({
                        ...current,
                        repo_root: event.target.value,
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
                <div className="action-row">
                  <button
                    className="primary-button"
                    disabled={createBusy || !createFormValid}
                    onClick={() => void handleCreateProject()}
                  >
                    {createBusy ? t("settings.creating") : t("settings.create")}
                  </button>
                  {createdProjectId ? (
                    <>
                      <button
                        className="secondary-button"
                        onClick={() =>
                          navigate(
                            buildRoute("/workspace", {
                              project: createdProjectId,
                            }),
                          )
                        }
                      >
                        {t("settings.openWorkspace")}
                      </button>
                      <button
                        className="secondary-button"
                        onClick={() =>
                          navigate(
                            buildRoute("/plan", { project: createdProjectId }),
                          )
                        }
                      >
                        {t("settings.openPlan")}
                      </button>
                      <button
                        className="secondary-button"
                        onClick={() =>
                          navigate(
                            buildRoute("/execution", {
                              project: createdProjectId,
                            }),
                          )
                        }
                      >
                        {t("settings.openExecution")}
                      </button>
                    </>
                  ) : null}
                </div>
                {createError ? (
                  <p className="error-copy">{createError}</p>
                ) : null}
                {createMessage ? (
                  <p className="muted-copy">{createMessage}</p>
                ) : null}
                {!createError && !createMessage ? (
                  <p className="muted-copy">
                    {t("settings.createHint")}
                  </p>
                ) : null}
              </div>
            </section>

            <section className="data-panel">
              <div className="panel-header">
                <h3>{t("settings.healthChecks")}</h3>
              </div>
              <div className="stack-list">
                <article className="list-card">
                  <div className="list-card-head">
                    <strong>{t("settings.system")}</strong>
                    <span className="status-pill">
                      {systemState.data.status}
                    </span>
                  </div>
                  <p>
                    {systemState.data.service} · {systemState.data.version} ·{" "}
                    {systemState.data.timestamp}
                  </p>
                </article>
                <article className="list-card">
                  <div className="list-card-head">
                    <strong>{t("settings.runtimeHealth")}</strong>
                    <span className="status-pill">
                      {runtimeState.data.status}
                    </span>
                  </div>
                  <p>{runtimeState.data.base_url}</p>
                </article>
              </div>
              {desktopRuntimeState.data?.dataDirectory ? (
                <div className="action-row">
                  <button
                    className="secondary-button"
                    onClick={() => void handleOpenDataDirectory("show")}
                  >
                    {t("settings.showDataFolder")}
                  </button>
                  <button
                    className="secondary-button"
                    onClick={() => void handleOpenDataDirectory("open")}
                  >
                    {t("settings.openDataPath")}
                  </button>
                </div>
              ) : null}
            </section>
          </div>
        </section>
      ) : null}
    </QueryPanel>
  );
}
