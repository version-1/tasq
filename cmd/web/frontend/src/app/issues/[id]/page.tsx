import { Suspense } from "react";
import { IssueDetailPage } from "@/features/issues/components/issue-detail-page";

export default function IssueDetailRoute() {
  return (
    <Suspense fallback={null}>
      <IssueDetailPage />
    </Suspense>
  );
}
