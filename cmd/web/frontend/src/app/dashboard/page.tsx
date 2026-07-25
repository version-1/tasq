"use client";

import { useLayoutData } from "@/components/layout";
import { IssuesView } from "@/features/issues/components/issues-view";

export default function DashboardPage() {
  const { summary, onAddIssue, onRejectIssue, onStatusChange } = useLayoutData();

  return (
    <IssuesView
      showFilterSortActions={false}
      summary={summary}
      onAddIssue={onAddIssue}
      onRejectIssue={onRejectIssue}
      onStatusChange={onStatusChange}
    />
  );
}
