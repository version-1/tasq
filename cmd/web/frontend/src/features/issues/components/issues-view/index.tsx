"use client";

import type { IssueStatus, Summary } from "@/lib/types";
import { IssueBoard } from "@/features/issues/components/board";
import type { RejectIssueHandler, StatusChangeHandler } from "./types";
import styles from "./index.module.css";

export function IssuesView({
  showFilterSortActions = true,
  summary,
  onAddIssue,
  onRejectIssue,
  onStatusChange,
}: {
  showFilterSortActions?: boolean;
  summary: Summary;
  onAddIssue: (status?: IssueStatus) => void;
  onRejectIssue?: RejectIssueHandler;
  onStatusChange: StatusChangeHandler;
}) {
  return (
    <div className={styles.issuesLayout}>
      <IssueBoard
        showFilterSortActions={showFilterSortActions}
        summary={summary}
        onAddIssue={onAddIssue}
        onRejectIssue={onRejectIssue}
        onStatusChange={onStatusChange}
      />
    </div>
  );
}
