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
            <input
              aria-label={t("addProject.fields.location")}
              value={values.location}
              onChange={(event) => setValues({ ...values, location: event.target.value })}
              placeholder={t("addProject.placeholders.location")}
              autoFocus
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

function isAbsoluteLocation(location: string): boolean {
  return location.startsWith("/") || /^[A-Za-z]:[\\/]/.test(location);
}
