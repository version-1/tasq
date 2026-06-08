import { Link } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { Button } from "@/components/ui/button";
import { Breadcrumb } from "./breadcrumb";
import styles from "./index.module.css";

type HeaderPage = "issues" | "dashboard" | "settings";

type HeaderProps = {
  activePage: HeaderPage;
  projectName: string | null;
  issueCount: number | null;
  onAddTask: () => void;
  showViewNavigation?: boolean;
  showAddTaskButton?: boolean;
};

const pages = [
  { key: "issues", href: "/issues", titleKey: "header.board" },
  { key: "settings", href: "/settings", titleKey: "header.settings" },
] as const;

export function Header({
  activePage,
  projectName,
  onAddTask,
  showViewNavigation = true,
  showAddTaskButton = true,
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
        </div>
      </div>

      <div className={styles.titleRow}>
        <div className={styles.titleGroup}>
          <h1>
            {activePage === "issues"
              ? displayedProjectName
              : t(pageHeadingKey(activePage))}
          </h1>
        </div>

        {showAddTaskButton ? (
          <div className={styles.statusStrip}>
            <Button onClick={onAddTask}>
              <span aria-hidden="true">＋</span>
              {t("header.addTask")}
            </Button>
          </div>
        ) : null}
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
    case "dashboard":
      return "header.dashboard";
    case "settings":
      return "header.settings";
  }
}
