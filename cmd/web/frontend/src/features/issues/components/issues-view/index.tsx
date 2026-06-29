"use client";

import type { IssueStatus, Summary } from "@/lib/types";
import { IssueBoard } from "@/features/issues/components/board";
import type { StatusChangeHandler } from "./types";
import styles from "./index.module.css";

export function IssuesView({
  showFilterSortActions = true,
  summary,
  onAddIssue,
  onStatusChange,
}: {
  showFilterSortActions?: boolean;
  summary: Summary;
  onAddIssue: (status?: IssueStatus) => void;
  onStatusChange: StatusChangeHandler;
}) {
  return (
    <div className={styles.issuesLayout}>
      <IssueBoard
        showFilterSortActions={showFilterSortActions}
        summary={summary}
        onAddIssue={onAddIssue}
        onStatusChange={onStatusChange}
      />
    </div>
  );
}
