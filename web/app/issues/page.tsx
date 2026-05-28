"use client";

import { useLayoutData } from "@/components/layout";
import { IssuesView } from "./_components/issues-view";

export default function IssuesPage() {
  const { summary, selectedIssue, onSelectIssue, onStatusChange } = useLayoutData();

  return (
    <IssuesView
      summary={summary}
      selectedIssue={selectedIssue}
      onSelectIssue={onSelectIssue}
      onStatusChange={onStatusChange}
    />
  );
}
