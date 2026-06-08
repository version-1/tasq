import type { IssueStatus } from "@/lib/types";

type BoardColumnKey = "draft" | "todo" | "inProgress" | "inReview";

export type BoardColumn = {
  key: BoardColumnKey;
  titleKey: string;
  statuses: IssueStatus[];
};

export const boardColumns: BoardColumn[] = [
  {
    key: "draft",
    titleKey: "issues.board.draft",
    statuses: ["backlog", "blocked", "failed"],
  },
  { key: "todo", titleKey: "issues.board.todo", statuses: ["ready"] },
  {
    key: "inProgress",
    titleKey: "issues.board.inProgress",
    statuses: ["in_progress"],
  },
  { key: "inReview", titleKey: "issues.board.inReview", statuses: ["review"] },
];
