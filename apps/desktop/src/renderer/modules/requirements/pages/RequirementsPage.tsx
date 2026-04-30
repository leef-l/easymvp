import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import { apiGet, apiPost } from "@/shared/lib/api";
import { useProjectState } from "@/shared/lib/project";
import { useQuery } from "@/shared/lib/query";
import type { CommandResponse, RequirementAnalysis } from "@/shared/lib/types";
import { QueryPanel } from "@/shared/ui/QueryPanel";

export function RequirementsPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { projectId } = useProjectState();

  if (!projectId) {
    return (
      <section className="placeholder-page">
        <div className="empty-state-panel">
          <h4>{t("requirements.noProjectTitle")}</h4>
          <p>{t("requirements.noProjectDescription")}</p>
          <div className="action-row">
            <button className="primary-button" onClick={() => navigate("/projects")}>
              {t("requirements.goToProjects")}
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
  const [rawInput, setRawInput] = useState("");

  const state = useQuery(
    () =>
      apiGet<RequirementAnalysis>(
        `/api/v3/projects/${encodeURIComponent(projectId)}/requirements`,
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
        error instanceof Error ? error.message : t("requirements.actionFailed"),
      );
    } finally {
      setBusyAction("");
    }
  }

  function handleAnalyze() {
    if (!rawInput.trim()) return;
    void runAction("analyze_requirements", () =>
      apiPost<CommandResponse>("/api/v3/requirements/analyze", {
        project_id: projectId,
        raw_input: rawInput.trim(),
      }),
    );
  }

  function handleConfirm() {
    if (!state.data?.id) return;
    void runAction("confirm_requirements", () =>
      apiPost<CommandResponse>(
        `/api/v3/requirements/${encodeURIComponent(state.data!.id)}/confirm`,
        { project_id: projectId },
      ),
    );
  }

  const requirement = state.data;
  const statusLabel = requirement?.status || "draft";

  return (
    <QueryPanel
      loading={state.loading}
      refreshing={state.refreshing}
      stale={state.stale}
      error={state.error}
      title={t("requirements.title")}
      onRetry={() => setRefreshTick((v) => v + 1)}
      recoveryMessage={t("requirements.recovery")}
    >
      <section className="dashboard-page">
        <div className="dashboard-intro">
          <div>
            <p className="placeholder-section">{t("requirements.section")}</p>
            <h3 className="placeholder-title">{t("requirements.title")}</h3>
            <p className="placeholder-description">
              {t("requirements.description")}
            </p>
          </div>
          <div className="action-row">
            <span className={`status-pill ${statusLabel === "confirmed" ? "pill-success" : statusLabel === "analyzing" ? "pill-advisory" : ""}`}>
              {statusLabel}
            </span>
            <button
              className="secondary-button"
              onClick={() => setRefreshTick((v) => v + 1)}
            >
              {t("requirements.refresh")}
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

        {/* Input Section */}
        <section className="data-panel">
          <div className="panel-header">
            <h3>{t("requirements.inputTitle")}</h3>
          </div>
          <div className="form-grid">
            <label className="form-field">
              <span>{t("requirements.rawInputLabel")}</span>
              <textarea
                className="project-input project-textarea"
                rows={8}
                placeholder={t("requirements.rawInputPlaceholder")}
                value={rawInput || requirement?.raw_input || ""}
                onChange={(e) => setRawInput(e.target.value)}
                disabled={busyAction !== "" || statusLabel === "confirmed"}
              />
            </label>
            <div className="action-row">
              <button
                className="primary-button"
                disabled={
                  busyAction !== "" ||
                  !rawInput.trim() ||
                  statusLabel === "confirmed"
                }
                onClick={handleAnalyze}
              >
                {busyAction === "analyze_requirements"
                  ? t("requirements.analyzing")
                  : t("requirements.analyzeButton")}
              </button>
            </div>
          </div>
        </section>

        {/* Analysis Results */}
        {requirement &&
        (requirement.core_features?.length ||
          requirement.non_functional_requirements?.length ||
          requirement.tech_stack?.length ||
          requirement.constraints?.length) ? (
          <div className="content-grid">
            {/* Core Features */}
            <section className="data-panel">
              <div className="panel-header">
                <h3>{t("requirements.coreFeatures")}</h3>
                <span className="status-pill">
                  {requirement.core_features?.length || 0}
                </span>
              </div>
              <div className="stack-list">
                {(requirement.core_features || []).map((feature, idx) => (
                  <article key={idx} className="list-card">
                    <p>{feature}</p>
                  </article>
                ))}
                {!requirement.core_features?.length ? (
                  <article className="list-card">
                    <p className="muted-copy">
                      {t("requirements.noFeatures")}
                    </p>
                  </article>
                ) : null}
              </div>
            </section>

            {/* Non-Functional Requirements */}
            <section className="data-panel">
              <div className="panel-header">
                <h3>{t("requirements.nonFunctional")}</h3>
                <span className="status-pill">
                  {requirement.non_functional_requirements?.length || 0}
                </span>
              </div>
              <div className="stack-list">
                {(requirement.non_functional_requirements || []).map(
                  (item, idx) => (
                    <article key={idx} className="list-card">
                      <p>{item}</p>
                    </article>
                  ),
                )}
                {!requirement.non_functional_requirements?.length ? (
                  <article className="list-card">
                    <p className="muted-copy">
                      {t("requirements.noNonFunctional")}
                    </p>
                  </article>
                ) : null}
              </div>
            </section>

            {/* Tech Stack */}
            <section className="data-panel">
              <div className="panel-header">
                <h3>{t("requirements.techStack")}</h3>
                <span className="status-pill">
                  {requirement.tech_stack?.length || 0}
                </span>
              </div>
              <div className="stack-list">
                {(requirement.tech_stack || []).map((item, idx) => (
                  <article key={idx} className="list-card">
                    <p>{item}</p>
                  </article>
                ))}
                {!requirement.tech_stack?.length ? (
                  <article className="list-card">
                    <p className="muted-copy">
                      {t("requirements.noTechStack")}
                    </p>
                  </article>
                ) : null}
              </div>
            </section>

            {/* Constraints */}
            <section className="data-panel">
              <div className="panel-header">
                <h3>{t("requirements.constraints")}</h3>
                <span className="status-pill">
                  {requirement.constraints?.length || 0}
                </span>
              </div>
              <div className="stack-list">
                {(requirement.constraints || []).map((item, idx) => (
                  <article key={idx} className="list-card">
                    <p>{item}</p>
                  </article>
                ))}
                {!requirement.constraints?.length ? (
                  <article className="list-card">
                    <p className="muted-copy">
                      {t("requirements.noConstraints")}
                    </p>
                  </article>
                ) : null}
              </div>
            </section>
          </div>
        ) : null}

        {/* Confirm Section */}
        {requirement && statusLabel === "analyzing" ? (
          <section className="data-panel">
            <div className="panel-header">
              <h3>{t("requirements.confirmTitle")}</h3>
            </div>
            <p className="muted-copy">{t("requirements.confirmHint")}</p>
            <div className="action-row" style={{ marginTop: 12 }}>
              <button
                className="primary-button"
                disabled={busyAction !== ""}
                onClick={handleConfirm}
              >
                {busyAction === "confirm_requirements"
                  ? t("requirements.confirming")
                  : t("requirements.confirmButton")}
              </button>
            </div>
          </section>
        ) : null}

        {requirement && statusLabel === "confirmed" ? (
          <section className="data-panel">
            <div className="panel-header">
              <h3>{t("requirements.confirmedTitle")}</h3>
              <span className="status-pill pill-success">
                {t("requirements.confirmedStatus")}
              </span>
            </div>
            <p className="muted-copy">{t("requirements.confirmedHint")}</p>
          </section>
        ) : null}
      </section>
    </QueryPanel>
  );
}
