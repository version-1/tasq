"use client";

import type { IssueStatus, Summary } from "@/lib/types";
import { IssueBoard } from "@/features/issues/components/board";
import type { RejectIssueHandler, RejectShortcutHandler, StatusChangeHandler } from "./types";
import styles from "./index.module.css";

export function IssuesView({
  showFilterSortActions = true,
  summary,
  onAddIssue,
  onRejectIssue,
  onRejectShortcut,
  onStatusChange,
}: {
  showFilterSortActions?: boolean;
  summary: Summary;
  onAddIssue: (status?: IssueStatus) => void;
  onRejectIssue?: RejectIssueHandler;
  onRejectShortcut?: RejectShortcutHandler;
  onStatusChange: StatusChangeHandler;
}) {
  return (
    <div className={styles.issuesLayout}>
      <IssueBoard
        showFilterSortActions={showFilterSortActions}
        summary={summary}
        onAddIssue={onAddIssue}
        onRejectIssue={onRejectIssue}
        onRejectShortcut={onRejectShortcut}
        onStatusChange={onStatusChange}
      />
    </div>
  );
}
