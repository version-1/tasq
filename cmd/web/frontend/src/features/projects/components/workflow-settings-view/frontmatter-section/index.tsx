import { useTranslation } from "react-i18next";
import type { ProjectWorkflowFrontmatter } from "@/lib/types";
import { FrontmatterTable } from "../frontmatter-table";
import { toFrontmatterRows } from "../frontmatter-table/rows";
import styles from "./index.module.css";

type FrontmatterSectionProps = {
  frontmatter: ProjectWorkflowFrontmatter;
};

export function FrontmatterSection({ frontmatter }: FrontmatterSectionProps) {
  const { t } = useTranslation();
  const hasFrontmatter = Object.keys(frontmatter).length > 0;

  return (
    <section className={styles.panel} aria-labelledby="project-workflow-frontmatter">
      <h2 className={styles.heading} id="project-workflow-frontmatter">
        {t("projectSettings.frontmatter")}
      </h2>
      {hasFrontmatter ? (
        <FrontmatterTable rows={toFrontmatterRows(frontmatter)} />
      ) : (
        <p className={styles.emptyText}>{t("projectSettings.emptyFrontmatter")}</p>
      )}
    </section>
  );
}
