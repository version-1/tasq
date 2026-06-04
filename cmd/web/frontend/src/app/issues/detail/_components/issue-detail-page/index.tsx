"use client";

import { Link } from "react-router-dom";
import { useSearchParams } from "react-router-dom";
import { useCallback, useEffect, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { fetchComments, fetchIssue, updateIssueStatus } from "@/lib/api";
import { issueStatuses } from "@/lib/types";
import type { Comment, Issue, IssueStatus } from "@/lib/types";
import { Markdown } from "../../../_components/markdown";
import styles from "./index.module.css";

const commentPageSize = 20;

type IssueLoadState =
  | { kind: "loading" }
  | { kind: "ready"; issue: Issue }
  | { kind: "error"; message: string };

export function IssueDetailPage() {
  const { t } = useTranslation();
  const [searchParams] = useSearchParams();
  const issueID = parseIssueID(searchParams.get("id"));
  const [issueState, setIssueState] = useState<IssueLoadState>({ kind: "loading" });
  const [comments, setComments] = useState<Comment[]>([]);
  const [nextCursor, setNextCursor] = useState<number | null>(null);
  const [commentsError, setCommentsError] = useState("");
  const [isLoadingComments, setIsLoadingComments] = useState(false);
  const [isUpdatingStatus, setIsUpdatingStatus] = useState(false);

  const loadIssue = useCallback(async () => {
    if (!issueID) {
      setIssueState({ kind: "error", message: t("issues.detailPage.invalidIssue") });
      return;
    }

    setIssueState({ kind: "loading" });
    try {
      const issue = await fetchIssue(issueID);
      setIssueState({ kind: "ready", issue });
    } catch (error) {
      setIssueState({
        kind: "error",
        message: error instanceof Error ? error.message : t("issues.detailPage.failedToLoadIssue"),
      });
    }
  }, [issueID, t]);

  const loadComments = useCallback(
    async (cursor?: number) => {
      if (!issueID) return;

      setIsLoadingComments(true);
      setCommentsError("");
      try {
        const page = await fetchComments(issueID, cursor, commentPageSize);
        setComments((current) => (cursor ? [...current, ...page.data] : page.data));
        setNextCursor(page.meta.nextCursor);
      } catch (error) {
        setCommentsError(
          error instanceof Error ? error.message : t("issues.detailPage.failedToLoadComments"),
        );
      } finally {
        setIsLoadingComments(false);
      }
    },
    [issueID, t],
  );

  useEffect(() => {
    void loadIssue();
  }, [loadIssue]);

  useEffect(() => {
    setComments([]);
    setNextCursor(null);
    setCommentsError("");
    void loadComments();
  }, [loadComments]);

  async function handleStatusChange(status: IssueStatus) {
    if (issueState.kind !== "ready" || issueState.issue.status === status) return;

    setIsUpdatingStatus(true);
    try {
      const issue = await updateIssueStatus(issueState.issue.id, status);
      setIssueState({ kind: "ready", issue });
    } catch (error) {
      setIssueState({
        kind: "error",
        message: error instanceof Error ? error.message : t("layout.failedToUpdateIssue"),
      });
    } finally {
      setIsUpdatingStatus(false);
    }
  }

  const sortedComments = useMemo(() => {
    return [...comments].sort((left, right) => left.id - right.id);
  }, [comments]);

  return (
    <div className={styles.page}>
      <Link className={styles.backLink} to="/issues">
        {t("issues.detailPage.backToList")}
      </Link>

      {issueState.kind === "loading" ? (
        <PanelMessage title={t("layout.loading")} />
      ) : null}

      {issueState.kind === "error" ? (
        <PanelMessage title={t("issues.detailPage.failedTitle")} detail={issueState.message} />
      ) : null}

      {issueState.kind === "ready" ? (
        <>
          <IssueHeader issue={issueState.issue} />
          <StatusActions
            currentStatus={issueState.issue.status}
            disabled={isUpdatingStatus}
            onStatusChange={handleStatusChange}
          />
          <IssueDescription issue={issueState.issue} />
          <CommentList
            comments={sortedComments}
            error={commentsError}
            hasMore={nextCursor !== null}
            isLoading={isLoadingComments}
            onLoadMore={() => void loadComments(nextCursor ?? undefined)}
          />
        </>
      ) : null}
    </div>
  );
}

function IssueHeader({ issue }: { issue: Issue }) {
  const { t } = useTranslation();

  return (
    <header className={styles.header}>
      <div className={styles.titleBlock}>
        <p className={styles.issueID}>#{issue.id}</p>
        <h2>{issue.title}</h2>
      </div>
      <div className={styles.badges}>
        <span className={styles.statusBadge}>{t(`statuses.${issue.status}`)}</span>
        <span className={priorityClassName(issue.priority)}>{t(`priorities.${issue.priority}`)}</span>
      </div>
      <dl className={styles.metaGrid}>
        <MetaItem label={t("issues.detail.project")} value={issue.projectKey} />
        <MetaItem label={t("issues.assignee")} value={issue.assignee || t("issues.unassigned")} />
        <MetaItem label={t("issues.detailPage.createdAt")} value={formatDateTime(issue.createdAt)} />
        <MetaItem label={t("issues.detailPage.updatedAt")} value={formatDateTime(issue.updatedAt)} />
      </dl>
    </header>
  );
}

function StatusActions({
  currentStatus,
  disabled,
  onStatusChange,
}: {
  currentStatus: IssueStatus;
  disabled: boolean;
  onStatusChange: (status: IssueStatus) => Promise<void>;
}) {
  const { t } = useTranslation();

  return (
    <section className={styles.section} aria-labelledby="issue-status-actions">
      <div className={styles.sectionHeader}>
        <h3 id="issue-status-actions">{t("issues.detailPage.statusActions")}</h3>
      </div>
      <div className={styles.actions}>
        {issueStatuses.map((status) => (
          <button
            key={status}
            type="button"
            onClick={() => void onStatusChange(status)}
            disabled={disabled || status === currentStatus}
          >
            {t(`statuses.${status}`)}
          </button>
        ))}
      </div>
    </section>
  );
}

function IssueDescription({ issue }: { issue: Issue }) {
  const { t } = useTranslation();

  return (
    <section className={styles.section} aria-labelledby="issue-description">
      <div className={styles.sectionHeader}>
        <h3 id="issue-description">{t("issues.detailPage.description")}</h3>
      </div>
      <Markdown
        className={styles.markdown}
        content={issue.description}
        emptyText={t("issues.noDescription")}
      />
    </section>
  );
}

function CommentList({
  comments,
  error,
  hasMore,
  isLoading,
  onLoadMore,
}: {
  comments: Comment[];
  error: string;
  hasMore: boolean;
  isLoading: boolean;
  onLoadMore: () => void;
}) {
  const { t } = useTranslation();

  return (
    <section className={styles.section} aria-labelledby="issue-comments">
      <div className={styles.sectionHeader}>
        <h3 id="issue-comments">{t("issues.detailPage.comments")}</h3>
        <span>{t("issues.detailPage.commentCount", { count: comments.length })}</span>
      </div>
      {comments.length === 0 && !isLoading ? (
        <p className={styles.emptyText}>{t("issues.detailPage.noComments")}</p>
      ) : null}
      <div className={styles.commentList}>
        {comments.map((comment) => (
          <article key={comment.id} className={styles.comment}>
            <div className={styles.commentHeader}>
              <div>
                <strong>{comment.author}</strong>
                <span>{t(`comments.types.${comment.type}`)}</span>
              </div>
              <time dateTime={comment.createdAt}>{formatDateTime(comment.createdAt)}</time>
            </div>
            <Markdown
              className={styles.markdown}
              content={comment.body}
              emptyText={t("issues.detailPage.emptyComment")}
            />
          </article>
        ))}
      </div>
      {error ? <p className={styles.errorText}>{error}</p> : null}
      {hasMore ? (
        <button type="button" onClick={onLoadMore} disabled={isLoading}>
          {isLoading ? t("issues.detailPage.loadingComments") : t("issues.detailPage.loadMore")}
        </button>
      ) : null}
    </section>
  );
}

function MetaItem({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <dt>{label}</dt>
      <dd>{value}</dd>
    </div>
  );
}

function PanelMessage({ title, detail }: { title: string; detail?: string }) {
  return (
    <section className={styles.section}>
      <h2>{title}</h2>
      {detail ? <p className={styles.errorText}>{detail}</p> : null}
    </section>
  );
}

function parseIssueID(value: string | null): number | null {
  const id = value ? Number.parseInt(value, 10) : Number.NaN;
  return Number.isSafeInteger(id) && id > 0 ? id : null;
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

function priorityClassName(priority: Issue["priority"]): string {
  if (priority === "high" || priority === "urgent") {
    return `${styles.priorityBadge} ${styles.warningPriority}`;
  }
  return styles.priorityBadge;
}
