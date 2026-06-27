import { Link } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { Bell, ChevronDown, MoreHorizontal, Plus, Search } from "lucide-react";
import { useTabs } from "@/context/tabs";
import { Breadcrumb } from "./breadcrumb";
import styles from "./index.module.css";

export type HeaderPage = "issues" | "dashboard" | "settings";

type HeaderProps = {
  activePage: HeaderPage;
  projectName: string | null;
  issueCount: number | null;
  onAddTask: () => void;
  showViewNavigation?: boolean;
  showAddTaskButton?: boolean;
};

export function Header({
  activePage,
  onAddTask,
  projectName,
  showViewNavigation = true,
  showAddTaskButton = true,
}: HeaderProps) {
  const { t } = useTranslation();
  const tabs = useTabs();

  return (
    <header className={styles.header}>
      <div className={styles.utilityRow}>
        <Breadcrumb />

        <div className={styles.globalActions}>
          <label className={styles.search}>
            <Search aria-hidden="true" size={16} strokeWidth={1.8} />
            <input type="search" placeholder={t("header.searchPlaceholder")} />
            <kbd>{t("header.commandKey")}</kbd>
          </label>
          <button
            type="button"
            className={styles.notificationButton}
            aria-label={t("header.notifications")}
          >
            <Bell aria-hidden="true" size={18} strokeWidth={1.8} />
          </button>
        </div>
      </div>

      <div className={styles.titleRow}>
        <div className={styles.titleGroup}>
          <h1>
            {activePage === "issues"
              ? (projectName ?? t("header.issueList"))
              : t(pageHeadingKey(activePage))}
          </h1>
          {activePage === "issues" ? (
            <button
              type="button"
              className={styles.moreButton}
              aria-label={t("header.moreProjectActions")}
            >
              <MoreHorizontal aria-hidden="true" size={19} strokeWidth={1.8} />
            </button>
          ) : null}
        </div>

        {showAddTaskButton ? (
          <div className={styles.createActions}>
            <button type="button" className={styles.createButton} onClick={onAddTask}>
              <Plus aria-hidden="true" size={17} strokeWidth={1.8} />
              {t("header.addTask")}
            </button>
            <button
              type="button"
              className={styles.createSplitButton}
              aria-label={t("header.moreProjectActions")}
              onClick={onAddTask}
            >
              <ChevronDown aria-hidden="true" size={16} strokeWidth={1.8} />
            </button>
          </div>
        ) : null}
      </div>

      {showViewNavigation && tabs.pages.length > 0 ? (
        <div className={styles.viewRow}>
          <nav className={styles.tabs} aria-label={t("header.views")}>
            {tabs.pages.map((page) => (
              <Link
                key={page.key}
                className={
                  tabs.activeKey === page.key
                    ? `${styles.tab} ${styles.active}`
                    : styles.tab
                }
                to={page.href}
              >
                {t(page.titleKey)}
              </Link>
            ))}
          </nav>
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
