import { useTranslation } from "react-i18next";
import type { Issue } from "@/lib/types";
import { MarkdownEditor } from "@/components/ui/markdown-editor";
import styles from "./index.module.css";

export function IssueDescription({
  issue,
  isSaving = false,
  onSave,
}: {
  issue: Issue;
  isSaving?: boolean;
  onSave: (description: string) => Promise<void>;
}) {
  const { t } = useTranslation();

  return (
    <section className={styles.section} aria-labelledby="issue-description">
      <MarkdownEditor
        isSaving={isSaving}
        labels={{
          cancel: t("markdownEditor.cancel"),
          edit: t("markdownEditor.edit"),
          empty: t("issues.noDescription"),
          preview: t("markdownEditor.preview"),
          raw: t("markdownEditor.raw"),
          save: t("markdownEditor.save"),
          saving: t("markdownEditor.saving"),
          textarea: t("issues.detailPage.description"),
        }}
        title={t("issues.detailPage.description")}
        titleID="issue-description"
        value={issue.description}
        onSave={onSave}
        rows={128}
      />
    </section>
  );
}
