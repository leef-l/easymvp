import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import { apiGet, apiPost } from "@/shared/lib/api";
import { useProjectState } from "@/shared/lib/project";
import { useQuery } from "@/shared/lib/query";
import type { CommandResponse, SolutionDesign } from "@/shared/lib/types";
import { QueryPanel } from "@/shared/ui/QueryPanel";

export function DesignPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { projectId, routes } = useProjectState();

  if (!projectId) {
    return (
      <section className="placeholder-page">
        <div className="empty-state-panel">
          <h4>{t("design.noProjectTitle")}</h4>
          <p>{t("design.noProjectDescription")}</p>
          <div className="action-row">
            <button className="primary-button" onClick={() => navigate("/projects")}>
              {t("design.goToProjects")}
            </button>
          </div>
        </div>
      </section>
    );
  }

  const [refreshTick, setRefreshTick] = useState(0);
  const [busyAction, setBusyAction] = useState("");
  const [actionError, setActionError] = useState("");
  const [actionMessage, setActionMessage] = useState("");

  const state = useQuery(
    () =>
      apiGet<SolutionDesign>(
        `/api/v3/projects/${encodeURIComponent(projectId)}/design`,
      ),
    [projectId, refreshTick],
  );

  async function runAction(
    actionKey: string,
    execute: () => Promise<CommandResponse>,
  ) {
    setBusyAction(actionKey);
    setActionError("");
    setActionMessage("");
    try {
      const result = await execute();
      setActionMessage(`${actionKey} accepted · ${result.next_action}`);
      setRefreshTick((v) => v + 1);
    } catch (error) {
      setActionError(
        error instanceof Error ? error.message : t("design.actionFailed"),
      );
    } finally {
      setBusyAction("");
    }
  }

  function handleConfirm() {
    if (!state.data?.id) return;
    void runAction("confirm_design", () =>
      apiPost<CommandResponse>(
        `/api/v3/designs/${encodeURIComponent(state.data!.id)}/confirm`,
        { project_id: projectId },
      ),
    );
  }

  function handleReject() {
    if (!state.data?.id) return;
    void runAction("reject_design", () =>
      apiPost<CommandResponse>(
        `/api/v3/designs/${encodeURIComponent(state.data!.id)}/reject`,
        { project_id: projectId },
      ),
    );
  }

  function handleRegenerate() {
    void runAction("regenerate_design", () =>
      apiPost<CommandResponse>("/api/v3/designs/generate", {
        project_id: projectId,
      }),
    );
  }

  const design = state.data;
  const statusLabel = design?.status || "draft";

  return (
    <QueryPanel
      loading={state.loading}
      refreshing={state.refreshing}
      stale={state.stale}
      error={state.error}
      title={t("design.title")}
      onRetry={() => setRefreshTick((v) => v + 1)}
      recoveryMessage={t("design.recovery")}
    >
      <section className="dashboard-page">
        <div className="dashboard-intro">
          <div>
            <p className="placeholder-section">{t("design.section")}</p>
            <h3 className="placeholder-title">{t("design.title")}</h3>
            <p className="placeholder-description">
              {t("design.description")}
            </p>
          </div>
          <div className="action-row">
            <span
              className={`status-pill ${statusLabel === "confirmed" ? "pill-success" : statusLabel === "rejected" ? "pill-blocking" : ""}`}
            >
              {statusLabel}
            </span>
            <button
              className="secondary-button"
              onClick={() => setRefreshTick((v) => v + 1)}
            >
              {t("design.refresh")}
            </button>
          </div>
        </div>

        {actionError ? (
          <section className="data-panel is-error">
            <p className="error-copy">{actionError}</p>
          </section>
        ) : null}
        {actionMessage ? (
          <section className="data-panel">
            <p className="muted-copy">{actionMessage}</p>
          </section>
        ) : null}

        {/* Action Buttons */}
        {design && statusLabel !== "confirmed" ? (
          <section className="data-panel">
            <div className="panel-header">
              <h3>{t("design.actions")}</h3>
            </div>
            <div className="action-row">
              <button
                className="primary-button"
                disabled={busyAction !== "" || statusLabel === "rejected"}
                onClick={handleConfirm}
              >
                {busyAction === "confirm_design"
                  ? t("design.confirming")
                  : t("design.confirmButton")}
              </button>
              <button
                className="secondary-button"
                disabled={busyAction !== ""}
                onClick={handleReject}
              >
                {busyAction === "reject_design"
                  ? t("design.rejecting")
                  : t("design.rejectButton")}
              </button>
              <button
                className="secondary-button"
                disabled={busyAction !== ""}
                onClick={handleRegenerate}
              >
                {busyAction === "regenerate_design"
                  ? t("design.regenerating")
                  : t("design.regenerateButton")}
              </button>
            </div>
          </section>
        ) : null}

        {design ? (
          <>
            {/* Architecture Description */}
            {design.architecture_description ? (
              <section className="data-panel">
                <div className="panel-header">
                  <h3>{t("design.architecture")}</h3>
                </div>
                <div style={{ padding: "12px 16px" }}>
                  <p>{design.architecture_description}</p>
                </div>
              </section>
            ) : null}

            {/* Modules */}
            <section className="data-panel">
              <div className="panel-header">
                <h3>{t("design.modules")}</h3>
                <span className="status-pill">
                  {design.modules?.length || 0}
                </span>
              </div>
              <div className="stack-list">
                {(design.modules || []).map((mod, idx) => (
                  <article key={idx} className="list-card">
                    <div className="list-card-head">
                      <strong>{mod.module_name}</strong>
                    </div>
                    <p>{mod.description}</p>
                    <p className="muted-copy">{mod.responsibility}</p>
                  </article>
                ))}
                {!design.modules?.length ? (
                  <article className="list-card">
                    <p className="muted-copy">{t("design.noModules")}</p>
                  </article>
                ) : null}
              </div>
            </section>

            <div className="content-grid">
              {/* Data Models */}
              <section className="data-panel">
                <div className="panel-header">
                  <h3>{t("design.dataModels")}</h3>
                  <span className="status-pill">
                    {design.data_models?.length || 0}
                  </span>
                </div>
                <div className="stack-list">
                  {(design.data_models || []).map((model, idx) => (
                    <article key={idx} className="list-card">
                      <div className="list-card-head">
                        <strong>{model.model_name}</strong>
                      </div>
                      <p>{model.description}</p>
                      <div className="inline-metrics">
                        {model.fields.map((field) => (
                          <span key={field} className="status-pill">
                            {field}
                          </span>
                        ))}
                      </div>
                    </article>
                  ))}
                  {!design.data_models?.length ? (
                    <article className="list-card">
                      <p className="muted-copy">{t("design.noDataModels")}</p>
                    </article>
                  ) : null}
                </div>
              </section>

              {/* Page Designs */}
              <section className="data-panel">
                <div className="panel-header">
                  <h3>{t("design.pageDesigns")}</h3>
                  <span className="status-pill">
                    {design.page_designs?.length || 0}
                  </span>
                </div>
                <div className="stack-list">
                  {(design.page_designs || []).map((page, idx) => (
                    <article key={idx} className="list-card">
                      <div className="list-card-head">
                        <strong>{page.page_name}</strong>
                        <span className="status-pill">{page.route}</span>
                      </div>
                      <p>{page.description}</p>
                      <div className="inline-metrics">
                        {page.components.map((comp) => (
                          <span key={comp} className="status-pill">
                            {comp}
                          </span>
                        ))}
                      </div>
                    </article>
                  ))}
                  {!design.page_designs?.length ? (
                    <article className="list-card">
                      <p className="muted-copy">
                        {t("design.noPageDesigns")}
                      </p>
                    </article>
                  ) : null}
                </div>
              </section>
            </div>

            {/* Task Drafts */}
            <section className="data-panel">
              <div className="panel-header">
                <h3>{t("design.taskDrafts")}</h3>
                <span className="status-pill">
                  {design.task_drafts?.length || 0}
                </span>
              </div>
              <div className="stack-list">
                {(design.task_drafts || []).map((task, idx) => (
                  <article key={idx} className="list-card">
                    <div className="list-card-head">
                      <strong>{task.task_name}</strong>
                      <div className="action-row">
                        <span className="status-pill">{task.priority}</span>
                        <span className="status-pill">
                          {task.estimated_hours}h
                        </span>
                      </div>
                    </div>
                    <p>{task.description}</p>
                    {task.dependencies.length > 0 ? (
                      <div className="inline-metrics">
                        {task.dependencies.map((dep) => (
                          <span key={dep} className="status-pill">
                            {dep}
                          </span>
                        ))}
                      </div>
                    ) : null}
                  </article>
                ))}
                {!design.task_drafts?.length ? (
                  <article className="list-card">
                    <p className="muted-copy">{t("design.noTaskDrafts")}</p>
                  </article>
                ) : null}
              </div>
            </section>
          </>
        ) : (
          <section className="data-panel">
            <div className="panel-header">
              <h3>{t("design.noDesign")}</h3>
            </div>
            <p className="muted-copy">{t("design.noDesignHint")}</p>
            <div className="action-row" style={{ marginTop: 12 }}>
              <button
                className="primary-button"
                disabled={busyAction !== ""}
                onClick={handleRegenerate}
              >
                {busyAction === "regenerate_design"
                  ? t("design.regenerating")
                  : t("design.regenerateButton")}
              </button>
            </div>
          </section>
        )}
      </section>
    </QueryPanel>
  );
}
