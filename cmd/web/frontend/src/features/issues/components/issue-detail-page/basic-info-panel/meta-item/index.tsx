import type { ReactNode } from "react";
import styles from "./index.module.css";

export function MetaItem({ label, value }: { label: string; value: ReactNode }) {
  return (
    <div className={styles.metaItem}>
      <dt>{label}</dt>
      <dd>{value}</dd>
    </div>
  );
}
