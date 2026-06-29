import { Link } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { MoreHorizontal, Plus, Search } from "lucide-react";
import { useTabs } from "@/context/tabs";
import { supportedLanguages, type SupportedLanguage } from "@/lib/i18n";
import { Breadcrumb } from "./breadcrumb";
import styles from "./index.module.css";

export type HeaderPage = "issues" | "dashboard" | "settings";

type HeaderProps = {
  activePage: HeaderPage;
  projectName: string | null;
  issueCount: number | null;
  isIssueDetailPage?: boolean;
  language: SupportedLanguage;
  searchQuery: string;
  onLanguageChange: (language: SupportedLanguage) => void;
  onSearchQueryChange: (query: string) => void;
  onAddTask: () => void;
  showViewNavigation?: boolean;
  showAddTaskButton?: boolean;
};

export function Header({
  activePage,
  isIssueDetailPage = false,
  language,
  onLanguageChange,
  onSearchQueryChange,
  onAddTask,
  projectName,
  searchQuery,
  showViewNavigation = true,
  showAddTaskButton = true,
}: HeaderProps) {
  const { t } = useTranslation();
  const tabs = useTabs();
  const title = isIssueDetailPage
    ? (projectName ?? t("issues.detailPage.detailTab"))
    : activePage === "issues"
      ? (projectName ?? t("header.issueList"))
      : t(pageHeadingKey(activePage));

  return (
    <header className={styles.header}>
      <div className={styles.utilityRow}>
        <Breadcrumb />

        <div className={styles.globalActions}>
          <label className={styles.search}>
            <Search aria-hidden="true" size={16} strokeWidth={1.8} />
            <input
              type="search"
              placeholder={t("header.searchPlaceholder")}
              value={searchQuery}
              onChange={(event) => onSearchQueryChange(event.target.value)}
            />
            <kbd>{t("header.commandKey")}</kbd>
          </label>
          <label className={styles.languageSelector}>
            <select
              aria-label={t("header.language")}
              value={language}
              onChange={(event) => onLanguageChange(event.target.value as SupportedLanguage)}
            >
              {supportedLanguages.map((item) => (
                <option key={item} value={item}>
                  {t(`header.languages.${item}`)}
                </option>
              ))}
            </select>
          </label>
        </div>
      </div>

      <div className={styles.titleRow}>
        <div className={styles.titleGroup}>
          <h1>{title}</h1>
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
