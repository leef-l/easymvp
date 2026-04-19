import { Link } from "react-router-dom";
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
  const { projectId, routes } = useProjectState();
  const state = useQuery(
    () => apiGet<RepairDraftView>(`/api/v3/projects/${encodeURIComponent(projectId)}/repair-draft`),
    [projectId],
  );
  const repairDraft = state.data?.repair_draft;

  return (
    <QueryPanel loading={state.loading} error={state.error} title="Repair draft">
      {repairDraft ? (
        <section className="dashboard-page">
          <div className="dashboard-intro">
            <div>
              <p className="placeholder-section">Repair Draft</p>
              <h3 className="placeholder-title">{repairDraft.id || "No repair draft yet"}</h3>
              <p className="placeholder-description">{repairDraft.reasoning_summary || "The latest failed acceptance run has not produced a repair draft."}</p>
            </div>
            <div className="summary-stack">
              <span className="status-pill">{repairDraft.status}</span>
              <span className="status-pill">{repairDraft.updated_at || "no timestamp"}</span>
            </div>
          </div>

          <div className="content-grid repair-grid">
            <section className="data-panel">
              <div className="panel-header">
                <h3>Replaced Constraints</h3>
              </div>
              <div className="stack-list">
                {(repairDraft.replaced_constraints ?? []).map((item) => (
                  <article key={item} className="list-card">
                    <p>{item}</p>
                  </article>
                ))}
                {(repairDraft.replaced_constraints ?? []).length === 0 ? (
                  <article className="list-card">
                    <p>No replaced constraints recorded.</p>
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
              <p className="placeholder-section">Repair Draft</p>
              <h3 className="placeholder-title">No repair draft available</h3>
              <p className="placeholder-description">
                The latest acceptance flow has not produced a repair draft for this project.
              </p>
            </div>
            <div className="summary-stack">
              <Link className="secondary-button" to={routes.plan}>
                Open Plan
              </Link>
              <Link className="secondary-button" to={routes.workspace}>
                Back to Workspace
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
