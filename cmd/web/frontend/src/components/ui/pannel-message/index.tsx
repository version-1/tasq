import styles from "./index.module.css";

export function PanelMessage({ title, detail }: { title: string; detail?: string }) {
  return (
    <section className={styles.widePanel}>
      <h2>{title}</h2>
      {detail ? <p>{detail}</p> : null}
    </section>
  );
}
