import { useTranslation } from "react-i18next";
import { Badge } from "@/components/ui/badge";
import type { ChangeRequestStatus } from "@/lib/types";

export const changeRequestStatuses = ["open", "in_progress", "resolved", "canceled"] satisfies ChangeRequestStatus[];

export function ChangeRequestStatusBadge({ status }: { status: ChangeRequestStatus }) {
  const { t } = useTranslation();

  return (
    <Badge variant={changeRequestStatusBadgeVariant(status)}>
      {t(`changeRequests.statuses.${status}`)}
    </Badge>
  );
}

function changeRequestStatusBadgeVariant(status: ChangeRequestStatus):
  | "status-review"
  | "status-in-progress"
  | "status-done"
  | "status-muted" {
  switch (status) {
    case "open":
      return "status-review";
    case "in_progress":
      return "status-in-progress";
    case "resolved":
      return "status-done";
    case "canceled":
      return "status-muted";
  }
}
