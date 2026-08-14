import { useEffect, useRef, useState, type FormEvent, type ReactNode } from "react";
import { useTranslation } from "react-i18next";
import { MarkdownEditor } from "@/components/ui/markdown-editor";
import { createChangeRequest } from "@/lib/api";
import type { ChangeRequestShortcut } from "@/features/issues/change-request-shortcuts";
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
  isSubmissionDisabled = false,
  context,
  onCancel,
  onMoveIssueReady,
  onSuccess,
  shortcuts = [],
  variant,
}: {
  error?: string;
  isMovingIssue?: boolean;
  issueID: number;
  issueTitle: string;
  initialBody?: string;
  initialRequestCreated?: boolean;
  isSubmissionDisabled?: boolean;
  context?: ReactNode;
  onCancel: () => void;
  onMoveIssueReady: () => Promise<void>;
  onSuccess: () => void;
  shortcuts?: readonly ChangeRequestShortcut[];
  variant: ChangeRequestDialogVariant;
}) {
  const { t } = useTranslation();
  const translationKey = variant === "continue" ? "issues.continueWithComment" : "issues.reject";
  const [body, setBody] = useState(initialBody);
  const [validationError, setValidationError] = useState("");
  const [submitError, setSubmitError] = useState("");
  const [isCreatingRequest, setIsCreatingRequest] = useState(false);
  const [hasCreatedRequest, setHasCreatedRequest] = useState(initialRequestCreated);
  const [editorRevision, setEditorRevision] = useState(0);
  const isSubmittingRef = useRef(false);
  const isSubmitting = isCreatingRequest || isMovingIssue === true;

  useEffect(() => {
    setBody(initialBody);
    setValidationError("");
    setSubmitError("");
    setIsCreatingRequest(false);
    setHasCreatedRequest(initialRequestCreated);
    setEditorRevision(0);
    isSubmittingRef.current = false;
  }, [initialBody, initialRequestCreated, issueID]);

  async function submit(requestBody: string) {
    if (isSubmittingRef.current || isSubmitting || isSubmissionDisabled) return;

    const trimmedBody = requestBody.trim();
    if (!trimmedBody) {
      setValidationError(t(`${translationKey}.errors.bodyRequired`));
      return;
    }

    setValidationError("");
    setSubmitError("");
    setBody(trimmedBody);
    isSubmittingRef.current = true;
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
      isSubmittingRef.current = false;
      setIsCreatingRequest(false);
    }
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    await submit(body);
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

          {context}

          {shortcuts.length > 0 ? (
            <div className={styles.shortcuts} aria-label={t(`${translationKey}.shortcuts`)}>
              {shortcuts.map((shortcut) => (
                <button
                  key={shortcut.id}
                  type="button"
                  disabled={isSubmitting || isSubmissionDisabled}
                  onClick={() => {
                    setBody(shortcut.body);
                    setEditorRevision((current) => current + 1);
                    void submit(shortcut.body);
                  }}
                >
                  {shortcut.label}
                </button>
              ))}
            </div>
          ) : null}

          <div className={styles.markdownEditorField}>
            <span>{t(`${translationKey}.fields.body`)}</span>
            <MarkdownEditor
              key={`${editorRevision}-${hasCreatedRequest ? `submitted-${body}` : "editable"}`}
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
            <button type="submit" disabled={isSubmitting || isSubmissionDisabled}>
              {isSubmitting ? t(`${translationKey}.saving`) : t(`${translationKey}.submit`)}
            </button>
          </div>
        </form>
      </section>
    </div>
  );
}
