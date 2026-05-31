"use client";

import { useTranslation } from "react-i18next";
import styles from "./index.module.css";

export function AgentsView() {
  const { t } = useTranslation();

  return (
    <section className={styles.panelGrid}>
      <div className={styles.widePanel}>
        <h2>{t("agents.title")}</h2>
        <p className={styles.empty}>{t("agents.empty")}</p>
      </div>
    </section>
  );
}
