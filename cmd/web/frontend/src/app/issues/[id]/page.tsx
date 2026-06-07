import { Suspense } from "react";
import { IssueDetailPage } from "./_components/issue-detail-page";

export default function IssueDetailRoute() {
  return (
    <Suspense fallback={null}>
      <IssueDetailPage />
    </Suspense>
  );
}
