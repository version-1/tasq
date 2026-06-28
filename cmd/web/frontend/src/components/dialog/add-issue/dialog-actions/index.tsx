import styles from "../index.module.css";

export function DialogActions({
  cancelLabel,
  isSubmitting,
  savingLabel,
  submitLabel,
  onCancel,
}: {
  cancelLabel: string;
  isSubmitting: boolean;
  savingLabel: string;
  submitLabel: string;
  onCancel: () => void;
}) {
  return (
    <div className={styles.dialogActions}>
      <button type="button" onClick={onCancel} disabled={isSubmitting}>
        {cancelLabel}
      </button>
      <button type="submit" disabled={isSubmitting}>
        {isSubmitting ? savingLabel : submitLabel}
      </button>
    </div>
  );
}
