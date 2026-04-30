import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import { apiGet, apiPost } from "@/shared/lib/api";
import { useProjectState } from "@/shared/lib/project";
import { useQuery } from "@/shared/lib/query";
import type { CommandResponse, ReviewLoopView } from "@/shared/lib/types";
import { QueryPanel } from "@/shared/ui/QueryPanel";

export function ReviewPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { projectId, routes } = useProjectState();

  if (!projectId) {
    return (
      <section className="placeholder-page">
        <div className="empty-state-panel">
          <h4>{t("review.noProjectTitle")}</h4>
          <p>{t("review.noProjectDescription")}</p>
          <div className="action-row">
            <button className="primary-button" onClick={() => navigate("/projects")}>
              {t("review.goToProjects")}
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
  const [interveneAction, setInterveneAction] = useState<
    "override_approve" | "abort" | "restart"
  >("override_approve");
  const [interveneReason, setInterveneReason] = useState("");
  const [showIntervene, setShowIntervene] = useState(false);

  const state = useQuery(
    () =>
      apiGet<ReviewLoopView>(
        `/api/v3/projects/${encodeURIComponent(projectId)}/review-loop`,
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
        error instanceof Error ? error.message : t("review.actionFailed"),
      );
    } finally {
      setBusyAction("");
    }
  }

  function handleStartReview() {
    void runAction("start_review_loop", () =>
      apiPost<CommandResponse>("/api/v3/reviews/loop", {
        project_id: projectId,
      }),
    );
  }

  function handleIntervene() {
    void runAction("human_intervene", () =>
      apiPost<CommandResponse>("/api/v3/reviews/intervene", {
        project_id: projectId,
        action: interveneAction,
        reason: interveneReason.trim() || undefined,
      }),
    );
    setShowIntervene(false);
    setInterveneReason("");
  }

  const review = state.data;
  const statusLabel = review?.status || "idle";
  const rounds = review?.rounds || [];

  return (
    <QueryPanel
      loading={state.loading}
      refreshing={state.refreshing}
      stale={state.stale}
      error={state.error}
      title={t("review.title")}
      onRetry={() => setRefreshTick((v) => v + 1)}
      recoveryMessage={t("review.recovery")}
    >
      <section className="dashboard-page">
        <div className="dashboard-intro">
          <div>
            <p className="placeholder-section">{t("review.section")}</p>
            <h3 className="placeholder-title">{t("review.title")}</h3>
            <p className="placeholder-description">
              {t("review.description")}
            </p>
          </div>
          <div className="action-row">
            <span
              className={`status-pill ${statusLabel === "passed" ? "pill-success" : statusLabel === "failed" ? "pill-blocking" : statusLabel === "running" ? "pill-advisory" : ""}`}
            >
              {statusLabel}
            </span>
            <button
              className="secondary-button"
              onClick={() => setRefreshTick((v) => v + 1)}
            >
              {t("review.refresh")}
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

        {/* Overview Metrics */}
        {review ? (
          <div className="metrics-grid">
            <MetricCard
              label={t("review.currentRound")}
              value={`${review.current_round} / ${review.max_rounds}`}
              tone="neutral"
            />
            <MetricCard
              label={t("review.totalRounds")}
              value={String(rounds.length)}
              tone="neutral"
            />
            <MetricCard
              label={t("review.finalScore")}
              value={review.final_score !== undefined ? String(review.final_score) : "--"}
              tone={
                review.final_score !== undefined && review.final_score >= 80
                  ? "calm"
                  : review.final_score !== undefined && review.final_score < 60
                    ? "warn"
                    : "neutral"
              }
            />
            <MetricCard
              label={t("review.finalDecision")}
              value={review.final_decision || "--"}
              tone={
                review.final_decision === "approved"
                  ? "calm"
                  : review.final_decision === "rejected"
                    ? "warn"
                    : "neutral"
              }
            />
          </div>
        ) : null}

        {/* Action Section */}
        <section className="data-panel">
          <div className="panel-header">
            <h3>{t("review.actions")}</h3>
          </div>
          <div className="action-row">
            <button
              className="primary-button"
              disabled={busyAction !== "" || statusLabel === "running"}
              onClick={handleStartReview}
            >
              {busyAction === "start_review_loop"
                ? t("review.starting")
                : t("review.startReview")}
            </button>
            <button
              className="secondary-button"
              disabled={busyAction !== ""}
              onClick={() => setShowIntervene(!showIntervene)}
            >
              {t("review.humanIntervene")}
            </button>
          </div>

          {showIntervene ? (
            <div className="form-grid" style={{ marginTop: 12 }}>
              <label className="form-field">
                <span>{t("review.interveneAction")}</span>
                <select
                  className="project-input"
                  value={interveneAction}
                  onChange={(e) =>
                    setInterveneAction(
                      e.target.value as
                        | "override_approve"
                        | "abort"
                        | "restart",
                    )
                  }
                >
                  <option value="override_approve">
                    {t("review.overrideApprove")}
                  </option>
                  <option value="abort">{t("review.abort")}</option>
                  <option value="restart">{t("review.restart")}</option>
                </select>
              </label>
              <label className="form-field">
                <span>{t("review.interveneReason")}</span>
                <input
                  className="project-input"
                  value={interveneReason}
                  onChange={(e) => setInterveneReason(e.target.value)}
                  placeholder={t("review.interveneReasonPlaceholder")}
                />
              </label>
              <div className="action-row">
                <button
                  className="primary-button"
                  disabled={busyAction !== ""}
                  onClick={handleIntervene}
                >
                  {busyAction === "human_intervene"
                    ? t("review.intervening")
                    : t("review.confirmIntervene")}
                </button>
                <button
                  className="secondary-button"
                  onClick={() => {
                    setShowIntervene(false);
                    setInterveneReason("");
                  }}
                >
                  {t("review.cancel")}
                </button>
              </div>
            </div>
          ) : null}
        </section>

        {/* Human Intervention History */}
        {review?.human_intervention ? (
          <section className="data-panel">
            <div className="panel-header">
              <h3>{t("review.humanInterventionHistory")}</h3>
              <span className="status-pill pill-advisory">
                {review.human_intervention.action}
              </span>
            </div>
            <p>
              {review.human_intervention.reason ||
                t("review.noReasonProvided")}
            </p>
            {review.human_intervention.performed_at ? (
              <p className="muted-copy">
                {review.human_intervention.performed_at}
              </p>
            ) : null}
          </section>
        ) : null}

        {/* Review Rounds Timeline */}
        <section className="data-panel">
          <div className="panel-header">
            <h3>{t("review.roundsTimeline")}</h3>
            <span className="status-pill">{rounds.length}</span>
          </div>
          <div className="stack-list">
            {rounds.map((round) => (
              <article
                key={round.round_number}
                className="list-card"
              >
                <div className="list-card-head">
                  <strong>
                    {t("review.roundLabel", {
                      number: round.round_number,
                    })}
                  </strong>
                  <div className="action-row">
                    <span
                      className={`status-pill ${round.passed ? "pill-success" : round.status === "failed" ? "pill-blocking" : "pill-advisory"}`}
                    >
                      {round.passed
                        ? t("review.passed")
                        : round.status}
                    </span>
                    <span className="status-pill">
                      {t("review.score")}: {round.score}
                    </span>
                  </div>
                </div>

                {/* Issues */}
                {round.issues.length > 0 ? (
                  <div style={{ marginTop: 8 }}>
                    <p
                      className="muted-copy"
                      style={{ fontWeight: 700, marginBottom: 4 }}
                    >
                      {t("review.issues")} ({round.issues.length})
                    </p>
                    {round.issues.map((issue) => (
                      <div
                        key={issue.issue_id}
                        style={{
                          padding: "6px 0",
                          borderBottom: "1px solid rgba(0,0,0,0.05)",
                        }}
                      >
                        <div className="list-card-head">
                          <span>{issue.description}</span>
                          <span
                            className={`severity-badge severity-${issue.severity}`}
                          >
                            {issue.severity}
                          </span>
                        </div>
                        <p className="muted-copy">
                          {issue.category}
                          {issue.suggestion
                            ? ` · ${issue.suggestion}`
                            : ""}
                        </p>
                      </div>
                    ))}
                  </div>
                ) : null}

                {/* Corrections */}
                {round.corrections.length > 0 ? (
                  <div style={{ marginTop: 8 }}>
                    <p
                      className="muted-copy"
                      style={{ fontWeight: 700, marginBottom: 4 }}
                    >
                      {t("review.corrections")} ({round.corrections.length})
                    </p>
                    {round.corrections.map((correction) => (
                      <div
                        key={correction.correction_id}
                        style={{
                          padding: "6px 0",
                          borderBottom: "1px solid rgba(0,0,0,0.05)",
                        }}
                      >
                        <div className="list-card-head">
                          <span>{correction.description}</span>
                          <span
                            className={`status-pill ${correction.applied ? "pill-success" : ""}`}
                          >
                            {correction.applied
                              ? t("review.applied")
                              : t("review.pending")}
                          </span>
                        </div>
                      </div>
                    ))}
                  </div>
                ) : null}

                {/* Timestamps */}
                <div className="inline-metrics" style={{ marginTop: 8 }}>
                  {round.started_at ? (
                    <span className="status-pill">{round.started_at}</span>
                  ) : null}
                  {round.finished_at ? (
                    <span className="status-pill">{round.finished_at}</span>
                  ) : null}
                </div>
              </article>
            ))}
            {rounds.length === 0 ? (
              <article className="list-card">
                <p className="muted-copy">{t("review.noRounds")}</p>
                <div className="action-row" style={{ marginTop: 8 }}>
                  <button
                    className="primary-button"
                    disabled={busyAction !== ""}
                    onClick={handleStartReview}
                  >
                    {t("review.startReview")}
                  </button>
                </div>
              </article>
            ) : null}
          </div>
        </section>
      </section>
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
