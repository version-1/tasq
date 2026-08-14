import type { IssueStatus } from "@/lib/types";
import type { ChangeRequestShortcut } from "@/features/issues/change-request-shortcuts";

export type StatusChangeHandler = (id: number, status: IssueStatus) => Promise<void>;
export type RejectIssueHandler = (id: number) => void;
export type ResolveIssueHandler = (id: number) => void;
export type RejectShortcutHandler = (id: number, shortcut: ChangeRequestShortcut) => Promise<void>;
