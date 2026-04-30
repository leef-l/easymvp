import { NavLink, Outlet, useLocation, useNavigate } from "react-router-dom";

import { useTranslation } from "react-i18next";

import { useMemo } from "react";

import { useProjectState } from "@/shared/lib/project";

import { LanguageSwitcher } from "@/shared/ui/LanguageSwitcher";



export function AppShell() {

  const { t } = useTranslation();

  const { projectId, updateProjectId, routes, buildRoute } = useProjectState();

  const location = useLocation();

  const navigate = useNavigate();



  const hasProject = !!projectId;



  const globalNavigation = useMemo(

    () => [

      { to: "/projects", label: t("nav.projects"), icon: "?" },

      { to: "/settings", label: t("nav.settings"), icon: "?" },

    ],

    [t],

  );



  const projectNavigation = useMemo(

    () => [

      { to: routes.workspace, label: t("nav.workspace"), icon: "?" },

      { to: routes.requirements, label: t("nav.requirements"), icon: "?" },

      { to: routes.design, label: t("nav.design"), icon: "?" },

      { to: routes.architectChat, label: t("nav.architectChat"), icon: "?" },

      { to: routes.plan, label: t("nav.plan"), icon: "?" },

      { to: routes.execution, label: t("nav.execution"), icon: "?" },

      { to: routes.review, label: t("nav.review"), icon: "?" },

      { to: routes.acceptance, label: t("nav.acceptance"), icon: "✅" },

      { to: routes.delivery, label: t("nav.delivery"), icon: "?" },

      { to: routes.retrospective, label: t("nav.retrospective"), icon: "?" },

      { to: routes.audit, label: t("nav.audit"), icon: "?" },

      { to: routes.diagnostics, label: t("nav.diagnostics"), icon: "?" },

    ],

    [routes, t],

  );



  const currentPath = location.pathname;

  const isProjectPage =

    hasProject &&

    currentPath !== "/projects" &&

    currentPath !== "/settings" &&

    currentPath !== "/recovery";



  return (

    <div className="app-shell">

      <aside className="app-sidebar">

        <div className="brand-block">

          <p className="brand-eyebrow">{t("brand.eyebrow")}</p>

          <h1 className="brand-title">{t("brand.title")}</h1>

          <p className="brand-copy">{t("brand.copy")}</p>

        </div>



        <nav className="app-nav" aria-label="Primary">

          {globalNavigation.map((item) => (

            <NavLink

              key={item.to}

              to={item.to}

              className={({ isActive }) =>

                isActive ? "nav-link is-active" : "nav-link"

              }

            >

              {item.label}

            </NavLink>

          ))}

        </nav>



        {isProjectPage ? (

          <>

            <div className="nav-section-label">{t("nav.projectContext")}</div>

            <nav className="app-nav" aria-label="Project">

              {projectNavigation.map((item) => (

                <NavLink

                  key={item.to}

                  to={item.to}

                  className={({ isActive }) =>

                    isActive ? "nav-link is-active" : "nav-link"

                  }

                >

                  {item.label}

                </NavLink>

              ))}

            </nav>



            <section className="project-switcher">

              <label className="project-label" htmlFor="project-id">

                {t("project.currentLabel")}

              </label>

              <input

                id="project-id"

                className="project-input"

                value={projectId}

                readOnly

                title={projectId}

              />

              <div className="action-row" style={{ marginTop: 10 }}>

                <button

                  className="secondary-button"

                  style={{ flex: 1, fontSize: 12, padding: "8px 10px" }}

                  onClick={() => {

                    updateProjectId("");

                    navigate("/projects");

                  }}

                >

                  {t("project.backToList")}

                </button>

              </div>

              <p className="project-help">{t("project.help")}</p>

            </section>

          </>

        ) : null}



        <LanguageSwitcher />

      </aside>

      <main className="app-main">

        <header className="app-header">

          <div>

            <p className="header-kicker">{t("header.kicker")}</p>

            <h2 className="header-title">{t("header.title")}</h2>

          </div>

          <div className="header-badge">{t("header.badge")}</div>

        </header>



        {isProjectPage ? (

          <div className="project-context-bar">

            <div className="project-context-info">

              <span className="project-context-name">

                {t("project.currentProject")}

              </span>

              <span className="status-pill">{projectId}</span>

            </div>

            <div className="action-row">

              <button

                className="secondary-button"

                onClick={() => {

                  updateProjectId("");

                  navigate("/projects");

                }}

              >

                {t("project.switchProject")}

              </button>

            </div>

          </div>

        ) : null}



        <div className="app-content">

          <Outlet />

        </div>

      </main>

    </div>

  );

}

