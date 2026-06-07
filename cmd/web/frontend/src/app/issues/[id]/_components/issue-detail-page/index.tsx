"use client";

import { Link, useParams } from "react-router-dom";
import { useCallback, useEffect, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { fetchComments, fetchIssue, updateIssueStatus } from "@/lib/api";
import type { Comment, Issue, IssueStatus } from "@/lib/types";
import { CommentList } from "./comment-list";
import { IssueDescription } from "./issue-description";
import { IssueHeader } from "./issue-header";
import { PanelMessage } from "./panel-message";
import { StatusActions } from "./status-actions";
import styles from "./index.module.css";

const commentPageSize = 20;

type IssueLoadState =
  | { kind: "loading" }
  | { kind: "ready"; issue: Issue }
  | { kind: "error"; message: string };

export function IssueDetailPage() {
  const { t } = useTranslation();
  const { id } = useParams();
  const issueID = parseIssueID(id);
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

function parseIssueID(value: string | undefined): number | null {
  const id = value ? Number.parseInt(value, 10) : Number.NaN;
  return Number.isSafeInteger(id) && id > 0 ? id : null;
}
