import { useEffect, useState, type FormEvent } from "react";
import { useTranslation } from "react-i18next";
import { MarkdownEditor } from "@/components/ui/markdown-editor";
import { createChangeRequest } from "@/lib/api";
import styles from "./index.module.css";

const rejectAuthor = "reviewer";

export function RejectIssueDialog({
  error,
  isMovingIssue,
  issueID,
  issueTitle,
  onCancel,
  onMoveIssueReady,
  onSuccess,
}: {
  error?: string;
  isMovingIssue?: boolean;
  issueID: number;
  issueTitle: string;
  onCancel: () => void;
  onMoveIssueReady: () => Promise<void>;
  onSuccess: () => void;
}) {
  const { t } = useTranslation();
  const [body, setBody] = useState("");
  const [validationError, setValidationError] = useState("");
  const [submitError, setSubmitError] = useState("");
  const [isCreatingRequest, setIsCreatingRequest] = useState(false);
  const [hasCreatedRequest, setHasCreatedRequest] = useState(false);
  const isSubmitting = isCreatingRequest || isMovingIssue === true;

  useEffect(() => {
    setBody("");
    setValidationError("");
    setSubmitError("");
    setIsCreatingRequest(false);
    setHasCreatedRequest(false);
  }, [issueID]);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const trimmedBody = body.trim();
    if (!trimmedBody) {
      setValidationError(t("issues.reject.errors.bodyRequired"));
      return;
    }

    setValidationError("");
    setSubmitError("");
    try {
      if (!hasCreatedRequest) {
        setIsCreatingRequest(true);
        await createChangeRequest(issueID, {
          author: rejectAuthor,
          body: trimmedBody,
        }, { silent: true });
        setHasCreatedRequest(true);
      }
      setIsCreatingRequest(false);
      await onMoveIssueReady();
      onSuccess();
    } catch (caught) {
      setSubmitError(caught instanceof Error ? caught.message : t("issues.reject.errors.submitFailed"));
    } finally {
      setIsCreatingRequest(false);
    }
  }

  const errorMessage = validationError || submitError || error || "";

  return (
    <div className={styles.dialogBackdrop} role="presentation">
      <section
        aria-labelledby="reject-issue-title"
        aria-modal="true"
        className={styles.dialog}
        role="dialog"
      >
        <form className={styles.form} onSubmit={(event) => void handleSubmit(event)}>
          <div className={styles.dialogHeader}>
            <h2 id="reject-issue-title">{t("issues.reject.title", { id: issueID })}</h2>
            <button
              type="button"
              aria-label={t("issues.reject.close")}
              disabled={isSubmitting}
              onClick={onCancel}
            >
              ×
            </button>
          </div>

          <p className={styles.issueTitle}>{issueTitle}</p>

          <div className={styles.markdownEditorField}>
            <span>{t("issues.reject.fields.body")}</span>
            <MarkdownEditor
              initialMode="edit"
              initialTab="raw"
              labels={{
                cancel: t("markdownEditor.cancel"),
                edit: t("markdownEditor.edit"),
                empty: t("issues.reject.emptyRequest"),
                preview: t("markdownEditor.preview"),
                raw: t("markdownEditor.raw"),
                save: t("markdownEditor.save"),
                saving: t("markdownEditor.saving"),
                textarea: t("issues.reject.fields.body"),
              }}
              showActions={false}
              stablePanelRows={12}
              value={body}
              onChange={setBody}
              rows={12}
            />
          </div>

          {hasCreatedRequest ? <p className={styles.retryNote}>{t("issues.reject.retryNote")}</p> : null}
          {errorMessage ? <p className={styles.formError}>{errorMessage}</p> : null}

          <div className={styles.dialogActions}>
            <button type="button" disabled={isSubmitting} onClick={onCancel}>
              {t("issues.reject.cancel")}
            </button>
            <button type="submit" disabled={isSubmitting}>
              {isSubmitting ? t("issues.reject.saving") : t("issues.reject.submit")}
            </button>
          </div>
        </form>
      </section>
    </div>
  );
}
