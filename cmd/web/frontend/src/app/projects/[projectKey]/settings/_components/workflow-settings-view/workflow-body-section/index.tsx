import { useTranslation } from "react-i18next";
import { Markdown } from "@/components/issue/markdown";
import styles from "./index.module.css";

type WorkflowBodySectionProps = {
  body: string;
};

export function WorkflowBodySection({ body }: WorkflowBodySectionProps) {
  const { t } = useTranslation();

  return (
    <section className={styles.panel} aria-labelledby="project-workflow-body">
      <h2 className={styles.heading} id="project-workflow-body">
        {t("projectSettings.body")}
      </h2>
      <Markdown
        className={styles.markdown}
        content={body}
        emptyText={t("projectSettings.emptyBody")}
      />
    </section>
  );
}
