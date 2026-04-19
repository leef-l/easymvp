import { BrowserRouter, Navigate, Route, Routes } from "react-router-dom";
import { AppShell } from "./shell";
import { AcceptancePage } from "@/modules/acceptance/pages/AcceptancePage";
import { AuditPage } from "@/modules/audit/pages/AuditPage";
import { DiagnosticsPage } from "@/modules/diagnostics/pages/DiagnosticsPage";
import { ExecutionPage } from "@/modules/execution/pages/ExecutionPage";
import { PlanPage } from "@/modules/plan/pages/PlanPage";
import { RepairDraftPage } from "@/modules/plan/pages/RepairDraftPage";
import { SettingsPage } from "@/modules/settings/pages/SettingsPage";
import { WorkspacePage } from "@/modules/workspace/pages/WorkspacePage";

export function AppRouter() {
  return (
    <BrowserRouter>
      <Routes>
        <Route element={<AppShell />}>
          <Route path="/" element={<Navigate to="/workspace" replace />} />
          <Route path="/workspace" element={<WorkspacePage />} />
          <Route path="/plan" element={<PlanPage />} />
          <Route path="/repair-draft" element={<RepairDraftPage />} />
          <Route path="/execution" element={<ExecutionPage />} />
          <Route path="/replay" element={<ExecutionPage />} />
          <Route path="/acceptance" element={<AcceptancePage />} />
          <Route path="/audit" element={<AuditPage />} />
          <Route path="/diagnostics" element={<DiagnosticsPage />} />
          <Route path="/settings" element={<SettingsPage />} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}
