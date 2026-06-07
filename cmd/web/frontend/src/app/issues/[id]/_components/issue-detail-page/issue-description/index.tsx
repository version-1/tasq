import { useTranslation } from "react-i18next";
import type { Issue } from "@/lib/types";
import { Markdown } from "@/components/issue/markdown";
import styles from "./index.module.css";

export function IssueDescription({ issue }: { issue: Issue }) {
  const { t } = useTranslation();

  return (
    <section className={styles.section} aria-labelledby="issue-description">
      <div className={styles.sectionHeader}>
        <h3 id="issue-description">{t("issues.detailPage.description")}</h3>
      </div>
      <Markdown
        className={styles.markdown}
        content={issue.description}
        emptyText={t("issues.noDescription")}
      />
    </section>
  );
}
