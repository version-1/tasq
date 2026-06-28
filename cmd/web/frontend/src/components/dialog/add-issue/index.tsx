import { useEffect, useState, type FormEvent } from "react";
import { useTranslation } from "react-i18next";
import { issueStatuses, priorities } from "@/lib/types";
import { MarkdownEditor } from "@/components/ui/markdown-editor";
import type {
  CreateIssueInput,
  IssueStatus,
  IssueSummary,
  Priority,
  Project,
} from "@/lib/types";
import { DependencySelect } from "./dependency-select";
import { DialogActions } from "./dialog-actions";
import { DialogHeader } from "./dialog-header";
import { FormField } from "./form-field";
import { FormSelect } from "./form-select";
import styles from "./index.module.css";
import type { AddIssueFormValues } from "./types";

const defaultAddIssuePriority: Priority = "normal";
const dependencySelectableStatuses = new Set<IssueStatus>([
  "backlog",
  "blocked",
  "ready",
  "in_progress",
]);

export function AddIssueDialog({
  dependencyOptions,
  error,
  initialStatus,
  project,
  projects,
  onCancel,
  onSubmit,
}: {
  dependencyOptions: IssueSummary[];
  error: string;
  initialStatus: IssueStatus;
  project: Project | null;
  projects: Project[];
  onCancel: () => void;
  onSubmit: (input: CreateIssueInput) => Promise<void>;
}) {
  const { t } = useTranslation();
  const initialSelectedProjectID = project?.id ?? projects[0]?.id ?? null;
  const [values, setValues] = useState<AddIssueFormValues>(() =>
    initialAddIssueValues(initialStatus, initialSelectedProjectID),
  );
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [validationError, setValidationError] = useState("");
  const selectedProject = values.projectID === null
    ? null
    : (projects.find((candidate) => candidate.id === values.projectID) ?? null);
  const selectableDependencyOptions = dependencyOptions.filter((issue) =>
    issue.projectId === values.projectID && dependencySelectableStatuses.has(issue.status),
  );

  useEffect(() => {
    setValues(initialAddIssueValues(initialStatus, initialSelectedProjectID));
    setValidationError("");
    setIsSubmitting(false);
  }, [initialStatus, initialSelectedProjectID]);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const title = values.title.trim();
    if (!title) {
      setValidationError(t("addIssue.errors.titleRequired"));
      return;
    }
    if (!selectedProject) {
      setValidationError(t("addIssue.errors.projectRequired"));
      return;
    }

    setIsSubmitting(true);
    setValidationError("");
    await onSubmit(toCreateIssueInput(values, title, selectedProject.id));
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
          <DialogHeader
            closeLabel={t("addIssue.close")}
            title={t("addIssue.title")}
            titleID="add-issue-title"
            onCancel={onCancel}
          />

          <FormField label={t("addIssue.fields.title")}>
            <input
              autoFocus
              value={values.title}
              onChange={(event) => setValues({ ...values, title: event.target.value })}
              placeholder={t("addIssue.placeholders.title")}
            />
          </FormField>

          <div className={styles.markdownEditorField}>
            <span>{t("addIssue.fields.description")}</span>
            <MarkdownEditor
              initialMode="edit"
              initialTab="raw"
              labels={{
                cancel: t("markdownEditor.cancel"),
                edit: t("markdownEditor.edit"),
                empty: t("issues.noDescription"),
                preview: t("markdownEditor.preview"),
                raw: t("markdownEditor.raw"),
                save: t("markdownEditor.save"),
                saving: t("markdownEditor.saving"),
                textarea: t("addIssue.fields.description"),
              }}
              showActions={false}
              stablePanelRows={32}
              value={values.description}
              onChange={(description) => setValues({ ...values, description })}
              rows={32}
            />
          </div>

          <FormSelect
            label={t("addIssue.fields.project")}
            options={projects.map((projectOption) => ({
              label: projectOption.name,
              value: String(projectOption.id),
            }))}
            placeholder={{
              label: t("addIssue.placeholders.project"),
              value: "",
            }}
            value={values.projectID === null ? "" : String(values.projectID)}
            onValueChange={(value) =>
              setValues({
                ...values,
                dependencyIDs: [],
                projectID: parseProjectID(value),
              })
            }
          />

          <div className={styles.formGrid}>
            <FormSelect
              label={t("addIssue.fields.status")}
              options={issueStatuses.map((status) => ({
                label: t(`statuses.${status}`),
                value: status,
              }))}
              value={values.status}
              onValueChange={(value) => setValues({ ...values, status: value as IssueStatus })}
            />

            <FormSelect
              label={t("addIssue.fields.priority")}
              options={priorities.map((priority) => ({
                label: t(`priorities.${priority}`),
                value: priority,
              }))}
              value={values.priority}
              onValueChange={(value) => setValues({ ...values, priority: value as Priority })}
            />
          </div>

          <FormField label={t("addIssue.fields.assignee")}>
            <input
              value={values.assignee}
              onChange={(event) => setValues({ ...values, assignee: event.target.value })}
              placeholder={t("addIssue.placeholders.assignee")}
            />
          </FormField>

          <DependencySelect
            emptyLabel={t("addIssue.emptyDependencies")}
            getStatusLabel={(issue) => t(`statuses.${issue.status}`)}
            label={t("addIssue.fields.dependencies")}
            options={selectableDependencyOptions}
            placeholder={t("addIssue.placeholders.dependencies")}
            resetKey={values.projectID}
            selectedCountLabel={(count) => t("addIssue.dependencySelectedCount", { count })}
            selectedIDs={values.dependencyIDs}
            onChange={(dependencyIDs) => setValues({ ...values, dependencyIDs })}
          />

          {errorMessage ? <p className={styles.formError}>{errorMessage}</p> : null}

          <DialogActions
            cancelLabel={t("addIssue.cancel")}
            isSubmitting={isSubmitting}
            savingLabel={t("addIssue.saving")}
            submitLabel={t("addIssue.submit")}
            onCancel={onCancel}
          />
        </form>
      </section>
    </div>
  );
}

function initialAddIssueValues(status: IssueStatus, projectID: number | null): AddIssueFormValues {
  return {
    title: "",
    description: "",
    projectID,
    status,
    priority: defaultAddIssuePriority,
    assignee: "",
    dependencyIDs: [],
  };
}

function toCreateIssueInput(values: AddIssueFormValues, title: string, projectID: number): CreateIssueInput {
  const description = descriptionForCreate(values.description);
  const input: CreateIssueInput = {
    projectId: projectID,
    title,
    status: values.status,
    priority: values.priority,
    assignee: values.assignee.trim() || undefined,
  };
  if (description !== undefined) {
    input.description = description;
  }
  if (values.dependencyIDs.length > 0) {
    input.dependency_ids = values.dependencyIDs;
  }
  return input;
}

function descriptionForCreate(description: string): string | undefined {
  return description.trim() === "" ? undefined : description;
}

function parseProjectID(value: string): number | null {
  const id = Number.parseInt(value, 10);
  return Number.isSafeInteger(id) && id > 0 ? id : null;
}
