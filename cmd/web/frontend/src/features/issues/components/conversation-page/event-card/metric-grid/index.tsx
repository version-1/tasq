import styles from "../index.module.css";

export function MetricGrid({ items }: { items: Array<[string, string]> }) {
  return (
    <dl className={styles.metricGrid}>
      {items.map(([label, value]) => (
        <div key={label} className={styles.metric}>
          <dt>{label}</dt>
          <dd>{value}</dd>
        </div>
      ))}
    </dl>
  );
}
