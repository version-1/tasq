"use client";

import type { IssueStatus, Summary } from "@/lib/types";
import { IssueBoard } from "./issue-board";
import type { StatusChangeHandler } from "./types";
import styles from "./index.module.css";

export function IssuesView({
  summary,
  onAddIssue,
  onStatusChange,
}: {
  summary: Summary;
  onAddIssue: (status?: IssueStatus) => void;
  onStatusChange: StatusChangeHandler;
}) {
  return (
    <div className={styles.issuesLayout}>
      <IssueBoard
        summary={summary}
        onAddIssue={onAddIssue}
        onStatusChange={onStatusChange}
      />
    </div>
  );
}
