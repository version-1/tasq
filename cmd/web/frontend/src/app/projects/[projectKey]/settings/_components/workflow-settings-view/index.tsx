"use client";

import type { CSSProperties } from "react";
import { useTranslation } from "react-i18next";
import { Markdown } from "@/components/issue/markdown";
import type { ProjectWorkflow, ProjectWorkflowFrontmatter } from "@/lib/types";
import styles from "./index.module.css";

type FrontmatterRow = {
  depth: number;
  id: string;
  kind: "branch" | "leaf";
  key: string;
  value: string;
};

export function WorkflowSettingsView({ workflow }: { workflow: ProjectWorkflow }) {
  const { t } = useTranslation();
  const hasFrontmatter = Object.keys(workflow.frontmatter).length > 0;

  return (
    <section className={styles.layout}>
      <div className={styles.metaPanel}>
        <span className={styles.metaLabel}>{t("projectSettings.syncedAt")}</span>
        <time className={styles.metaValue} dateTime={workflow.updatedAt}>
          {formatDateTime(workflow.updatedAt)}
        </time>
      </div>

      <section className={styles.panel} aria-labelledby="project-workflow-frontmatter">
        <h2 className={styles.heading} id="project-workflow-frontmatter">
          {t("projectSettings.frontmatter")}
        </h2>
        {hasFrontmatter ? (
          <FrontmatterTable rows={toFrontmatterRows(workflow.frontmatter)} />
        ) : (
          <p className={styles.emptyText}>{t("projectSettings.emptyFrontmatter")}</p>
        )}
      </section>

      <section className={styles.panel} aria-labelledby="project-workflow-body">
        <h2 className={styles.heading} id="project-workflow-body">
          {t("projectSettings.body")}
        </h2>
        <Markdown
          className={styles.markdown}
          content={workflow.body}
          emptyText={t("projectSettings.emptyBody")}
        />
      </section>
    </section>
  );
}

function FrontmatterTable({ rows }: { rows: FrontmatterRow[] }) {
  const { t } = useTranslation();

  return (
    <div className={styles.tableWrap}>
      <table className={styles.frontmatterTable}>
        <thead>
          <tr>
            <th className={styles.tableHeadCell} scope="col">
              {t("projectSettings.frontmatterKey")}
            </th>
            <th className={styles.tableHeadCell} scope="col">
              {t("projectSettings.frontmatterValue")}
            </th>
          </tr>
        </thead>
        <tbody>
          {rows.map((row) => (
            <tr key={row.id}>
              <th className={styles.keyCell} scope="row">
                <span
                  className={`${styles.keyContent} ${row.kind === "branch" ? styles.branchKey : ""}`}
                  style={{ "--frontmatter-depth": row.depth } as CSSProperties}
                >
                  {row.kind === "branch" ? <span className={styles.branchMarker}>{">"}</span> : null}
                  <span className={styles.keyText}>{row.key}</span>
                </span>
              </th>
              <td className={styles.valueCell}>{row.value}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function toFrontmatterRows(value: ProjectWorkflowFrontmatter): FrontmatterRow[] {
  return Object.entries(value).flatMap(([key, nestedValue]) =>
    flattenFrontmatterValue({
      depth: 0,
      key,
      path: key,
      value: nestedValue,
    }),
  );
}

function flattenFrontmatterValue({
  depth,
  key,
  path,
  value,
}: {
  depth: number;
  key: string;
  path: string;
  value: unknown;
}): FrontmatterRow[] {
  if (isRecord(value)) {
    const entries = Object.entries(value);
    const row: FrontmatterRow = {
      depth,
      id: path,
      key,
      kind: entries.length > 0 ? "branch" : "leaf",
      value: entries.length > 0 ? "{...}" : "{}",
    };
    return [
      row,
      ...entries.flatMap(([nestedKey, nestedValue]) =>
        flattenFrontmatterValue({
          depth: depth + 1,
          key: nestedKey,
          path: `${path}.${nestedKey}`,
          value: nestedValue,
        }),
      ),
    ];
  }

  if (Array.isArray(value)) {
    const row: FrontmatterRow = {
      depth,
      id: path,
      key,
      kind: value.length > 0 ? "branch" : "leaf",
      value: value.length > 0 ? "[...]" : "[]",
    };
    return [
      row,
      ...value.flatMap((item, index) =>
        flattenFrontmatterValue({
          depth: depth + 1,
          key: `[${index}]`,
          path: `${path}[${index}]`,
          value: item,
        }),
      ),
    ];
  }

  return [{
    depth,
    id: path,
    key,
    kind: "leaf",
    value: formatScalar(value),
  }];
}

function isRecord(value: unknown): value is ProjectWorkflowFrontmatter {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function formatScalar(value: unknown): string {
  if (value === null) {
    return "null";
  }
  if (typeof value === "string") {
    return value;
  }
  if (typeof value === "number" || typeof value === "boolean") {
    return String(value);
  }
  return JSON.stringify(value);
}

function formatDateTime(value: string): string {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return new Intl.DateTimeFormat(undefined, {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(date);
}
