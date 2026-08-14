"use client";

import { useLayoutData, useLayoutShellData } from "@/components/layout";
import { IssuesView } from "@/features/issues/components/issues-view";

export default function IssuesPage() {
  const { summary, onAddIssue, onRejectIssue, onRejectShortcut, onResolveIssue, onStatusChange } = useLayoutData();
  const { isProjectIssueScope } = useLayoutShellData();

  return (
    <IssuesView
      showFilterSortActions={!isProjectIssueScope}
      summary={summary}
      onAddIssue={onAddIssue}
      onRejectIssue={onRejectIssue}
      onRejectShortcut={onRejectShortcut}
      onResolveIssue={onResolveIssue}
      onStatusChange={onStatusChange}
    />
  );
}
