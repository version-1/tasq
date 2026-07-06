import { useState, type FormEvent } from "react";
import { useTranslation } from "react-i18next";
import type { Project } from "@/lib/types";
import styles from "./index.module.css";

export function DeleteProjectDialog({
  error,
  isDeleting,
  onCancel,
  onConfirm,
  project,
}: {
  error: string;
  isDeleting: boolean;
  onCancel: () => void;
  onConfirm: () => Promise<void>;
  project: Project;
}) {
  const { t } = useTranslation();
  const [confirmationKey, setConfirmationKey] = useState("");
  const canDelete = confirmationKey === project.key && !isDeleting;

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!canDelete) return;
    void onConfirm();
  }

  return (
    <div className={styles.dialogBackdrop} role="presentation">
      <section
        aria-labelledby="delete-project-title"
        aria-modal="true"
        className={styles.dialog}
        role="dialog"
      >
        <form className={styles.dialogContent} onSubmit={handleSubmit}>
          <div className={styles.dialogHeader}>
            <div>
              <h2 id="delete-project-title">{t("deleteProject.title", { key: project.key })}</h2>
              <p>{t("deleteProject.description")}</p>
            </div>
            <button type="button" aria-label={t("deleteProject.close")} onClick={onCancel}>
              x
            </button>
          </div>

          <div className={styles.warningBox}>
            <strong>{t("deleteProject.irreversibleTitle")}</strong>
            <p>{t("deleteProject.irreversibleDescription")}</p>
          </div>

          <ul className={styles.deletedItems}>
            <li>{t("deleteProject.deletedItems.issues")}</li>
            <li>{t("deleteProject.deletedItems.runs")}</li>
            <li>{t("deleteProject.deletedItems.workflow")}</li>
          </ul>

          <label className={styles.field}>
            <span>{t("deleteProject.confirmationLabel", { key: project.key })}</span>
            <input
              autoComplete="off"
              value={confirmationKey}
              onChange={(event) => setConfirmationKey(event.target.value)}
            />
          </label>

          {error ? <p className={styles.errorMessage}>{error}</p> : null}

          <div className={styles.dialogActions}>
            <button type="button" onClick={onCancel}>
              {t("deleteProject.cancel")}
            </button>
            <button className={styles.deleteButton} disabled={!canDelete} type="submit">
              {isDeleting ? t("deleteProject.deleting") : t("deleteProject.submit")}
            </button>
          </div>
        </form>
      </section>
    </div>
  );
}
