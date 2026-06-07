import { Link } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { Button } from "@/components/ui/button";
import { Breadcrumb } from "./breadcrumb";
import styles from "./index.module.css";

type HeaderPage = "issues" | "agents" | "dashboard" | "settings";

type HeaderProps = {
  activePage: HeaderPage;
  projectName: string | null;
  issueCount: number | null;
  onAddTask: () => void;
  showViewNavigation?: boolean;
};

const pages = [
  { key: "issues", href: "/issues", titleKey: "header.board" },
  { key: "agents", href: "/agents", titleKey: "header.agents" },
  { key: "settings", href: "/settings", titleKey: "header.settings" },
] as const;

export function Header({
  activePage,
  projectName,
  issueCount,
  onAddTask,
  showViewNavigation = true,
}: HeaderProps) {
  const { t } = useTranslation();
  const displayedProjectName = projectName ?? t("header.projectName");

  return (
    <header className={styles.header}>
      <div className={styles.utilityRow}>
        <Breadcrumb projectName={displayedProjectName} />

        <div className={styles.globalActions}>
          <label className={styles.search}>
            <span aria-hidden="true">⌕</span>
            <input type="search" placeholder={t("header.searchPlaceholder")} />
            <kbd>{t("header.commandKey")}</kbd>
          </label>
          <button
            className={styles.notificationButton}
            type="button"
            aria-label={t("header.notifications")}
          >
            ♡
          </button>
        </div>
      </div>

      <div className={styles.titleRow}>
        <div className={styles.titleGroup}>
          <h1>
            {activePage === "issues"
              ? displayedProjectName
              : t(pageHeadingKey(activePage))}
          </h1>
          <button
            className={styles.moreButton}
            type="button"
            aria-label={t("header.moreProjectActions")}
          >
            ···
          </button>
        </div>

        <div className={styles.statusStrip}>
          <span>
            {issueCount === null
              ? t("header.loading")
              : t("header.issueCount", { count: issueCount })}
          </span>
          <Button onClick={onAddTask}>
            <span aria-hidden="true">＋</span>
            {t("header.addTask")}
          </Button>
        </div>
      </div>

      {showViewNavigation ? (
        <div className={styles.viewRow}>
          <nav className={styles.tabs} aria-label={t("header.views")}>
            {pages.map((page) => (
              <Link
                key={page.key}
                className={
                  activePage === page.key
                    ? `${styles.tab} ${styles.active}`
                    : styles.tab
                }
                to={page.href}
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
      ) : null}
    </header>
  );
}

function pageHeadingKey(page: Exclude<HeaderPage, "issues">): string {
  switch (page) {
    case "agents":
      return "header.agentRuns";
    case "dashboard":
      return "header.dashboard";
    case "settings":
      return "header.settings";
  }
}
