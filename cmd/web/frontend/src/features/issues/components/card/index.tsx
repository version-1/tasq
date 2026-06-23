import { useRef, useState } from "react";
import { Link } from "react-router-dom";
import { useTranslation } from "react-i18next";
import type { IssueStatus, IssueSummary, Priority } from "@/lib/types";
import { Badge } from "@/components/ui/badge";
import {
  ContextMenu,
  ContextMenuGroupLabel,
  ContextMenuHelp,
  ContextMenuItem,
} from "@/components/ui/context-menu";
import { IconProxy, type IconProxyName } from "@/components/ui/icon-proxy";
import styles from "./index.module.css";

type IssueStatusChangeHandler = (id: number, status: IssueStatus) => Promise<void>;

type IssueMetric = {
  icon: IconProxyName;
  label: string;
  value: number;
};

const statusTransitionTargets: Partial<Record<IssueStatus, IssueStatus[]>> = {
  backlog: ["cancelled", "done", "duplicate"],
  blocked: ["cancelled", "done", "duplicate"],
  ready: ["backlog", "cancelled", "done", "duplicate"],
  review: ["backlog", "cancelled", "done", "duplicate"],
};

const quickStatusTargets: Partial<Record<IssueStatus, IssueStatus>> = {
  backlog: "ready",
  blocked: "ready",
  review: "done",
};

const priorityIcons = {
  high: "arrow-up",
  low: "arrow-down",
  normal: null,
  urgent: "arrow-up",
} satisfies Record<Priority, IconProxyName | null>;

export function IssueCard({
  issue,
  onStatusChange,
  readonly = false,
  runCount,
}: {
  issue: IssueSummary;
  onStatusChange: IssueStatusChangeHandler;
  readonly?: boolean;
  runCount?: number;
}) {
  const { t } = useTranslation();
  const statusOptions = statusOptionsFor(issue.status);
  const canChangeStatus = !readonly && statusOptions.length > 1;
  const quickStatusTarget = readonly ? undefined : quickStatusTargets[issue.status];
  const cardRef = useRef<HTMLElement>(null);
  const [isMenuOpen, setIsMenuOpen] = useState(false);
  const menuID = `issue-card-menu-${issue.id}`;
  const menuLabel = t("issues.card.actionsLabel", { title: issue.title });
  const lockedStatusLabel = t("issues.card.statusLocked", { status: t(`statuses.${issue.status}`) });
  const metrics: IssueMetric[] = [
    {
      icon: "message-square",
      label: t("issues.card.commentCount", { count: issue.stats.commentCount }),
      value: issue.stats.commentCount,
    },
    {
      icon: "history",
      label: t("issues.card.runCount", { count: runCount ?? 0 }),
      value: runCount ?? 0,
    },
  ];

  return (
    <article className={styles.taskCard} ref={cardRef}>
      <div className={styles.titleRow}>
        <Link className={styles.taskTitle} to={`/issues/${issue.id}`}>
          #{issue.id} {issue.title}
        </Link>
        <ContextMenu
          boundaryRef={cardRef}
          id={menuID}
          isOpen={isMenuOpen}
          label={menuLabel}
          onOpenChange={setIsMenuOpen}
          trigger={(triggerProps) => (
            <button
              {...triggerProps}
              aria-label={menuLabel}
              className={styles.menuButton}
              title={menuLabel}
              type="button"
            >
              <IconProxy name="ellipsis" size={16} />
            </button>
          )}
        >
          <ContextMenuGroupLabel>{t("issues.card.changeStatus")}</ContextMenuGroupLabel>
          {statusOptions.map((status) => {
            const isCurrent = status === issue.status;
            const isDisabled = isCurrent || !canChangeStatus;
            const itemLabel = isCurrent
              ? t("issues.card.currentStatus", { status: t(`statuses.${status}`) })
              : t(`statuses.${status}`);

            return (
              <ContextMenuItem
                key={status}
                disabled={isDisabled}
                label={isDisabled && !isCurrent ? lockedStatusLabel : itemLabel}
                title={isDisabled && !isCurrent ? lockedStatusLabel : itemLabel}
                onSelect={() => {
                  setIsMenuOpen(false);
                  void onStatusChange(issue.id, status);
                }}
              >
                {itemLabel}
              </ContextMenuItem>
            );
          })}
          {!canChangeStatus && statusOptions.length === 1 ? (
            <ContextMenuHelp>{lockedStatusLabel}</ContextMenuHelp>
          ) : null}
        </ContextMenu>
      </div>

      <div className={styles.metaRow}>
        <Badge
          variant="project"
          icon={<IconProxy name="folder" size={17} strokeWidth={2.3} />}
        >
          {issue.projectKey}
        </Badge>
        <Badge variant={priorityBadgeVariant(issue.priority)} icon={<PriorityIcon priority={issue.priority} />}>
          {t(`priorities.${issue.priority}`)}
        </Badge>
      </div>

      <div className={styles.footerRow}>
        <div className={styles.metrics}>
          {metrics.map((metric) => (
            <span
              key={metric.icon}
              aria-label={metric.label}
              className={styles.metric}
              title={metric.label}
            >
              <IconProxy name={metric.icon} size={14} />
              {metric.value}
            </span>
          ))}
        </div>
        {quickStatusTarget ? (
          <button
            className={quickActionClassName(quickStatusTarget)}
            type="button"
            onClick={() => {
              void onStatusChange(issue.id, quickStatusTarget);
            }}
          >
            {t(`statuses.${quickStatusTarget}`)}
            {quickActionShowsIcon(quickStatusTarget) ? (
              <IconProxy className={styles.quickActionIcon} name="arrow-right" size={14} strokeWidth={2.4} />
            ) : null}
          </button>
        ) : null}
      </div>
    </article>
  );
}

function statusOptionsFor(status: IssueStatus): IssueStatus[] {
  return [status, ...(statusTransitionTargets[status] ?? [])];
}

function PriorityIcon({ priority }: { priority: Priority }) {
  const icon = priorityIcons[priority];
  if (icon === null) {
    return <span aria-hidden="true" className={styles.priorityDot} />;
  }
  return <IconProxy name={icon} size={16} strokeWidth={2.4} />;
}

function priorityBadgeVariant(priority: IssueSummary["priority"]): "priority-high" | "priority-normal" | "priority-low" {
  if (priority === "high" || priority === "urgent") {
    return "priority-high";
  }
  if (priority === "low") {
    return "priority-low";
  }
  return "priority-normal";
}

function quickActionClassName(status: IssueStatus): string {
  return `${styles.quickActionButton} ${styles[`quickAction-${status}`]}`;
}

function quickActionShowsIcon(status: IssueStatus): boolean {
  return status === "ready";
}
