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
import { StatusBadge } from "@/features/issues/components/status-badge";
import { RejectAction } from "@/features/issues/components/reject-action";
import type { ChangeRequestShortcut } from "@/features/issues/change-request-shortcuts";
import { useIssueThreadID } from "@/features/issues/hooks/use-issue-thread-id";
import { pullRequestArtifact } from "@/features/issues/artifacts";
import { PendingBadge } from "./pending-badge";
import styles from "./index.module.css";

type IssueStatusChangeHandler = (id: number, status: IssueStatus) => Promise<void>;
type IssueRejectHandler = (id: number) => void;
type IssueResolveHandler = (id: number) => void;
type IssueRejectShortcutHandler = (id: number, shortcut: ChangeRequestShortcut) => Promise<void>;

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
  review: "done",
};

export function IssueCard({
  issue,
  onStatusChange,
  onRejectIssue,
  onRejectShortcut,
  onResolveIssue,
  readonly = false,
  runCount,
}: {
  issue: IssueSummary;
  onStatusChange: IssueStatusChangeHandler;
  onRejectIssue?: IssueRejectHandler;
  onRejectShortcut?: IssueRejectShortcutHandler;
  onResolveIssue?: IssueResolveHandler;
  readonly?: boolean;
  runCount?: number;
}) {
  const { t } = useTranslation();
  const statusOptions = statusOptionsFor(issue.status);
  const canChangeStatus = !readonly && statusOptions.length > 1;
  const quickStatusTarget = readonly ? undefined : quickStatusTargets[issue.status];
  const canResolve = !readonly && issue.status === "blocked" && onResolveIssue !== undefined;
  const canReject = !readonly && issue.status === "review" && onRejectIssue !== undefined && onRejectShortcut !== undefined;
  const cardRef = useRef<HTMLElement>(null);
  const [isMenuOpen, setIsMenuOpen] = useState(false);
  const { isThreadIDLoading, threadID } = useIssueThreadID(issue.id, isMenuOpen);
  const menuID = `issue-card-menu-${issue.id}`;
  const menuLabel = t("issues.card.actionsLabel", { title: issue.title });
  const lockedStatusLabel = t("issues.card.statusLocked", { status: t(`statuses.${issue.status}`) });
  const pullRequest = pullRequestArtifact(issue.artifacts);
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

  function handleOpenPullRequest() {
    if (!pullRequest) {
      return;
    }

    setIsMenuOpen(false);
    window.open(pullRequest.data_value, "_blank", "noopener,noreferrer");
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
            icon={<IconProxy name="copy" />}
            label={t("issues.card.copyThreadID")}
            title={t("issues.card.copyThreadID")}
            onSelect={() => {
              void handleCopyThreadID();
            }}
          >
            {t("issues.card.copyThreadID")}
          </ContextMenuItem>
          {pullRequest ? (
            <ContextMenuItem
              icon={<IconProxy name="git-pull-request" />}
              label={t("issues.card.openPullRequest")}
              title={t("issues.card.openPullRequest")}
              onSelect={handleOpenPullRequest}
            >
              {t("issues.card.openPullRequest")}
            </ContextMenuItem>
          ) : null}
          <ContextMenuSeparator />
          <ContextMenuGroupLabel>{t("issues.card.changeStatus")}</ContextMenuGroupLabel>
          {statusOptions.map((status, index) => {
            const isCurrent = status === issue.status;
            const isDisabled = isCurrent || !canChangeStatus;
            const itemLabel = isCurrent
              ? t("issues.card.currentStatus", { status: t(`statuses.${status}`) })
              : t(`statuses.${status}`);

            return (
              <Fragment key={status}>
                {status === "duplicate" && index > 0 ? <ContextMenuSeparator /> : null}
                <ContextMenuItem
                  disabled={isDisabled}
                  icon={
                    <IconProxy
                      className={statusIconClassName(status)}
                      name={statusIconName(status)}
                    />
                  }
                  label={isDisabled && !isCurrent ? lockedStatusLabel : itemLabel}
                  title={isDisabled && !isCurrent ? lockedStatusLabel : itemLabel}
                  trailingIcon={isCurrent ? <IconProxy name="check" /> : undefined}
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
        {quickStatusTarget || canReject || canResolve ? (
          <div className={styles.actionGroup}>
            {canReject ? (
              <RejectAction
                onOpenDialog={() => onRejectIssue(issue.id)}
                onSelectShortcut={(shortcut) => onRejectShortcut(issue.id, shortcut)}
              />
            ) : null}
            {canResolve ? (
              <button
                className={quickActionClassName("ready")}
                type="button"
                onClick={() => onResolveIssue(issue.id)}
              >
                {t("issues.resolve.action")}
                <IconProxy className={styles.quickActionIcon} name="arrow-right" size={14} strokeWidth={2.4} />
              </button>
            ) : quickStatusTarget ? (
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

const statusIcons = {
  backlog: "inbox",
  blocked: "circle-alert",
  cancelled: "ban",
  done: "circle-check",
  duplicate: "copy",
  failed: "circle-x",
  in_progress: "play",
  ready: "circle-play",
  review: "eye",
} satisfies Record<IssueStatus, IconProxyName>;

function statusIconName(status: IssueStatus): IconProxyName {
  return statusIcons[status];
}

function statusIconClassName(status: IssueStatus): string {
  if (status === "ready" || status === "done") {
    return styles.statusIconPositive;
  }
  if (status === "cancelled" || status === "failed") {
    return styles.statusIconDanger;
  }
  if (status === "blocked" || status === "review") {
    return styles.statusIconWarning;
  }
  return styles.statusIconMuted;
}

function quickActionClassName(status: IssueStatus): string {
  return `${styles.quickActionButton} ${styles[`quickAction-${status}`]}`;
}

function quickActionShowsIcon(status: IssueStatus): boolean {
  return status === "ready";
}
