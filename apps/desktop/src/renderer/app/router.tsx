import {
  BrowserRouter,
  HashRouter,
  Navigate,
  Route,
  Routes,
} from "react-router-dom";
import { useTranslation } from "react-i18next";
import { AppShell } from "./shell";
import { AcceptancePage } from "@/modules/acceptance/pages/AcceptancePage";
import { AuditPage } from "@/modules/audit/pages/AuditPage";
import { DesignPage } from "@/modules/design/pages/DesignPage";
import { DiagnosticsPage } from "@/modules/diagnostics/pages/DiagnosticsPage";
import { RecoveryPage } from "@/modules/diagnostics/pages/RecoveryPage";
import { ExecutionPage } from "@/modules/execution/pages/ExecutionPage";
import { PlanPage } from "@/modules/plan/pages/PlanPage";
import { ArchitectChatPage } from "@/modules/architect_chat/pages/ArchitectChatPage";
import { RepairDraftPage } from "@/modules/plan/pages/RepairDraftPage";
import { ProjectsPage } from "@/modules/projects/pages/ProjectsPage";
import { RequirementsPage } from "@/modules/requirements/pages/RequirementsPage";
import { ReviewPage } from "@/modules/review/pages/ReviewPage";
import { DeliveryPage } from "@/modules/delivery/pages/DeliveryPage";
import { RetrospectivePage } from "@/modules/retrospective/pages/RetrospectivePage";
import { SettingsPage } from "@/modules/settings/pages/SettingsPage";
import { WorkspacePage } from "@/modules/workspace/pages/WorkspacePage";
import { getDesktopRuntimeInfo } from "@/shared/lib/preferences";
import { useQuery } from "@/shared/lib/query";

function StartupRedirect() {
  const { t } = useTranslation();
  const runtimeState = useQuery(() => getDesktopRuntimeInfo(), []);

  if (runtimeState.loading) {
    return (
      <section className="placeholder-page">
        <div className="placeholder-hero">
          <div>
            <p className="placeholder-section">{t("startup.section")}</p>
            <h3 className="placeholder-title">{t("startup.title")}</h3>
            <p className="placeholder-description">
              {t("startup.description")}
            </p>
          </div>
          <span className="status-pill">{t("status.probing")}</span>
        </div>
      </section>
    );
  }

  if (
    runtimeState.error ||
    runtimeState.data?.issue ||
    !runtimeState.data?.coreReachable
  ) {
    return <Navigate to="/recovery" replace />;
  }

  return <Navigate to="/projects" replace />;
}

export function AppRouter() {
  const Router =
    typeof window !== "undefined" && window.location.protocol === "file:"
      ? HashRouter
      : BrowserRouter;

  return (
    <Router>
      <Routes>
        <Route element={<AppShell />}>
          <Route path="/" element={<StartupRedirect />} />
          <Route path="/projects" element={<ProjectsPage />} />
          <Route path="/workspace" element={<WorkspacePage />} />
          <Route path="/requirements" element={<RequirementsPage />} />
          <Route path="/design" element={<DesignPage />} />
          <Route path="/architect-chat" element={<ArchitectChatPage />} />
          <Route path="/plan" element={<PlanPage />} />
          <Route path="/repair-draft" element={<RepairDraftPage />} />
          <Route path="/execution" element={<ExecutionPage />} />
          <Route path="/review" element={<ReviewPage />} />
          <Route path="/acceptance" element={<AcceptancePage />} />
          <Route path="/delivery" element={<DeliveryPage />} />
          <Route path="/retrospective" element={<RetrospectivePage />} />
          <Route path="/audit" element={<AuditPage />} />
          <Route path="/diagnostics" element={<DiagnosticsPage />} />
          <Route path="/recovery" element={<RecoveryPage />} />
          <Route path="/settings" element={<SettingsPage />} />
        </Route>
      </Routes>
    </Router>
  );
}
