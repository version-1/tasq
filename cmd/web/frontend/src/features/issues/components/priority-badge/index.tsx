import { useTranslation } from "react-i18next";
import { Badge } from "@/components/ui/badge";
import { IconProxy, type IconProxyName } from "@/components/ui/icon-proxy";
import type { Priority } from "@/lib/types";
import styles from "./index.module.css";

const priorityIcons = {
  high: "arrow-up",
  low: "arrow-down",
  normal: null,
  urgent: "arrow-up",
} satisfies Record<Priority, IconProxyName | null>;

export function PriorityBadge({ priority }: { priority: Priority }) {
  const { t } = useTranslation();

  return (
    <Badge variant={priorityBadgeVariant(priority)} icon={<PriorityIcon priority={priority} />}>
      {t(`priorities.${priority}`)}
    </Badge>
  );
}

function PriorityIcon({ priority }: { priority: Priority }) {
  const icon = priorityIcons[priority];
  if (icon === null) {
    return <span aria-hidden="true" className={styles.priorityDot} />;
  }
  return <IconProxy name={icon} size={16} strokeWidth={2.4} />;
}

function priorityBadgeVariant(priority: Priority): "priority-high" | "priority-normal" | "priority-low" {
  if (priority === "high" || priority === "urgent") {
    return "priority-high";
  }
  if (priority === "low") {
    return "priority-low";
  }
  return "priority-normal";
}
