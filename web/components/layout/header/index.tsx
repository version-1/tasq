import Link from "next/link";
import { useTranslation } from "react-i18next";
import styles from "./index.module.css";

type HeaderPage = "issues" | "agents" | "workspace" | "settings";

type HeaderProps = {
  activePage: HeaderPage;
  projectName: string | null;
  issueCount: number | null;
  runCount: number | null;
  onRefresh: () => void;
};

const pages = [
  { key: "issues", href: "/issues", titleKey: "header.board" },
  { key: "agents", href: "/agents", titleKey: "header.agents" },
  { key: "workspace", href: "/workspace", titleKey: "header.workspace" },
  { key: "settings", href: "/settings", titleKey: "header.settings" },
] as const;

export function Header({
  activePage,
  projectName,
  issueCount,
  runCount,
  onRefresh,
}: HeaderProps) {
  const { t } = useTranslation();
  const displayedProjectName = projectName ?? t("header.projectName");

  return (
    <header className={styles.header}>
      <div className={styles.utilityRow}>
        <div className={styles.breadcrumb} aria-label={t("header.breadcrumb")}>
          <span>{t("header.project")}</span>
          <span aria-hidden="true">/</span>
          <button className={styles.projectSwitch} type="button">
            {displayedProjectName}
            <span aria-hidden="true">⌄</span>
          </button>
        </div>

        <div className={styles.globalActions}>
          <label className={styles.search}>
            <span aria-hidden="true">⌕</span>
            <input type="search" placeholder={t("header.searchPlaceholder")} />
            <kbd>{t("header.commandKey")}</kbd>
          </label>
          <button className={styles.notificationButton} type="button" aria-label={t("header.notifications")}>
            ♡
          </button>
        </div>
      </div>

      <div className={styles.titleRow}>
        <div className={styles.titleGroup}>
          <h1>{activePage === "issues" ? displayedProjectName : t(pageHeadingKey(activePage))}</h1>
          <button className={styles.moreButton} type="button" aria-label={t("header.moreProjectActions")}>
            ···
          </button>
        </div>

        <div className={styles.statusStrip}>
          <span>{issueCount === null ? t("header.loading") : t("header.issueCount", { count: issueCount })}</span>
          <span>{runCount === null ? t("header.trackedRunLoading") : t("header.trackedRunCount", { count: runCount })}</span>
          <button className={styles.primaryButton} type="button" onClick={onRefresh}>
            <span aria-hidden="true">＋</span>
            {t("header.addTask")}
          </button>
          <button className={styles.splitButton} type="button" aria-label={t("header.taskOptions")}>
            ⌄
          </button>
        </div>
      </div>

      <div className={styles.viewRow}>
        <nav className={styles.tabs} aria-label={t("header.views")}>
          {pages.map((page) => (
            <Link
              key={page.key}
              className={activePage === page.key ? `${styles.tab} ${styles.active}` : styles.tab}
              href={page.href}
            >
              {t(page.titleKey)}
            </Link>
          ))}
        </nav>

        <div className={styles.viewActions}>
          <button type="button">{t("header.filter")}</button>
          <button type="button">{t("header.sort")}</button>
          <button type="button">{t("header.view")}</button>
        </div>
      </div>
    </header>
  );
}

function pageHeadingKey(page: Exclude<HeaderPage, "issues">): string {
  switch (page) {
    case "agents":
      return "header.agentRuns";
    case "workspace":
      return "header.workspace";
    case "settings":
      return "header.settings";
  }
}
