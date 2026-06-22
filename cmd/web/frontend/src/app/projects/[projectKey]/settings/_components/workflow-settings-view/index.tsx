"use client";

import { useTranslation } from "react-i18next";
import { Markdown } from "@/components/issue/markdown";
import type { ProjectWorkflow, ProjectWorkflowFrontmatter } from "@/lib/types";
import styles from "./index.module.css";

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
          <FrontmatterTree value={workflow.frontmatter} />
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

function FrontmatterTree({ value }: { value: ProjectWorkflowFrontmatter }) {
  return (
    <dl className={styles.tree}>
      {Object.entries(value).map(([key, nestedValue]) => (
        <FrontmatterEntry key={key} name={key} value={nestedValue} />
      ))}
    </dl>
  );
}

function FrontmatterEntry({ name, value }: { name: string; value: unknown }) {
  if (isRecord(value)) {
    const entries = Object.entries(value);
    return (
      <div className={styles.branch}>
        <dt className={styles.key}>{name}</dt>
        <dd className={styles.value}>
          {entries.length > 0 ? (
            <dl className={styles.tree}>
              {entries.map(([nestedKey, nestedValue]) => (
                <FrontmatterEntry key={nestedKey} name={nestedKey} value={nestedValue} />
              ))}
            </dl>
          ) : (
            <span className={styles.scalar}>{"{}"}</span>
          )}
        </dd>
      </div>
    );
  }

  if (Array.isArray(value)) {
    return (
      <div className={styles.branch}>
        <dt className={styles.key}>{name}</dt>
        <dd className={styles.value}>
          {value.length > 0 ? (
            <ol className={styles.array}>
              {value.map((item, index) => (
                <li className={styles.arrayItem} key={index}>
                  {isRecord(item) ? (
                    <FrontmatterTree value={item} />
                  ) : (
                    formatScalar(item)
                  )}
                </li>
              ))}
            </ol>
          ) : (
            <span className={styles.scalar}>[]</span>
          )}
        </dd>
      </div>
    );
  }

  return (
    <div className={styles.leaf}>
      <dt className={styles.key}>{name}</dt>
      <dd className={styles.scalar}>{formatScalar(value)}</dd>
    </div>
  );
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
