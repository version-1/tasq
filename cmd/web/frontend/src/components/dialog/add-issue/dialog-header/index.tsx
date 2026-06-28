import styles from "../index.module.css";

export function DialogHeader({
  closeLabel,
  title,
  titleID,
  onCancel,
}: {
  closeLabel: string;
  title: string;
  titleID: string;
  onCancel: () => void;
}) {
  return (
    <div className={styles.dialogHeader}>
      <h2 id={titleID}>{title}</h2>
      <button type="button" aria-label={closeLabel} onClick={onCancel}>
        ×
      </button>
    </div>
  );
}
