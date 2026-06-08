"use client";

import { Link, useLocation, useParams } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { DefaultLayout } from "@/components/layout/default";
import type { ReactNode } from "react";
import styles from "./index.module.css";

export function IssueDetailLayout({ children }: { children: ReactNode }) {
  const { t } = useTranslation();
  const { id } = useParams();
  const issuePath = id ? `/issues/${id}` : "/issues";
  const conversationPath = `${issuePath}/conversations`;
  const location = useLocation();
  const isConversationRoute = location.pathname.includes("/conversations");
  const isDetailRoute = !isConversationRoute;

  return (
    <DefaultLayout>
      <div className={styles.layout}>
        <Link className={styles.backLink} to="/issues">
          {t("issues.detailPage.backToList")}
        </Link>
        <nav
          aria-label={t("issues.detailPage.navigation")}
          className={styles.tabs}
        >
          <Link
            aria-current={isDetailRoute ? "page" : undefined}
            className={tabClass(isDetailRoute)}
            to={issuePath}
          >
            {t("issues.detailPage.detailTab")}
          </Link>
          <Link
            aria-current={isConversationRoute ? "page" : undefined}
            className={tabClass(isConversationRoute)}
            to={conversationPath}
          >
            {t("issues.detailPage.conversationTab")}
          </Link>
        </nav>
        {children}
      </div>
    </DefaultLayout>
  );
}

function tabClass(isActive: boolean): string {
  return isActive ? `${styles.tab} ${styles.activeTab}` : styles.tab;
}
