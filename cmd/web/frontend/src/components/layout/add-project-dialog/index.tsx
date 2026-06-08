import { useState, type FormEvent } from "react";
import { useTranslation } from "react-i18next";
import type { CreateProjectInput } from "@/lib/types";
import styles from "./index.module.css";

type AddProjectFormValues = {
  key: string;
  name: string;
  description: string;
  location: string;
};

export function AddProjectDialog({
  error,
  onCancel,
  onSubmit,
}: {
  error: string;
  onCancel: () => void;
  onSubmit: (input: CreateProjectInput) => Promise<void>;
}) {
  const { t } = useTranslation();
  const [values, setValues] = useState<AddProjectFormValues>(initialAddProjectValues);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [validationError, setValidationError] = useState("");

  async function handleChooseDirectory() {
    const selectedDirectory = await chooseProjectDirectory();
    if (!selectedDirectory) return;

    setValues({
      ...values,
      key: values.key || toProjectKey(selectedDirectory.name),
      location: values.location || selectedDirectory.name,
      name: values.name || toProjectName(selectedDirectory.name),
    });
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const input = toCreateProjectInput(values);
    if (!input.location) {
      setValidationError(t("addProject.errors.locationRequired"));
      return;
    }
    if (!isAbsoluteLocation(input.location)) {
      setValidationError(t("addProject.errors.locationAbsolute"));
      return;
    }
    if (!input.key) {
      setValidationError(t("addProject.errors.keyRequired"));
      return;
    }
    if (!input.name) {
      setValidationError(t("addProject.errors.nameRequired"));
      return;
    }

    setIsSubmitting(true);
    setValidationError("");
    await onSubmit(input);
    setIsSubmitting(false);
  }

  const errorMessage = validationError || error;

  return (
    <div className={styles.dialogBackdrop} role="presentation">
      <section
        aria-labelledby="add-project-title"
        aria-modal="true"
        className={styles.dialog}
        role="dialog"
      >
        <form className={styles.addProjectForm} onSubmit={(event) => void handleSubmit(event)}>
          <div className={styles.dialogHeader}>
            <h2 id="add-project-title">{t("addProject.title")}</h2>
            <button type="button" aria-label={t("addProject.close")} onClick={onCancel}>
              x
            </button>
          </div>

          <label className={styles.field}>
            <span>{t("addProject.fields.location")}</span>
            <button
              aria-label={t("addProject.fields.chooseDirectory")}
              className={styles.directoryButton}
              type="button"
              autoFocus
              onClick={() => void handleChooseDirectory()}
            >
              {t("addProject.fields.chooseDirectory")}
            </button>
            <span className={styles.selectedLocation}>
              {values.location || t("addProject.placeholders.directory")}
            </span>
            <input
              aria-label={t("addProject.fields.location")}
              value={values.location}
              onChange={(event) => setValues({ ...values, location: event.target.value })}
              placeholder={t("addProject.placeholders.location")}
            />
          </label>

          <label className={styles.field}>
            <span>{t("addProject.fields.key")}</span>
            <input
              value={values.key}
              onChange={(event) => setValues({ ...values, key: event.target.value })}
              placeholder={t("addProject.placeholders.key")}
            />
          </label>

          <label className={styles.field}>
            <span>{t("addProject.fields.name")}</span>
            <input
              value={values.name}
              onChange={(event) => setValues({ ...values, name: event.target.value })}
              placeholder={t("addProject.placeholders.name")}
            />
          </label>

          <label className={styles.field}>
            <span>{t("addProject.fields.description")}</span>
            <textarea
              value={values.description}
              onChange={(event) => setValues({ ...values, description: event.target.value })}
              placeholder={t("addProject.placeholders.description")}
              rows={3}
            />
          </label>

          {errorMessage ? <p className={styles.formError}>{errorMessage}</p> : null}

          <div className={styles.dialogActions}>
            <button type="button" onClick={onCancel} disabled={isSubmitting}>
              {t("addProject.cancel")}
            </button>
            <button type="submit" disabled={isSubmitting}>
              {isSubmitting ? t("addProject.saving") : t("addProject.submit")}
            </button>
          </div>
        </form>
      </section>
    </div>
  );
}

const initialAddProjectValues: AddProjectFormValues = {
  key: "",
  name: "",
  description: "",
  location: "",
};

function toCreateProjectInput(values: AddProjectFormValues): CreateProjectInput {
  return {
    key: values.key.trim(),
    name: values.name.trim(),
    description: values.description.trim() || undefined,
    location: values.location.trim(),
  };
}

type SelectedProjectLocation = {
  name: string;
};

async function chooseProjectDirectory(): Promise<SelectedProjectLocation | null> {
  const picker = window as Window & {
    showDirectoryPicker?: () => Promise<{ name: string }>;
  };
  if (!picker.showDirectoryPicker) {
    return null;
  }

  try {
    const directory = await picker.showDirectoryPicker();
    if (!directory.name) {
      return null;
    }
    return { name: directory.name };
  } catch {
    return null;
  }
}

function isAbsoluteLocation(location: string): boolean {
  return location.startsWith("/") || /^[A-Za-z]:[\\/]/.test(location);
}

function toProjectKey(name: string): string {
  return name
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "")
    .slice(0, 64);
}

function toProjectName(name: string): string {
  return name
    .trim()
    .split(/[-_\s]+/)
    .filter(Boolean)
    .map((part) => `${part.charAt(0).toUpperCase()}${part.slice(1)}`)
    .join(" ");
}
