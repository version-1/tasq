import type { ReactNode } from "react";
import styles from "../index.module.css";

export function FormField({ children, label }: { children: ReactNode; label: string }) {
  return (
    <label className={styles.field}>
      <span>{label}</span>
      {children}
    </label>
  );
}
