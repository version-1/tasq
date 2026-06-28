import { Badge } from "@/components/ui/badge";
import { IconProxy } from "@/components/ui/icon-proxy";
import styles from "./index.module.css";

export function PendingBadge({ className }: { className?: string }) {
  return (
    <Badge
      className={[styles.pendingBadge, className].filter(Boolean).join(" ")}
      variant="status-muted"
      icon={<IconProxy name="pause" size={13} strokeWidth={2.2} />}
    >
      pending
    </Badge>
  );
}
