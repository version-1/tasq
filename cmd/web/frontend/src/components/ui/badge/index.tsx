import type { ReactNode } from "react";
import styles from "./index.module.css";

type BadgeVariant =
  | "project"
  | "priority-high"
  | "priority-normal"
  | "priority-low"
  | "status-backlog"
  | "status-ready"
  | "status-in-progress"
  | "status-review"
  | "status-done"
  | "status-blocked"
  | "status-failed"
  | "status-muted";

export function Badge({
  children,
  className,
  icon,
  variant,
}: {
  children: ReactNode;
  className?: string;
  icon?: ReactNode;
  variant: BadgeVariant;
}) {
  return (
    <span className={[styles.badge, styles[variant], className].filter(Boolean).join(" ")}>
      {icon}
      {children}
    </span>
  );
}
