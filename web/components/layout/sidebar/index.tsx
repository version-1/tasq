import Link from "next/link";
import { useTranslation } from "react-i18next";
import styles from "./index.module.css";

const projectItems = [
  { key: "all", labelKey: "sidebar.allProjects", state: "hollow" },
  { key: "product", labelKey: "sidebar.productWebsite", state: "active" },
  { key: "mobile", labelKey: "sidebar.mobileApp", state: "muted" },
  { key: "api", labelKey: "sidebar.apiBackend", state: "muted" },
  { key: "marketing", labelKey: "sidebar.marketing", state: "muted" },
] as const;

export function Sidebar() {
  const { t } = useTranslation();

  return (
    <aside className={styles.sidebar}>
      <Link className={styles.brand} href="/issues" aria-label={t("sidebar.home")}>
        tasq
      </Link>

      <section className={styles.projects} aria-labelledby="sidebar-projects-heading">
        <div className={styles.sectionHeader}>
          <h2 id="sidebar-projects-heading">{t("sidebar.projects")}</h2>
          <button type="button" aria-label={t("sidebar.addProject")}>＋</button>
        </div>
        <div className={styles.projectList}>
          {projectItems.map((item) => (
            <button
              key={item.key}
              className={item.state === "active" ? `${styles.projectItem} ${styles.active}` : styles.projectItem}
              type="button"
            >
              <span className={`${styles.projectDot} ${styles[item.state]}`} aria-hidden="true" />
              {t(item.labelKey)}
              {item.state === "active" ? <span className={styles.projectMore} aria-hidden="true">···</span> : null}
            </button>
          ))}
        </div>
      </section>

      <Link className={styles.settingsLink} href="/settings">
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
