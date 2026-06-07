import styles from "./index.module.css";

export function PanelMessage({ title }: { title: string }) {
  return (
    <section className={styles.widePanel}>
      <h2>{title}</h2>
    </section>
  );
}
