import { Link } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { useEffect, useRef, useState } from "react";
import { IconProxy } from "@/components/ui/icon-proxy";
import type { Issue, IssueStatus, IssueSummary } from "@/lib/types";
import { issueStatuses } from "@/lib/types";
import { PriorityBadge } from "@/features/issues/components/priority-badge";
import { ProjectBadge } from "@/features/issues/components/project-badge";
import { StatusBadge } from "@/features/issues/components/status-badge";
import { formatDateTime } from "../format";
import { MetaItem } from "./meta-item";
import styles from "./index.module.css";

export function BasicInfoPanel({
  disabled,
  issue,
  issueOptions = [],
  onStatusChange,
}: {
  disabled: boolean;
  issue: Issue;
  issueOptions?: IssueSummary[];
  onStatusChange: (status: IssueStatus) => Promise<void>;
}) {
  const { t } = useTranslation();
  const dependencyIssues = dependencyIssueLinks(issue.dependency_ids, issueOptions);

  return (
    <section className={styles.panel}>
      <dl className={styles.metaGrid}>
        <MetaItem label={t("issues.table.id")} value={`#${issue.id}`} />
        <MetaItem label={t("issues.detail.project")} value={<ProjectBadge projectKey={issue.projectKey} />} />
        <MetaItem label={t("issues.detail.priority")} value={<PriorityBadge priority={issue.priority} />} />
        <MetaItem
          label={t("issues.table.status")}
          value={
            <StatusDropdown
              currentStatus={issue.status}
              disabled={disabled}
              onStatusChange={onStatusChange}
            />
          }
        />
        <MetaItem label={t("issues.assignee")} value={issue.assignee || t("issues.unassigned")} />
        <MetaItem
          label={t("issues.detailPage.dependencies")}
          value={
            dependencyIssues.length > 0 ? (
              <div className={styles.dependencyList}>
                {dependencyIssues.map((dependency) => (
                  <Link key={dependency.id} className={styles.dependencyLink} to={`/issues/${dependency.id}`}>
                    #{dependency.id} {dependency.title}
                  </Link>
                ))}
              </div>
            ) : (
              t("issues.detailPage.noDependencies")
            )
          }
        />
        <MetaItem label={t("issues.detailPage.createdAt")} value={formatDateTime(issue.createdAt)} />
        <MetaItem label={t("issues.detailPage.updatedAt")} value={formatDateTime(issue.updatedAt)} />
      </dl>
    </section>
  );
}

function dependencyIssueLinks(
  dependencyIDs: number[],
  issueOptions: IssueSummary[],
): Array<{ id: number; title: string }> {
  return dependencyIDs
    .map((id) => {
      const issue = issueOptions.find((candidate) => candidate.id === id);
      return issue ? { id, title: issue.title } : null;
    })
    .filter((item): item is { id: number; title: string } => item !== null);
}

function StatusDropdown({
  currentStatus,
  disabled,
  onStatusChange,
}: {
  currentStatus: IssueStatus;
  disabled: boolean;
  onStatusChange: (status: IssueStatus) => Promise<void>;
}) {
  const { t } = useTranslation();
  const [isOpen, setIsOpen] = useState(false);
  const [draftStatus, setDraftStatus] = useState<IssueStatus>(currentStatus);
  const rootRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!isOpen) {
      setDraftStatus(currentStatus);
    }
  }, [currentStatus, isOpen]);

  useEffect(() => {
    if (!isOpen) {
      return;
    }

    function handlePointerDown(event: PointerEvent) {
      if (!rootRef.current?.contains(event.target as Node)) {
        setIsOpen(false);
      }
    }

    function handleKeyDown(event: KeyboardEvent) {
      if (event.key === "Escape") {
        setIsOpen(false);
      }
    }

    document.addEventListener("pointerdown", handlePointerDown);
    document.addEventListener("keydown", handleKeyDown);
    return () => {
      document.removeEventListener("pointerdown", handlePointerDown);
      document.removeEventListener("keydown", handleKeyDown);
    };
  }, [isOpen]);

  function handleApply() {
    setIsOpen(false);
    if (draftStatus !== currentStatus) {
      void onStatusChange(draftStatus);
    }
  }

  function handleCancel() {
    setDraftStatus(currentStatus);
    setIsOpen(false);
  }

  return (
    <div ref={rootRef} className={styles.statusDropdown}>
      <button
        type="button"
        className={styles.statusTrigger}
        aria-label={t("issues.detailPage.statusActions")}
        aria-expanded={isOpen}
        disabled={disabled}
        onClick={() => setIsOpen((current) => !current)}
      >
        <StatusBadge status={currentStatus} />
        <IconProxy className={styles.statusTriggerIcon} name="chevron-down" size={16} strokeWidth={2.2} />
      </button>
      {isOpen ? (
        <div className={styles.statusPopover}>
          {issueStatuses.map((status) => (
            <button
              key={status}
              type="button"
              className={status === draftStatus ? styles.statusOptionActive : styles.statusOption}
              disabled={disabled}
              onClick={() => setDraftStatus(status)}
            >
              {t(`statuses.${status}`)}
            </button>
          ))}
          <div className={styles.statusActions}>
            <button type="button" className={styles.statusCancelButton} onClick={handleCancel}>
              {t("issues.table.cancel")}
            </button>
            <button type="button" className={styles.statusApplyButton} disabled={disabled} onClick={handleApply}>
              {t("issues.table.apply")}
            </button>
          </div>
        </div>
      ) : null}
    </div>
  );
}
