"use client";

import { useTranslation } from "react-i18next";
import type { IssueStatus, IssueSummary, Summary } from "@/lib/types";
import { IssueDetail } from "@/components/issue/pane";
import { IssueBoard } from "./issue-board";
import { PanelMessage } from "./panel-message";
import type { StatusChangeHandler } from "./types";
import styles from "./index.module.css";

export function IssuesView({
  summary,
  selectedIssue,
  onSelectIssue,
  onAddIssue,
  onStatusChange,
}: {
  summary: Summary;
  selectedIssue: IssueSummary | null;
  onSelectIssue: (issueID: number) => void;
  onAddIssue: (status?: IssueStatus) => void;
  onStatusChange: StatusChangeHandler;
}) {
  const { t } = useTranslation();

  return (
    <div className={styles.issuesLayout}>
      <IssueBoard
        summary={summary}
        onSelectIssue={onSelectIssue}
        onAddIssue={onAddIssue}
        onStatusChange={onStatusChange}
      />
      {selectedIssue ? (
        <IssueDetail issue={selectedIssue} onStatusChange={onStatusChange} />
      ) : (
        <PanelMessage title={t("issues.noIssueSelected")} />
      )}
    </div>
  );
}
