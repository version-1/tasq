import type { ComponentPropsWithoutRef, ReactNode } from "react";
import { Link } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { IconProxy } from "@/components/ui/icon-proxy";
import { PriorityBadge } from "@/features/issues/components/priority-badge";
import { ProjectBadge } from "@/features/issues/components/project-badge";
import { StatusBadge, statusToneClassName } from "@/features/issues/components/status-badge";
import type { Issue } from "@/lib/types";
import styles from "./index.module.css";

type TableProps = ComponentPropsWithoutRef<"table"> & {
  minWidth?: string;
};

export type IssueTableSortBy = "id" | "priority" | "created_at" | "updated_at";
export type IssueTableSortDirection = "asc" | "desc";

export function Table({ children, className, minWidth = "920px", style, ...props }: TableProps) {
  return (
    <div className={styles.tableWrap}>
      <table
        className={[styles.table, className].filter(Boolean).join(" ")}
        style={{ ...style, minInlineSize: minWidth }}
        {...props}
      >
        {children}
      </table>
    </div>
  );
}

export function TableHeader({ children }: { children: ReactNode }) {
  return <thead>{children}</thead>;
}

export function TableBody({ children }: { children: ReactNode }) {
  return <tbody>{children}</tbody>;
}

export function TableRow(props: ComponentPropsWithoutRef<"tr">) {
  return <tr {...props} />;
}

export function TableHeadCell({ className, ...props }: ComponentPropsWithoutRef<"th">) {
  return <th className={[styles.headCell, className].filter(Boolean).join(" ")} {...props} />;
}

export function TableCell({ className, ...props }: ComponentPropsWithoutRef<"td">) {
  return <td className={[styles.cell, className].filter(Boolean).join(" ")} {...props} />;
}

export function IssueTable({
  issues,
  onSortChange,
  sortBy,
  sortDirection,
}: {
  issues: Issue[];
  onSortChange: (sortBy: IssueTableSortBy) => void;
  sortBy: IssueTableSortBy;
  sortDirection: IssueTableSortDirection;
}) {
  const { t } = useTranslation();

  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHeadCell>
            <SortButton
              active={sortBy === "id"}
              direction={sortDirection}
              label={t("issues.table.id")}
              onClick={() => onSortChange("id")}
            />
          </TableHeadCell>
          <TableHeadCell>{t("issues.table.title")}</TableHeadCell>
          <TableHeadCell>{t("issues.table.status")}</TableHeadCell>
          <TableHeadCell>{t("issues.table.assignee")}</TableHeadCell>
          <TableHeadCell>{t("issues.table.project")}</TableHeadCell>
          <TableHeadCell>
            <SortButton
              active={sortBy === "priority"}
              direction={sortDirection}
              label={t("issues.table.priority")}
              onClick={() => onSortChange("priority")}
            />
          </TableHeadCell>
          <TableHeadCell>
            <SortButton
              active={sortBy === "created_at"}
              direction={sortDirection}
              label={t("issues.table.createdAt")}
              onClick={() => onSortChange("created_at")}
            />
          </TableHeadCell>
          <TableHeadCell>
            <SortButton
              active={sortBy === "updated_at"}
              direction={sortDirection}
              label={t("issues.table.updatedAt")}
              onClick={() => onSortChange("updated_at")}
            />
          </TableHeadCell>
        </TableRow>
      </TableHeader>
      <TableBody>
        {issues.map((issue) => (
          <TableRow key={issue.id} className={`${styles.issueRow} ${statusToneClassName(issue.status)}`}>
            <TableCell className={styles.idCell}>#{issue.id}</TableCell>
            <TableCell>
              <Link className={styles.issueLink} to={`/issues/${issue.id}`}>
                {issue.title}
              </Link>
            </TableCell>
            <TableCell><StatusBadge status={issue.status} /></TableCell>
            <TableCell>{issue.assignee || t("issues.unassigned")}</TableCell>
            <TableCell><ProjectBadge projectKey={issue.projectKey} /></TableCell>
            <TableCell><PriorityBadge priority={issue.priority} /></TableCell>
            <TableCell>{formatDate(issue.createdAt)}</TableCell>
            <TableCell>{formatDate(issue.updatedAt)}</TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}

function SortButton({
  active,
  direction,
  label,
  onClick,
}: {
  active: boolean;
  direction: IssueTableSortDirection;
  label: string;
  onClick: () => void;
}) {
  const { t } = useTranslation();
  const icon = active
    ? direction === "asc" ? "arrow-up" : "arrow-down"
    : "arrow-up-down";

  return (
    <button
      type="button"
      className={styles.sortButton}
      aria-label={t("issues.table.sortBy", { field: label })}
      onClick={onClick}
    >
      <span>{label}</span>
      <IconProxy
        className={active ? styles.sortIconActive : styles.sortIcon}
        name={icon}
        size={14}
        strokeWidth={2.3}
      />
    </button>
  );
}

function formatDate(value: string): string {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return new Intl.DateTimeFormat(undefined, {
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    month: "short",
    year: "numeric",
  }).format(date);
}
