import type { IssueStatus } from "@/lib/types";

export type StatusChangeHandler = (id: number, status: IssueStatus) => Promise<void>;
