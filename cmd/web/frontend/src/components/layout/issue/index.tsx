"use client";

import type { ReactNode } from "react";
import { useTranslation } from "react-i18next";
import { Link, useLocation, useParams } from "react-router-dom";
import { DefaultLayout } from "@/components/layout/default";
import type { HeaderPageLink } from "@/components/layout/header";
import styles from "./index.module.css";

export function IssueDetailLayout({
  children,
  pages,
}: {
  children: ReactNode;
  pages: readonly HeaderPageLink[];
}) {
  const { t } = useTranslation();
  const { id } = useParams();
  const issuePath = id ? `/issues/${id}` : "/issues";
  const conversationPath = `${issuePath}/conversations`;
  const location = useLocation();

  return (
    <DefaultLayout pages={pages}>
      <div className={styles.layout}>
        {children}
        <Link className={styles.backLink} to="/issues">
          {t("issues.detailPage.backToList")}
        </Link>
      </div>
    </DefaultLayout>
  );
}
