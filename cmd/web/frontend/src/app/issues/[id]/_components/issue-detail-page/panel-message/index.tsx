import styles from "./index.module.css";

export function PanelMessage({ title, detail }: { title: string; detail?: string }) {
  return (
    <section className={styles.section}>
      <h2>{title}</h2>
      {detail ? <p className={styles.errorText}>{detail}</p> : null}
    </section>
  );
}
