"use client";

import type { ReactNode } from "react";
import { useTranslation } from "react-i18next";
import { Link } from "react-router-dom";
import { DefaultLayout } from "@/components/layout/default";
import styles from "./index.module.css";

export function IssueDetailLayout({ children }: { children: ReactNode }) {
  const { t } = useTranslation();

  return (
    <DefaultLayout>
      <div className={styles.layout}>
        {children}
        <Link className={styles.backLink} to="/dashboard">
          {t("issues.detailPage.backToList")}
        </Link>
      </div>
    </DefaultLayout>
  );
}
