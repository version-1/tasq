import { Fragment, useRef, useState } from "react";
import { Link } from "react-router-dom";
import { useTranslation } from "react-i18next";
import type { IssueStatus, IssueSummary } from "@/lib/types";
import {
  ContextMenu,
  ContextMenuGroupLabel,
  ContextMenuHelp,
  ContextMenuItem,
  ContextMenuSeparator,
} from "@/components/ui/context-menu";
import { IconProxy, type IconProxyName } from "@/components/ui/icon-proxy";
import { toast } from "@/lib/toast";
import { PriorityBadge } from "@/features/issues/components/priority-badge";
import { ProjectBadge } from "@/features/issues/components/project-badge";
import {
  StatusBadge,
  statusToneClassName,
} from "@/features/issues/components/status-badge";
import { useIssueThreadID } from "@/features/issues/hooks/use-issue-thread-id";
import { PendingBadge } from "./pending-badge";
import styles from "./index.module.css";

type IssueStatusChangeHandler = (id: number, status: IssueStatus) => Promise<void>;
type IssueRejectHandler = (id: number) => void;

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

const statusMenuIcons = {
  backlog: "circle",
  blocked: "ban",
  cancelled: "ban",
  done: "check",
  duplicate: "copy",
  failed: "x",
  in_progress: "play",
  ready: "play",
  review: "eye",
} satisfies Record<IssueStatus, IconProxyName>;

export function IssueCard({
  issue,
  onStatusChange,
  onRejectIssue,
  readonly = false,
  runCount,
}: {
  issue: IssueSummary;
  onStatusChange: IssueStatusChangeHandler;
  onRejectIssue?: IssueRejectHandler;
  readonly?: boolean;
  runCount?: number;
}) {
  const { t } = useTranslation();
  const statusOptions = statusOptionsFor(issue.status);
  const canChangeStatus = !readonly && statusOptions.length > 1;
  const quickStatusTarget = readonly ? undefined : quickStatusTargets[issue.status];
  const canReject = !readonly && issue.status === "review" && onRejectIssue !== undefined;
  const cardRef = useRef<HTMLElement>(null);
  const [isMenuOpen, setIsMenuOpen] = useState(false);
  const { isThreadIDLoading, threadID } = useIssueThreadID(issue.id, isMenuOpen);
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

  function handleMenuOpenChange(nextIsOpen: boolean) {
    setIsMenuOpen(nextIsOpen);
  }

  async function handleCopyThreadID() {
    if (!threadID) {
      return;
    }

    setIsMenuOpen(false);
    try {
      await navigator.clipboard.writeText(threadID);
      toast.success({ message: t("toast.success.threadIDCopied") });
    } catch {
      toast.error({ message: t("toast.error.clipboardUnavailable") });
    }
  }

  return (
    <article className={styles.taskCard} ref={cardRef}>
      <div className={styles.cardHeader}>
        <ProjectBadge projectKey={issue.projectKey} size="small" />
        <ContextMenu
          boundaryRef={cardRef}
          id={menuID}
          isOpen={isMenuOpen}
          label={menuLabel}
          onOpenChange={handleMenuOpenChange}
          size="wide"
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
          <ContextMenuItem
            disabled={isThreadIDLoading || !threadID}
            icon={<IconProxy name="clipboard" size={18} />}
            label={t("issues.card.copyThreadID")}
            title={t("issues.card.copyThreadID")}
            onSelect={() => {
              void handleCopyThreadID();
            }}
          >
            {t("issues.card.copyThreadID")}
          </ContextMenuItem>
          <ContextMenuSeparator />
          <ContextMenuGroupLabel>{t("issues.card.changeStatus")}</ContextMenuGroupLabel>
          {statusOptions.map((status) => {
            const isCurrent = status === issue.status;
            const isDisabled = isCurrent || !canChangeStatus;
            const itemLabel = isCurrent
              ? t("issues.card.currentStatus", { status: t(`statuses.${status}`) })
              : t(`statuses.${status}`);

            return (
              <Fragment key={status}>
                {status === "duplicate" ? <ContextMenuSeparator /> : null}
                <ContextMenuItem
                  accessory={isCurrent ? <IconProxy name="check" size={18} /> : undefined}
                  disabled={isDisabled}
                  icon={
                    <IconProxy
                      className={[
                        styles.statusMenuIcon,
                        statusToneClassName(status),
                        status === "cancelled" ? styles.cancelledStatusMenuIcon : "",
                      ].filter(Boolean).join(" ")}
                      name={statusMenuIcons[status]}
                      size={18}
                    />
                  }
                  label={isDisabled && !isCurrent ? lockedStatusLabel : itemLabel}
                  selected={isCurrent}
                  title={isDisabled && !isCurrent ? lockedStatusLabel : itemLabel}
                  onSelect={() => {
                    setIsMenuOpen(false);
                    void onStatusChange(issue.id, status);
                  }}
                >
                  {itemLabel}
                </ContextMenuItem>
              </Fragment>
            );
          })}
          {!canChangeStatus && statusOptions.length === 1 ? (
            <ContextMenuHelp>{lockedStatusLabel}</ContextMenuHelp>
          ) : null}
        </ContextMenu>
      </div>

      <div className={styles.titleRow}>
        <Link className={styles.taskTitle} to={`/issues/${issue.id}`}>
          #{issue.id} {issue.title}
        </Link>
      </div>

      <div className={styles.metaRow}>
        <div className={styles.statusGroup}>
          <StatusBadge status={issue.status} />
          <PriorityBadge priority={issue.priority} />
        </div>
        {issue.queueStatus === "pending" ? <PendingBadge className={styles.pendingBadge} /> : null}
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
        {quickStatusTarget || canReject ? (
          <div className={styles.actionGroup}>
            {canReject ? (
              <button
                className={styles.rejectActionButton}
                type="button"
                onClick={() => onRejectIssue(issue.id)}
              >
                {t("issues.reject.action")}
              </button>
            ) : null}
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
        ) : null}
      </div>
    </article>
  );
}

function statusOptionsFor(status: IssueStatus): IssueStatus[] {
  return [status, ...(statusTransitionTargets[status] ?? [])];
}

function quickActionClassName(status: IssueStatus): string {
  return `${styles.quickActionButton} ${styles[`quickAction-${status}`]}`;
}

function quickActionShowsIcon(status: IssueStatus): boolean {
  return status === "ready";
}
