import type { IssueStatus } from "@/lib/generated/issue-tracker";

export const summaryColumnFixtures: Array<{ status: IssueStatus; title: string }> = [
  { status: "backlog", title: "Backlog" },
  { status: "ready", title: "Ready" },
  { status: "in_progress", title: "In Progress" },
  { status: "review", title: "Review" },
  { status: "done", title: "Done" },
  { status: "blocked", title: "Blocked" },
  { status: "failed", title: "Failed" },
  { status: "cancelled", title: "Cancelled" },
  { status: "duplicate", title: "Duplicate" },
];
