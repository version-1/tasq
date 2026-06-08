import { Link } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { IconProxy } from "@/components/ui/icon-proxy";
import { Switch } from "@/components/ui/switch";
import type { Project } from "@/lib/types";
import styles from "./index.module.css";

type SidebarProps = {
  activePage: "issues" | "dashboard" | "settings";
  activeProjectID: number | null;
  projects: Project[];
  onAddProject: () => void;
};

export function Sidebar({
  activePage,
  activeProjectID,
  onAddProject,
  projects,
}: SidebarProps) {
  const { t } = useTranslation();

  return (
    <aside className={styles.sidebar}>
      <Link
        className={styles.brand}
        to="/issues"
        aria-label={t("sidebar.home")}
      >
        tasq
      </Link>

      <nav
        className={styles.primaryNavigation}
        aria-label={t("sidebar.primaryNavigation")}
      >
        <Link
          className={
            activePage === "dashboard"
              ? `${styles.navItem} ${styles.active}`
              : styles.navItem
          }
          to="/dashboard"
        >
          <IconProxy name="layout-dashboard" className={styles.navIcon} />
          {t("header.dashboard")}
        </Link>
        <Link
          className={
            activePage === "issues" && activeProjectID === null
              ? `${styles.navItem} ${styles.active}`
              : styles.navItem
          }
          to="/issues"
        >
          <IconProxy name="square-kanban" className={styles.navIcon} />
          {t("sidebar.board")}
        </Link>
      </nav>

      <section className={styles.projects} aria-label={t("sidebar.projects")}>
        <div className={styles.sectionHeader}>
          <span>{t("sidebar.projectsTitle")}</span>
          <button
            className={styles.addProjectButton}
            type="button"
            aria-label={t("sidebar.addProject")}
            title={t("sidebar.addProject")}
            onClick={onAddProject}
          >
            <IconProxy name="plus" size={15} />
          </button>
        </div>
        <div className={styles.projectList}>
          {projects.map((project) => {
            const isActive = project.id === activeProjectID;
            return (
              <Link
                key={project.id}
                className={
                  isActive
                    ? `${styles.projectItem} ${styles.active}`
                    : styles.projectItem
                }
                to={`/projects/${encodeURIComponent(project.key)}/issues`}
              >
                {project.name}
                {isActive ? (
                  <IconProxy name="ellipsis" className={styles.projectMore} />
                ) : null}
              </Link>
            );
          })}
        </div>
      </section>

      <Link className={styles.settingsLink} to="/settings">
        <IconProxy name="settings" className={styles.navIcon} />
        {t("sidebar.settings")}
      </Link>

      <label className={styles.themeSwitch}>
        <span>{t("sidebar.darkMode")}</span>
        <Switch aria-label={t("sidebar.darkMode")} />
      </label>
    </aside>
  );
}
