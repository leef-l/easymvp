import { NavLink, Outlet } from "react-router-dom";
import { useProjectState } from "@/shared/lib/project";

export function AppShell() {
  const { projectId, updateProjectId, routes } = useProjectState();
  const navigation = [
    { to: routes.workspace, label: "Workspace" },
    { to: routes.plan, label: "Plan" },
    { to: routes.execution, label: "Execution" },
    { to: routes.replay, label: "Replay" },
    { to: routes.acceptance, label: "Acceptance" },
    { to: routes.audit, label: "Audit" },
    { to: routes.settings, label: "Settings" },
  ];

  return (
    <div className="app-shell">
      <aside className="app-sidebar">
        <div className="brand-block">
          <p className="brand-eyebrow">EasyMVP V3</p>
          <h1 className="brand-title">Local Workbench</h1>
          <p className="brand-copy">Single-user project cockpit for plan, execution, and production acceptance.</p>
        </div>
        <nav className="app-nav" aria-label="Primary">
          {navigation.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              className={({ isActive }) => (isActive ? "nav-link is-active" : "nav-link")}
            >
              {item.label}
            </NavLink>
          ))}
        </nav>
        <section className="project-switcher">
          <label className="project-label" htmlFor="project-id">
            Project ID
          </label>
          <input
            id="project-id"
            className="project-input"
            value={projectId}
            onChange={(event) => updateProjectId(event.target.value)}
            placeholder="project-demo"
          />
          <p className="project-help">All workbench pages query `/api/v3/projects/{projectId}/*` using this value.</p>
        </section>
      </aside>
      <main className="app-main">
        <header className="app-header">
          <div>
            <p className="header-kicker">Realtime Workbench</p>
            <h2 className="header-title">Desktop Shell Bootstrap</h2>
          </div>
          <div className="header-badge">React + Electron</div>
        </header>
        <Outlet />
      </main>
    </div>
  );
}
