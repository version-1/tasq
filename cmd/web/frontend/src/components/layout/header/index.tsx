import { Link } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { useEffect, useState } from "react";
import { MoreHorizontal, Plus, Search } from "lucide-react";
import { ContextMenu, ContextMenuItem } from "@/components/ui/context-menu";
import { useTabs } from "@/context/tabs";
import { fetchIssues } from "@/lib/api";
import { supportedLanguages, type SupportedLanguage } from "@/lib/i18n";
import type { Issue } from "@/lib/types";
import { Breadcrumb } from "./breadcrumb";
import styles from "./index.module.css";

export type HeaderPage = "issues" | "dashboard" | "settings";

type HeaderProps = {
  activePage: HeaderPage;
  projectName: string | null;
  issueCount: number | null;
  isIssueDetailPage?: boolean;
  language: SupportedLanguage;
  onLanguageChange: (language: SupportedLanguage) => void;
  onAddTask: () => void;
  onDeleteProject?: () => void;
  canDeleteProject?: boolean;
  showViewNavigation?: boolean;
  showAddTaskButton?: boolean;
};

export function Header({
  activePage,
  isIssueDetailPage = false,
  language,
  onLanguageChange,
  onAddTask,
  onDeleteProject,
  projectName,
  canDeleteProject = false,
  showViewNavigation = true,
  showAddTaskButton = true,
}: HeaderProps) {
  const { t } = useTranslation();
  const tabs = useTabs();
  const [searchQuery, setSearchQuery] = useState("");
  const [searchResults, setSearchResults] = useState<Issue[]>([]);
  const [searchState, setSearchState] = useState<"idle" | "loading" | "ready" | "error">("idle");
  const [isSearchFocused, setIsSearchFocused] = useState(false);
  const [isProjectMenuOpen, setIsProjectMenuOpen] = useState(false);
  const title = isIssueDetailPage
    ? (projectName ?? t("issues.detailPage.detailTab"))
    : activePage === "issues"
      ? (projectName ?? t("header.issueList"))
      : t(pageHeadingKey(activePage));
  const trimmedSearchQuery = searchQuery.trim();
  const showSearchResults = isSearchFocused && trimmedSearchQuery !== "";
  const searchResultsID = "header-search-results";
  const projectMenuID = "header-project-actions";

  useEffect(() => {
    if (trimmedSearchQuery === "") {
      setSearchResults([]);
      setSearchState("idle");
      return;
    }

    let active = true;
    setSearchState("loading");
    const id = window.setTimeout(() => {
      void fetchIssues(
        {
          limit: 6,
          search: trimmedSearchQuery,
          sort_by: "updated_at",
          sort_direction: "desc",
        },
        { silent: true },
      )
        .then((response) => {
          if (!active) return;
          setSearchResults(response.data);
          setSearchState("ready");
        })
        .catch(() => {
          if (!active) return;
          setSearchResults([]);
          setSearchState("error");
        });
    }, 180);

    return () => {
      active = false;
      window.clearTimeout(id);
    };
  }, [trimmedSearchQuery]);

  function closeSearchResults() {
    setSearchQuery("");
    setIsSearchFocused(false);
  }

  function handleDeleteProjectSelect() {
    setIsProjectMenuOpen(false);
    onDeleteProject?.();
  }

  return (
    <header className={styles.header}>
      <div className={styles.utilityRow}>
        <Breadcrumb />

        <div className={styles.globalActions}>
          <div
            className={styles.searchShell}
            onBlur={(event) => {
              if (!event.currentTarget.contains(event.relatedTarget)) {
                setIsSearchFocused(false);
              }
            }}
          >
            <label className={styles.search}>
              <Search aria-hidden="true" size={16} strokeWidth={1.8} />
              <input
                type="search"
                placeholder={t("header.searchPlaceholder")}
                value={searchQuery}
                aria-label={t("header.searchPlaceholder")}
                aria-controls={showSearchResults ? searchResultsID : undefined}
                aria-expanded={showSearchResults}
                onChange={(event) => setSearchQuery(event.target.value)}
                onFocus={() => setIsSearchFocused(true)}
              />
              <kbd>{t("header.commandKey")}</kbd>
            </label>
            {showSearchResults ? (
              <div className={styles.searchResults} id={searchResultsID}>
                <p className={styles.searchResultsTitle}>{t("header.searchResults")}</p>
                {searchState === "loading" ? (
                  <p className={styles.searchMessage}>{t("header.searchLoading")}</p>
                ) : null}
                {searchState === "error" ? (
                  <p className={styles.searchMessage}>{t("header.searchError")}</p>
                ) : null}
                {searchState === "ready" && searchResults.length === 0 ? (
                  <p className={styles.searchMessage}>{t("header.searchEmpty")}</p>
                ) : null}
                {searchResults.length > 0 ? (
                  <ul className={styles.searchResultList}>
                    {searchResults.map((issue) => (
                      <li key={issue.id}>
                        <Link
                          className={styles.searchResultLink}
                          to={`/issues/${issue.id}`}
                          onClick={closeSearchResults}
                        >
                          <span className={styles.searchResultID}>#{issue.id}</span>
                          <span className={styles.searchResultTitle}>{issue.title}</span>
                          <span className={styles.searchResultMeta}>{issue.projectKey}</span>
                        </Link>
                      </li>
                    ))}
                  </ul>
                ) : null}
              </div>
            ) : null}
          </div>
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
          {activePage === "issues" && canDeleteProject ? (
            <ContextMenu
              id={projectMenuID}
              isOpen={isProjectMenuOpen}
              label={t("header.projectActionsMenu")}
              onOpenChange={setIsProjectMenuOpen}
              trigger={(triggerProps) => (
                <button
                  type="button"
                  className={styles.moreButton}
                  aria-label={t("header.moreProjectActions")}
                  {...triggerProps}
                >
                  <MoreHorizontal aria-hidden="true" size={19} strokeWidth={1.8} />
                </button>
              )}
            >
              <ContextMenuItem
                label={t("header.deleteProject")}
                onSelect={handleDeleteProjectSelect}
              >
                {t("header.deleteProject")}
              </ContextMenuItem>
            </ContextMenu>
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
