"use client";

import { useLayoutData } from "@/components/layout";
import { IssuesView } from "@/features/issues/components/issues-view";

export default function IssuesPage() {
  const { summary, onAddIssue, onStatusChange } = useLayoutData();

  return (
    <IssuesView
      summary={summary}
      onAddIssue={onAddIssue}
      onStatusChange={onStatusChange}
    />
  );
}
