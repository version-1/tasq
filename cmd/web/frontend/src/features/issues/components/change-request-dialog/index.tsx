import { useEffect, useState, type FormEvent } from "react";
import { useTranslation } from "react-i18next";
import { MarkdownEditor } from "@/components/ui/markdown-editor";
import { createChangeRequest } from "@/lib/api";
import styles from "./index.module.css";

const changeRequestAuthor = "reviewer";

type ChangeRequestDialogVariant = "continue" | "reject";

export function ChangeRequestDialog({
  error,
  isMovingIssue,
  issueID,
  issueTitle,
  initialBody = "",
  initialRequestCreated = false,
  onCancel,
  onMoveIssueReady,
  onSuccess,
  variant,
}: {
  error?: string;
  isMovingIssue?: boolean;
  issueID: number;
  issueTitle: string;
  initialBody?: string;
  initialRequestCreated?: boolean;
  onCancel: () => void;
  onMoveIssueReady: () => Promise<void>;
  onSuccess: () => void;
  variant: ChangeRequestDialogVariant;
}) {
  const { t } = useTranslation();
  const translationKey = variant === "continue" ? "issues.continueWithComment" : "issues.reject";
  const [body, setBody] = useState(initialBody);
  const [validationError, setValidationError] = useState("");
  const [submitError, setSubmitError] = useState("");
  const [isCreatingRequest, setIsCreatingRequest] = useState(false);
  const [hasCreatedRequest, setHasCreatedRequest] = useState(initialRequestCreated);
  const isSubmitting = isCreatingRequest || isMovingIssue === true;

  useEffect(() => {
    setBody(initialBody);
    setValidationError("");
    setSubmitError("");
    setIsCreatingRequest(false);
    setHasCreatedRequest(initialRequestCreated);
  }, [initialBody, initialRequestCreated, issueID]);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const trimmedBody = body.trim();
    if (!trimmedBody) {
      setValidationError(t(`${translationKey}.errors.bodyRequired`));
      return;
    }

    setValidationError("");
    setSubmitError("");
    try {
      if (!hasCreatedRequest) {
        setIsCreatingRequest(true);
        await createChangeRequest(issueID, {
          author: changeRequestAuthor,
          body: trimmedBody,
        }, { silent: true });
        setHasCreatedRequest(true);
      }
      setIsCreatingRequest(false);
      await onMoveIssueReady();
      onSuccess();
    } catch (caught) {
      setSubmitError(
        caught instanceof Error ? caught.message : t(`${translationKey}.errors.submitFailed`),
      );
    } finally {
      setIsCreatingRequest(false);
    }
  }

  const errorMessage = validationError || submitError || error || "";

  return (
    <div className={styles.dialogBackdrop} role="presentation">
      <section
        aria-labelledby="change-request-title"
        aria-modal="true"
        className={styles.dialog}
        role="dialog"
      >
        <form className={styles.form} onSubmit={(event) => void handleSubmit(event)}>
          <div className={styles.dialogHeader}>
            <h2 id="change-request-title">{t(`${translationKey}.title`, { id: issueID })}</h2>
            <button
              type="button"
              aria-label={t(`${translationKey}.close`)}
              disabled={isSubmitting}
              onClick={onCancel}
            >
              ×
            </button>
          </div>

          <p className={styles.issueTitle}>{issueTitle}</p>

          <div className={styles.markdownEditorField}>
            <span>{t(`${translationKey}.fields.body`)}</span>
            <MarkdownEditor
              initialMode="edit"
              initialTab="raw"
              labels={{
                cancel: t("markdownEditor.cancel"),
                edit: t("markdownEditor.edit"),
                empty: t(`${translationKey}.emptyRequest`),
                preview: t("markdownEditor.preview"),
                raw: t("markdownEditor.raw"),
                save: t("markdownEditor.save"),
                saving: t("markdownEditor.saving"),
                textarea: t(`${translationKey}.fields.body`),
              }}
              readOnly={hasCreatedRequest}
              showActions={false}
              stablePanelRows={12}
              value={body}
              onChange={setBody}
              rows={12}
            />
          </div>

          {hasCreatedRequest ? (
            <p className={styles.retryNote}>{t(`${translationKey}.retryNote`)}</p>
          ) : null}
          {errorMessage ? <p className={styles.formError}>{errorMessage}</p> : null}

          <div className={styles.dialogActions}>
            <button type="button" disabled={isSubmitting} onClick={onCancel}>
              {t(`${translationKey}.cancel`)}
            </button>
            <button type="submit" disabled={isSubmitting}>
              {isSubmitting ? t(`${translationKey}.saving`) : t(`${translationKey}.submit`)}
            </button>
          </div>
        </form>
      </section>
    </div>
  );
}
