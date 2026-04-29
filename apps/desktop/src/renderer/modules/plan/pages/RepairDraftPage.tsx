import { useTranslation } from "react-i18next";
import { Link, useNavigate } from "react-router-dom";
import { apiGet } from "@/shared/lib/api";
import { useProjectState } from "@/shared/lib/project";
import { useQuery } from "@/shared/lib/query";
import type { RepairDraftView } from "@/shared/lib/types";
import { QueryPanel } from "@/shared/ui/QueryPanel";

const jsonSections = [
  ["Failed Task Context", "failed_task_context_json"],
  ["Failure Reason", "failure_reason_json"],
  ["Original Contracts", "original_contracts_json"],
  ["Runtime Summary", "runtime_summary_json"],
  ["Repair Plan", "repair_plan_json"],
] as const;

export function RepairDraftPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { projectId, routes, buildRoute } = useProjectState();

  if (!projectId) {
    return (
      <section className="placeholder-page">
        <div className="empty-state-panel">
          <h4>{t("repair.noProjectTitle")}</h4>
          <p>{t("repair.noProjectDescription")}</p>
          <div className="action-row">
            <button className="primary-button" onClick={() => navigate("/projects")}>
              {t("repair.goToProjects")}
            </button>
            <button className="secondary-button" onClick={() => navigate("/settings")}>
              {t("settings.createProject")}
            </button>
          </div>
        </div>
      </section>
    );
  }
  const state = useQuery(
    () => apiGet<RepairDraftView>(`/api/v3/projects/${encodeURIComponent(projectId)}/repair-draft`),
    [projectId],
  );
  const repairDraft = state.data?.repair_draft;
  const failedTaskContext = safeParseJSONObject(repairDraft?.failed_task_context_json);
  const failureReason = safeParseJSONObject(repairDraft?.failure_reason_json);
  const runtimeSummary = safeParseJSONObject(repairDraft?.runtime_summary_json);
  const taskID = readStringField(failedTaskContext, failureReason, runtimeSummary, "acceptance_task_id", "task_id", "source_task_id");
  const runID = readStringField(runtimeSummary, failureReason, failedTaskContext, "run_id", "source_run_id");
  const bindingID = readStringField(runtimeSummary, failureReason, failedTaskContext, "binding_id", "run_binding_id");
  const replayID = readStringField(runtimeSummary, failureReason, "replay_id", "latest_replay_id");

  return (
    <QueryPanel loading={state.loading} error={state.error} title={t("repair.title")}>
      {repairDraft ? (
        <section className="dashboard-page">
          <div className="dashboard-intro">
            <div>
              <p className="placeholder-section">{t("repair.title")}</p>
              <h3 className="placeholder-title">{repairDraft.id || t("repair.noRepairDraftYet")}</h3>
              <p className="placeholder-description">{repairDraft.reasoning_summary || t("repair.noRepairDraftDescription")}</p>
            </div>
            <div className="summary-stack">
              <span className="status-pill">{repairDraft.status}</span>
              <span className="status-pill">{repairDraft.updated_at || "no timestamp"}</span>
              {taskID ? <span className="status-pill">task {taskID}</span> : null}
              {runID ? <span className="status-pill">run {runID}</span> : null}
            </div>
          </div>

          <div className="content-grid repair-grid">
            <section className="data-panel">
              <div className="panel-header">
                <h3>{t("repair.linkedActions")}</h3>
              </div>
              <div className="action-row">
                <Link className="secondary-button" to={routes.plan}>
                  {t("repair.openPlan")}
                </Link>
                <Link className="secondary-button" to={routes.diagnostics}>
                  {t("repair.openDiagnostics")}
                </Link>
                {taskID ? (
                  <>
                    <Link className="secondary-button" to={buildRoute("/acceptance", { task: taskID })}>
                      {t("repair.openAcceptance")}
                    </Link>
                    <Link
                      className="secondary-button"
                      to={buildRoute("/execution", {
                        binding: bindingID || undefined,
                        run: runID || undefined,
                        task: taskID,
                      })}
                    >
                      {t("repair.openExecution")}
                    </Link>
                  </>
                ) : null}
                <Link
                  className="secondary-button"
                  to={buildRoute("/replay", {
                    binding: bindingID || undefined,
                    run: runID || undefined,
                    replay: replayID || undefined,
                    task: taskID || undefined,
                  })}
                >
                  {t("repair.openReplay")}
                </Link>
                <Link
                  className="secondary-button"
                  to={buildRoute("/audit", {
                    run: runID || undefined,
                    task: taskID || undefined,
                  })}
                >
                  {t("repair.openAudit")}
                </Link>
              </div>
            </section>

            <section className="data-panel">
              <div className="panel-header">
                <h3>{t("repair.repairContext")}</h3>
                <span className="status-pill">{runID || taskID ? t("repair.linked") : t("repair.unlinked")}</span>
              </div>
              <div className="stack-list">
                <article className="list-card">
                  <div className="list-card-head">
                    <strong>{taskID || t("repair.noTaskContext")}</strong>
                    <span className="status-pill">{runID || "no run"}</span>
                  </div>
                  <p>
                    binding {bindingID || "n/a"} · replay {replayID || "n/a"}
                  </p>
                  <p>
                    {t("repair.contextDescription")}
                  </p>
                </article>
              </div>
            </section>

            <section className="data-panel">
              <div className="panel-header">
                <h3>{t("repair.replacedConstraints")}</h3>
              </div>
              <div className="stack-list">
                {(repairDraft.replaced_constraints ?? []).map((item) => (
                  <article key={item} className="list-card">
                    <p>{item}</p>
                  </article>
                ))}
                {(repairDraft.replaced_constraints ?? []).length === 0 ? (
                  <article className="list-card">
                    <p>{t("repair.noReplacedConstraints")}</p>
                  </article>
                ) : null}
              </div>
            </section>

            {jsonSections.map(([label, key]) => (
              <section key={key} className="data-panel">
                <div className="panel-header">
                  <h3>{label}</h3>
                </div>
                <pre className="json-block">{prettyJson(repairDraft[key] as string | undefined)}</pre>
              </section>
            ))}
          </div>
        </section>
      ) : (
        <section className="dashboard-page">
          <div className="dashboard-intro">
            <div>
              <p className="placeholder-section">{t("repair.title")}</p>
              <h3 className="placeholder-title">{t("repair.noRepairDraftAvailable")}</h3>
              <p className="placeholder-description">
                {t("repair.noRepairDraftAvailableDescription")}
              </p>
            </div>
            <div className="summary-stack">
              <Link className="secondary-button" to={routes.plan}>
                {t("repair.openPlan")}
              </Link>
              <Link className="secondary-button" to={routes.workspace}>
                {t("repair.backToWorkspace")}
              </Link>
            </div>
          </div>
        </section>
      )}
    </QueryPanel>
  );
}

function prettyJson(raw?: string) {
  if (!raw || raw.trim() === "") {
    return "{}";
  }
  try {
    return JSON.stringify(JSON.parse(raw), null, 2);
  } catch {
    return raw;
  }
}

function safeParseJSONObject(raw?: string): Record<string, unknown> | null {
  if (!raw || raw.trim() === "") {
    return null;
  }
  try {
    const parsed = JSON.parse(raw);
    if (parsed && typeof parsed === "object" && !Array.isArray(parsed)) {
      return parsed as Record<string, unknown>;
    }
  } catch {
    return null;
  }
  return null;
}

function readStringField(...args: Array<Record<string, unknown> | null | string>) {
  const keys = args.filter((item): item is string => typeof item === "string");
  const sources = args.filter(
    (item): item is Record<string, unknown> => Boolean(item) && typeof item === "object",
  );
  for (const source of sources) {
    for (const key of keys) {
      const value = source[key];
      if (typeof value === "string" && value.trim() !== "") {
        return value.trim();
      }
    }
  }
  return "";
}
