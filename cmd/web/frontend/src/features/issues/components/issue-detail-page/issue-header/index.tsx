import { useTranslation } from "react-i18next";
import { useEffect, useRef, useState } from "react";
import { IconProxy } from "@/components/ui/icon-proxy";
import type { Issue, IssueStatus } from "@/lib/types";
import { issueStatuses } from "@/lib/types";
import { PriorityBadge } from "@/features/issues/components/priority-badge";
import { StatusBadge } from "@/features/issues/components/status-badge";
import { formatDateTime } from "../format";
import { MetaItem } from "./meta-item";
import styles from "./index.module.css";

export function IssueHeader({
  disabled,
  issue,
  onStatusChange,
}: {
  disabled: boolean;
  issue: Issue;
  onStatusChange: (status: IssueStatus) => Promise<void>;
}) {
  const { t } = useTranslation();

  return (
    <header className={styles.header}>
      <dl className={styles.metaGrid}>
        <MetaItem label={t("issues.table.id")} value={`#${issue.id}`} />
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
        <MetaItem label={t("issues.detail.priority")} value={<PriorityBadge priority={issue.priority} />} />
        <MetaItem label={t("issues.detail.project")} value={issue.projectKey} />
        <MetaItem label={t("issues.assignee")} value={issue.assignee || t("issues.unassigned")} />
        <MetaItem label={t("issues.detailPage.createdAt")} value={formatDateTime(issue.createdAt)} />
        <MetaItem label={t("issues.detailPage.updatedAt")} value={formatDateTime(issue.updatedAt)} />
      </dl>
    </header>
  );
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
  const rootRef = useRef<HTMLDivElement>(null);

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

  function handleSelect(status: IssueStatus) {
    setIsOpen(false);
    if (status !== currentStatus) {
      void onStatusChange(status);
    }
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
              className={status === currentStatus ? styles.statusOptionActive : styles.statusOption}
              disabled={disabled}
              onClick={() => handleSelect(status)}
            >
              <StatusBadge status={status} />
            </button>
          ))}
        </div>
      ) : null}
    </div>
  );
}
