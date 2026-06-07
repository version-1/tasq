import { Link } from "react-router-dom";
import { useTranslation } from "react-i18next";
import type { Project } from "@/lib/types";
import styles from "./index.module.css";

type SidebarProps = {
  activePage: "issues" | "dashboard" | "settings";
  activeProjectID: number | null;
  projects: Project[];
};

export function Sidebar({
  activePage,
  activeProjectID,
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
          <span className={styles.navIcon} aria-hidden="true">
            ▦
          </span>
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
          <span className={styles.navIcon} aria-hidden="true">
            □
          </span>
          {t("sidebar.board")}
        </Link>
      </nav>

      <section className={styles.projects} aria-label={t("sidebar.projects")}>
        <div className={styles.sectionHeader}>
          <span>{t("sidebar.projectsTitle")}</span>
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
                  <span className={styles.projectMore} aria-hidden="true">
                    ···
                  </span>
                ) : null}
              </Link>
            );
          })}
        </div>
      </section>

      <Link className={styles.settingsLink} to="/settings">
        <span aria-hidden="true">⚙</span>
        {t("sidebar.settings")}
      </Link>

      <div className={styles.account}>
        <div className={styles.avatar} aria-hidden="true">
          YK
        </div>
        <div>
          <strong>{t("sidebar.userName")}</strong>
          <span>{t("sidebar.userEmail")}</span>
        </div>
        <button type="button" aria-label={t("sidebar.accountMenu")}>
          ⌄
        </button>
      </div>
    </aside>
  );
}
