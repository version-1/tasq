import type { ReactNode } from "react";
import styles from "./index.module.css";

type BadgeVariant = "project" | "priority-high" | "priority-normal" | "priority-low";

export function Badge({
  children,
  icon,
  variant,
}: {
  children: ReactNode;
  icon?: ReactNode;
  variant: BadgeVariant;
}) {
  return (
    <span className={`${styles.badge} ${styles[variant]}`}>
      {icon}
      {children}
    </span>
  );
}
