import { Link } from "react-router-dom";
import { useTranslation } from "react-i18next";
import type { Project } from "@/lib/types";
import styles from "./index.module.css";

type SidebarProps = {
  activeProjectID: number | null;
  projects: Project[];
};

export function Sidebar({ activeProjectID, projects }: SidebarProps) {
  const { t } = useTranslation();

  return (
    <aside className={styles.sidebar}>
      <Link className={styles.brand} to="/issues" aria-label={t("sidebar.home")}>
        tasq
      </Link>

      <section className={styles.projects} aria-labelledby="sidebar-projects-heading">
        <div className={styles.sectionHeader}>
          <h2 id="sidebar-projects-heading">{t("sidebar.projects")}</h2>
          <button type="button" aria-label={t("sidebar.addProject")}>＋</button>
        </div>
        <div className={styles.projectList}>
          <Link
            className={activeProjectID === null ? `${styles.projectItem} ${styles.active}` : styles.projectItem}
            to="/issues"
          >
            <span
              className={`${styles.projectDot} ${activeProjectID === null ? styles.active : styles.hollow}`}
              aria-hidden="true"
            />
            {t("sidebar.allProjects")}
            {activeProjectID === null ? <span className={styles.projectMore} aria-hidden="true">···</span> : null}
          </Link>
          {projects.map((project) => {
            const isActive = project.id === activeProjectID;
            return (
              <Link
                key={project.id}
                className={isActive ? `${styles.projectItem} ${styles.active}` : styles.projectItem}
                to={`/projects/${encodeURIComponent(project.key)}/issues`}
              >
                <span
                  className={`${styles.projectDot} ${isActive ? styles.active : styles.muted}`}
                  aria-hidden="true"
                />
                {project.name}
                {isActive ? <span className={styles.projectMore} aria-hidden="true">···</span> : null}
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
        <div className={styles.avatar} aria-hidden="true">YK</div>
        <div>
          <strong>{t("sidebar.userName")}</strong>
          <span>{t("sidebar.userEmail")}</span>
        </div>
        <button type="button" aria-label={t("sidebar.accountMenu")}>⌄</button>
      </div>
    </aside>
  );
}
