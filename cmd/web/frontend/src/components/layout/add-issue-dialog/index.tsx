import { useEffect, useState, type FormEvent } from "react";
import { useTranslation } from "react-i18next";
import { issueStatuses, priorities } from "@/lib/types";
import type { CreateIssueInput, IssueStatus, Priority, Project } from "@/lib/types";
import styles from "./index.module.css";

const defaultAddIssuePriority: Priority = "normal";

type AddIssueFormValues = {
  title: string;
  description: string;
  status: IssueStatus;
  priority: Priority;
  assignee: string;
};

export function AddIssueDialog({
  error,
  initialStatus,
  project,
  onCancel,
  onSubmit,
}: {
  error: string;
  initialStatus: IssueStatus;
  project: Project | null;
  onCancel: () => void;
  onSubmit: (input: CreateIssueInput) => Promise<void>;
}) {
  const { t } = useTranslation();
  const [values, setValues] = useState<AddIssueFormValues>(() => initialAddIssueValues(initialStatus));
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [validationError, setValidationError] = useState("");

  useEffect(() => {
    setValues(initialAddIssueValues(initialStatus));
    setValidationError("");
    setIsSubmitting(false);
  }, [initialStatus]);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const title = values.title.trim();
    if (!title) {
      setValidationError(t("addIssue.errors.titleRequired"));
      return;
    }
    if (!project) {
      setValidationError(t("addIssue.errors.projectRequired"));
      return;
    }

    setIsSubmitting(true);
    setValidationError("");
    await onSubmit(toCreateIssueInput(values, title, project.id));
    setIsSubmitting(false);
  }

  const errorMessage = validationError || error;

  return (
    <div className={styles.dialogBackdrop} role="presentation">
      <section
        aria-labelledby="add-issue-title"
        aria-modal="true"
        className={styles.dialog}
        role="dialog"
      >
        <form className={styles.addIssueForm} onSubmit={(event) => void handleSubmit(event)}>
          <div className={styles.dialogHeader}>
            <h2 id="add-issue-title">{t("addIssue.title")}</h2>
            <button type="button" aria-label={t("addIssue.close")} onClick={onCancel}>
              ×
            </button>
          </div>

          <label className={styles.field}>
            <span>{t("addIssue.fields.title")}</span>
            <input
              autoFocus
              value={values.title}
              onChange={(event) => setValues({ ...values, title: event.target.value })}
              placeholder={t("addIssue.placeholders.title")}
            />
          </label>

          <label className={styles.field}>
            <span>{t("addIssue.fields.description")}</span>
            <textarea
              value={values.description}
              onChange={(event) => setValues({ ...values, description: event.target.value })}
              placeholder={t("addIssue.placeholders.description")}
              rows={4}
            />
          </label>

          <div className={styles.formGrid}>
            <label className={styles.field}>
              <span>{t("addIssue.fields.status")}</span>
              <select
                value={values.status}
                onChange={(event) =>
                  setValues({ ...values, status: event.target.value as IssueStatus })
                }
              >
                {issueStatuses.map((status) => (
                  <option key={status} value={status}>
                    {t(`statuses.${status}`)}
                  </option>
                ))}
              </select>
            </label>

            <label className={styles.field}>
              <span>{t("addIssue.fields.priority")}</span>
              <select
                value={values.priority}
                onChange={(event) =>
                  setValues({ ...values, priority: event.target.value as Priority })
                }
              >
                {priorities.map((priority) => (
                  <option key={priority} value={priority}>
                    {t(`priorities.${priority}`)}
                  </option>
                ))}
              </select>
            </label>
          </div>

          <label className={styles.field}>
            <span>{t("addIssue.fields.assignee")}</span>
            <input
              value={values.assignee}
              onChange={(event) => setValues({ ...values, assignee: event.target.value })}
              placeholder={t("addIssue.placeholders.assignee")}
            />
          </label>

          {errorMessage ? <p className={styles.formError}>{errorMessage}</p> : null}

          <div className={styles.dialogActions}>
            <button type="button" onClick={onCancel} disabled={isSubmitting}>
              {t("addIssue.cancel")}
            </button>
            <button type="submit" disabled={isSubmitting}>
              {isSubmitting ? t("addIssue.saving") : t("addIssue.submit")}
            </button>
          </div>
        </form>
      </section>
    </div>
  );
}

function initialAddIssueValues(status: IssueStatus): AddIssueFormValues {
  return {
    title: "",
    description: "",
    status,
    priority: defaultAddIssuePriority,
    assignee: "",
  };
}

function toCreateIssueInput(values: AddIssueFormValues, title: string, projectID: number): CreateIssueInput {
  return {
    projectId: projectID,
    title,
    description: values.description.trim() || undefined,
    status: values.status,
    priority: values.priority,
    assignee: values.assignee.trim() || undefined,
  };
}
