import type { IssueStatus, Priority } from "@/lib/types";

export type AddIssueFormValues = {
  title: string;
  description: string;
  projectID: number | null;
  status: IssueStatus;
  priority: Priority;
  assignee: string;
  dependencyIDs: number[];
};
