import Link from "next/link";
import { useTranslation } from "react-i18next";
import type { TasqPage } from "@/components/layout";
import styles from "./index.module.css";

type SidebarProps = {
  activePage: TasqPage;
};

const primaryItems = [
  { key: "issues", href: "/issues", labelKey: "sidebar.board", icon: "▥" },
  { key: "myTasks", href: "/issues", labelKey: "sidebar.myTasks", icon: "◉" },
  { key: "calendar", href: "/issues", labelKey: "sidebar.calendar", icon: "□" },
  { key: "reports", href: "/agents", labelKey: "sidebar.reports", icon: "▥" },
  { key: "workflow", href: "/agents", labelKey: "sidebar.workflow", icon: "▱" },
  { key: "agents", href: "/agents", labelKey: "sidebar.members", icon: "♙" },
  { key: "settings", href: "/settings", labelKey: "sidebar.settings", icon: "⚙" },
] as const;

const activePages = ["issues", "agents", "settings"] as const;

const projectItems = [
  { key: "all", labelKey: "sidebar.allProjects", state: "hollow" },
  { key: "product", labelKey: "sidebar.productWebsite", state: "active" },
  { key: "mobile", labelKey: "sidebar.mobileApp", state: "muted" },
  { key: "api", labelKey: "sidebar.apiBackend", state: "muted" },
  { key: "marketing", labelKey: "sidebar.marketing", state: "muted" },
] as const;

export function Sidebar({ activePage }: SidebarProps) {
  const { t } = useTranslation();

  return (
    <aside className={styles.sidebar}>
      <Link className={styles.brand} href="/issues" aria-label={t("sidebar.home")}>
        tasq
      </Link>

      <nav className={styles.primaryNav} aria-label={t("sidebar.primaryNavigation")}>
        {primaryItems.map((item) => (
          <Link
            key={item.key}
            className={isActivePageKey(item.key) && activePage === item.key ? `${styles.navItem} ${styles.active}` : styles.navItem}
            href={item.href}
          >
            <span className={styles.navIcon} aria-hidden="true">{item.icon}</span>
            {t(item.labelKey)}
          </Link>
        ))}
      </nav>

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

      <Link className={styles.archive} href="/issues">
        <span aria-hidden="true">▱</span>
        {t("sidebar.archive")}
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

function isActivePageKey(key: string): key is TasqPage {
  return activePages.includes(key as TasqPage);
}
