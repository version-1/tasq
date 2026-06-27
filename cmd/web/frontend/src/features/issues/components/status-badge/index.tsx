import { Badge } from "@/components/ui/badge";
import { IconProxy, type IconProxyName } from "@/components/ui/icon-proxy";
import type { IssueStatus } from "@/lib/types";
import styles from "./index.module.css";

const statusIcons = {
  backlog: "circle",
  blocked: "ban",
  cancelled: "x",
  done: "check",
  duplicate: "copy",
  failed: "x",
  in_progress: "play",
  ready: "play",
  review: "eye",
} satisfies Record<IssueStatus, IconProxyName>;

export function StatusBadge({ status }: { status: IssueStatus }) {
  return (
    <Badge
      variant={statusBadgeVariant(status)}
      icon={<IconProxy name={statusIcons[status]} size={13} strokeWidth={2.2} />}
    >
      {status}
    </Badge>
  );
}

export function statusToneClassName(status: IssueStatus): string {
  switch (status) {
    case "backlog":
      return styles.toneBacklog;
    case "ready":
      return styles.toneReady;
    case "in_progress":
      return styles.toneInProgress;
    case "review":
      return styles.toneReview;
    case "done":
      return styles.toneDone;
    case "blocked":
      return styles.toneBlocked;
    case "failed":
      return styles.toneFailed;
    case "cancelled":
      return styles.toneCancelled;
    case "duplicate":
      return styles.toneDuplicate;
  }
}

function statusBadgeVariant(status: IssueStatus):
  | "status-backlog"
  | "status-ready"
  | "status-in-progress"
  | "status-review"
  | "status-done"
  | "status-blocked"
  | "status-failed"
  | "status-muted" {
  switch (status) {
    case "backlog":
      return "status-backlog";
    case "ready":
      return "status-ready";
    case "in_progress":
      return "status-in-progress";
    case "review":
      return "status-review";
    case "done":
      return "status-done";
    case "blocked":
      return "status-blocked";
    case "failed":
      return "status-failed";
    case "cancelled":
    case "duplicate":
      return "status-muted";
  }
}
