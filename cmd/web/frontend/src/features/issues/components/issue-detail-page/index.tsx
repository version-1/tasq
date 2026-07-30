"use client";

import { useParams, useSearchParams } from "react-router-dom";
import { useCallback, useEffect, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "@/lib/toast";
import { useOptionalLayoutShellData } from "@/components/layout";
import { PanelMessage } from "@/components/ui/pannel-message";
import {
  fetchChangeRequests,
  fetchComments,
  fetchIssueAttachments,
  fetchIssue,
  fetchOrchestratorConversation,
  fetchOrchestratorIssueRuntime,
  updateIssueDescription,
  updateIssueStatus,
} from "@/lib/api";
import type {
  Attachment,
  ChangeRequest,
  Comment,
  Issue,
  IssueStatus,
  OrchestratorConversation,
  OrchestratorIssueRun,
} from "@/lib/types";
import { AttachmentsSection } from "./attachments-section";
import { ChangeRequestList } from "./change-request-list";
import { CommentList } from "./comment-list";
import { ConversationTab, defaultConversationMessageTypes } from "./conversation-tab";
import { IssueDescription } from "./issue-description";
import { BasicInfoPanel } from "./basic-info-panel";
import { RunsSection } from "./runs-section";
import { ChangeRequestDialog } from "../change-request-dialog";
import styles from "./index.module.css";

const commentPageSize = 20;
const validTabs = ["details", "comments", "conversation"] as const;
type IssueDetailTab = (typeof validTabs)[number];
type ChangeRequestDialogVariant = "continue" | "reject";

type IssueLoadState =
  | { kind: "loading" }
  | { kind: "ready"; issue: Issue }
  | { kind: "error"; message: string };

export function IssueDetailPage() {
  const { t } = useTranslation();
  const layoutShellData = useOptionalLayoutShellData();
  const onIssueDetailTitleChange = layoutShellData?.onIssueDetailTitleChange;
  const { id } = useParams();
  const [searchParams, setSearchParams] = useSearchParams();
  const issueID = parseIssueID(id);
  const activeTab = parseTab(searchParams.get("tab"));
  const selectedRunID = searchParams.get("runId") ?? "";
  const selectedMessageTypes = parseMessageTypes(searchParams.get("messageTypes"));
  const [issueState, setIssueState] = useState<IssueLoadState>({ kind: "loading" });
  const [attachments, setAttachments] = useState<Attachment[]>([]);
  const [attachmentsError, setAttachmentsError] = useState("");
  const [isLoadingAttachments, setIsLoadingAttachments] = useState(false);
  const [comments, setComments] = useState<Comment[]>([]);
  const [nextCursor, setNextCursor] = useState<number | null>(null);
  const [commentsError, setCommentsError] = useState("");
  const [isLoadingComments, setIsLoadingComments] = useState(false);
  const [commentsLoaded, setCommentsLoaded] = useState(false);
  const [changeRequests, setChangeRequests] = useState<ChangeRequest[]>([]);
  const [changeRequestsError, setChangeRequestsError] = useState("");
  const [isLoadingChangeRequests, setIsLoadingChangeRequests] = useState(false);
  const [changeRequestsLoaded, setChangeRequestsLoaded] = useState(false);
  const [runs, setRuns] = useState<OrchestratorIssueRun[]>([]);
  const [runsError, setRunsError] = useState("");
  const [isLoadingRuns, setIsLoadingRuns] = useState(false);
  const [conversation, setConversation] = useState<OrchestratorConversation | null>(null);
  const [conversationError, setConversationError] = useState("");
  const [isLoadingConversation, setIsLoadingConversation] = useState(false);
  const [isUpdatingStatus, setIsUpdatingStatus] = useState(false);
  const [isUpdatingDescription, setIsUpdatingDescription] = useState(false);
  const [changeRequestDialogVariant, setChangeRequestDialogVariant] =
    useState<ChangeRequestDialogVariant | null>(null);
  const [changeRequestError, setChangeRequestError] = useState("");

  const updateSearchParams = useCallback(
    (updates: Record<string, string | null>) => {
      setSearchParams((current) => {
        const next = new URLSearchParams(current);
        for (const [key, value] of Object.entries(updates)) {
          if (value === null || value === "") {
            next.delete(key);
          } else {
            next.set(key, value);
          }
        }
        return next;
      });
    },
    [setSearchParams],
  );

  const loadIssue = useCallback(async () => {
    if (!issueID) {
      setIssueState({ kind: "error", message: t("issues.detailPage.invalidIssue") });
      return;
    }

    setIssueState({ kind: "loading" });
    try {
      const issue = await fetchIssue(issueID, { silent: true });
      setIssueState({ kind: "ready", issue });
    } catch (error) {
      setIssueState({
        kind: "error",
        message: error instanceof Error ? error.message : t("issues.detailPage.failedToLoadIssue"),
      });
    }
  }, [issueID, t]);

  const loadAttachments = useCallback(async () => {
    if (!issueID) return;

    setIsLoadingAttachments(true);
    setAttachmentsError("");
    try {
      const response = await fetchIssueAttachments(issueID, { silent: true });
      setAttachments(Array.isArray(response.data) ? response.data : []);
    } catch (error) {
      setAttachmentsError(
        error instanceof Error ? error.message : t("issues.detailPage.failedToLoadAttachments"),
      );
      setAttachments([]);
    } finally {
      setIsLoadingAttachments(false);
    }
  }, [issueID, t]);

  const loadComments = useCallback(
    async (cursor?: number) => {
      if (!issueID) return;

      setIsLoadingComments(true);
      setCommentsError("");
      try {
        const page = await fetchComments(issueID, cursor, commentPageSize, { silent: true });
        setComments((current) => (cursor ? [...current, ...page.data] : page.data));
        setNextCursor(page.meta.nextCursor);
        setCommentsLoaded(true);
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

  const loadChangeRequests = useCallback(async () => {
    if (!issueID) return;

    setIsLoadingChangeRequests(true);
    setChangeRequestsError("");
    try {
      const response = await fetchChangeRequests(issueID, 100, { silent: true });
      setChangeRequests(Array.isArray(response.data) ? response.data : []);
      setChangeRequestsLoaded(true);
    } catch (error) {
      setChangeRequestsError(
        error instanceof Error ? error.message : t("issues.detailPage.failedToLoadChangeRequests"),
      );
      setChangeRequests([]);
    } finally {
      setIsLoadingChangeRequests(false);
    }
  }, [issueID, t]);

  const loadRuns = useCallback(async () => {
    if (!issueID) return;

    setIsLoadingRuns(true);
    setRunsError("");
    try {
      const runtime = await fetchOrchestratorIssueRuntime(issueID, { silent: true });
      setRuns(runtime.runs);
    } catch (error) {
      setRunsError(
        error instanceof Error ? error.message : t("issues.detailPage.failedToLoadRuns"),
      );
      setRuns([]);
    } finally {
      setIsLoadingRuns(false);
    }
  }, [issueID, t]);

  useEffect(() => {
    void loadIssue();
  }, [loadIssue]);

  useEffect(() => {
    if (issueState.kind === "ready") {
      onIssueDetailTitleChange?.(issueState.issue.title);
      return () => onIssueDetailTitleChange?.(null);
    }
    onIssueDetailTitleChange?.(null);
    return undefined;
  }, [issueState, onIssueDetailTitleChange]);

  useEffect(() => {
    setComments([]);
    setNextCursor(null);
    setCommentsError("");
    setCommentsLoaded(false);
    setChangeRequests([]);
    setChangeRequestsError("");
    setChangeRequestsLoaded(false);
  }, [issueID]);

  useEffect(() => {
    setAttachments([]);
    setAttachmentsError("");
    void loadAttachments();
  }, [loadAttachments]);

  useEffect(() => {
    setRuns([]);
    setRunsError("");
    void loadRuns();
  }, [loadRuns]);

  useEffect(() => {
    if (activeTab === "comments" && !commentsLoaded && !isLoadingComments) {
      void loadComments();
    }
  }, [activeTab, commentsLoaded, isLoadingComments, loadComments]);

  useEffect(() => {
    if (activeTab === "comments" && !changeRequestsLoaded && !isLoadingChangeRequests) {
      void loadChangeRequests();
    }
  }, [activeTab, changeRequestsLoaded, isLoadingChangeRequests, loadChangeRequests]);

  useEffect(() => {
    if (activeTab !== "conversation" || runs.length === 0) {
      return;
    }
    const hasSelectedRun = selectedRunID && runs.some((run) => run.run_id === selectedRunID);
    if (!hasSelectedRun) {
      updateSearchParams({ runId: runs[0].run_id });
    }
  }, [activeTab, runs, selectedRunID, updateSearchParams]);

  useEffect(() => {
    if (activeTab !== "conversation" || !selectedRunID) {
      return;
    }
    if (!issueID) {
      return;
    }
    if (runs.length > 0 && !runs.some((run) => run.run_id === selectedRunID)) {
      return;
    }

    let active = true;
    setConversation(null);
    setIsLoadingConversation(true);
    setConversationError("");
    void fetchOrchestratorConversation(issueID, selectedRunID, { silent: true })
      .then((response) => {
        if (active) {
          setConversation(response);
        }
      })
      .catch((error) => {
        if (active) {
          setConversationError(
            error instanceof Error ? error.message : t("issues.detailPage.failedToLoadConversation"),
          );
          setConversation(null);
        }
      })
      .finally(() => {
        if (active) {
          setIsLoadingConversation(false);
        }
      });
    return () => {
      active = false;
    };
  }, [activeTab, issueID, runs, selectedRunID, t]);

  async function handleStatusChange(status: IssueStatus) {
    if (issueState.kind !== "ready" || issueState.issue.status === status) return;

    setIsUpdatingStatus(true);
    try {
      const issue = await updateIssueStatus(issueState.issue.id, status);
      setIssueState({ kind: "ready", issue });
      toast.success({ message: t("toast.success.issueStatusUpdated") });
    } catch {
      // Error toast is emitted by the API wrapper.
    } finally {
      setIsUpdatingStatus(false);
    }
  }

  async function handleMoveIssueReady() {
    if (issueState.kind !== "ready") return;

    setIsUpdatingStatus(true);
    setChangeRequestError("");
    try {
      const issue = await updateIssueStatus(issueState.issue.id, "ready", { silent: true });
      setIssueState({ kind: "ready", issue });
      toast.success({
        message:
          changeRequestDialogVariant === "continue"
            ? t("toast.success.continuedWithComment")
            : t("toast.success.issueRejected"),
      });
    } catch (error) {
      const message =
        error instanceof Error
          ? error.message
          : t(
              changeRequestDialogVariant === "continue"
                ? "issues.continueWithComment.errors.statusUpdateFailed"
                : "issues.reject.errors.statusUpdateFailed",
            );
      setChangeRequestError(message);
      throw new Error(message);
    } finally {
      setIsUpdatingStatus(false);
    }
  }

  async function handleDescriptionSave(description: string) {
    if (issueState.kind !== "ready") return;

    setIsUpdatingDescription(true);
    try {
      const issue = await updateIssueDescription(issueState.issue.id, description);
      setIssueState({ kind: "ready", issue });
      toast.success({ message: t("toast.success.issueUpdated") });
    } catch (error) {
      throw error;
    } finally {
      setIsUpdatingDescription(false);
    }
  }

  const sortedComments = useMemo(() => {
    return [...comments].sort((left, right) => left.id - right.id);
  }, [comments]);
  const sortedChangeRequests = useMemo(() => {
    return [...changeRequests].sort((left, right) => left.createdAt.localeCompare(right.createdAt));
  }, [changeRequests]);
  const latestBlockerCommentID = useMemo(() => {
    return comments.reduce<number | undefined>((latestID, comment) => {
      if (comment.type !== "blocker") return latestID;
      return latestID === undefined || comment.id > latestID ? comment.id : latestID;
    }, undefined);
  }, [comments]);

  function handleRunChange(runID: string) {
    updateSearchParams({ runId: runID });
  }

  function handleMessageTypesChange(types: string[]) {
    updateSearchParams({ messageTypes: types.length === 0 ? "all" : types.join(",") });
  }

  return (
    <div className={styles.page}>
      {issueState.kind === "loading" ? (
        <PanelMessage title={t("layout.loading")} />
      ) : null}

      {issueState.kind === "error" ? (
        <PanelMessage title={t("issues.detailPage.failedTitle")} detail={issueState.message} />
      ) : null}

      {issueState.kind === "ready" ? (
        <>
          {activeTab === "details" ? (
            <div className={styles.tabPanel}>
              <div className={styles.detailLayout}>
                <div className={styles.detailMain}>
                  <IssueDescription
                    issue={issueState.issue}
                    isSaving={isUpdatingDescription}
                    onSave={handleDescriptionSave}
                  />
                  <AttachmentsSection
                    attachments={attachments}
                    error={attachmentsError}
                    isLoading={isLoadingAttachments}
                  />
                </div>
                <aside className={styles.detailAside} aria-label={t("issues.detailPage.basicInfo")}>
                  <BasicInfoPanel
                    disabled={isUpdatingStatus}
                    issue={issueState.issue}
                    issueOptions={layoutShellData?.issues ?? []}
                    onRejectIssue={() => {
                      setChangeRequestError("");
                      setChangeRequestDialogVariant("reject");
                    }}
                    onStatusChange={handleStatusChange}
                  />
                </aside>
              </div>
            </div>
          ) : null}
          {activeTab === "comments" ? (
            <div className={styles.tabPanel}>
              <RunsSection
                issueID={issueState.issue.id}
                error={runsError}
                isLoading={isLoadingRuns}
                runs={runs}
              />
              <ChangeRequestList
                changeRequests={sortedChangeRequests}
                error={changeRequestsError}
                isLoading={isLoadingChangeRequests}
              />
              <CommentList
                comments={sortedComments}
                error={commentsError}
                hasMore={nextCursor !== null}
                isLoading={isLoadingComments}
                latestActionableBlockerCommentID={
                  issueState.issue.status === "blocked" ? latestBlockerCommentID : undefined
                }
                onLoadMore={() => void loadComments(nextCursor ?? undefined)}
                onContinueWithComment={() => {
                  setChangeRequestError("");
                  setChangeRequestDialogVariant("continue");
                }}
              />
            </div>
          ) : null}
          {activeTab === "conversation" ? (
            <ConversationTab
              conversation={conversation}
              error={conversationError}
              isLoading={isLoadingConversation}
              messageTypes={selectedMessageTypes}
              onMessageTypesChange={handleMessageTypesChange}
              onRunChange={handleRunChange}
              runs={runs}
              runsError={runsError}
              runsLoading={isLoadingRuns}
              selectedRunID={selectedRunID}
            />
          ) : null}
        </>
      ) : null}
      {issueState.kind === "ready" && changeRequestDialogVariant ? (
        <ChangeRequestDialog
          error={changeRequestError}
          isMovingIssue={isUpdatingStatus}
          issueID={issueState.issue.id}
          issueTitle={issueState.issue.title}
          onCancel={() => {
            setChangeRequestError("");
            setChangeRequestDialogVariant(null);
          }}
          onMoveIssueReady={handleMoveIssueReady}
          onSuccess={() => {
            setChangeRequestError("");
            setChangeRequestDialogVariant(null);
            if (changeRequestDialogVariant === "continue") {
              void loadChangeRequests();
            }
          }}
          variant={changeRequestDialogVariant}
        />
      ) : null}
    </div>
  );
}

function parseIssueID(value: string | undefined): number | null {
  const id = value ? Number.parseInt(value, 10) : Number.NaN;
  return Number.isSafeInteger(id) && id > 0 ? id : null;
}

function parseTab(value: string | null): IssueDetailTab {
  return validTabs.includes(value as IssueDetailTab) ? (value as IssueDetailTab) : "details";
}

function parseMessageTypes(value: string | null): string[] {
  if (value === null) {
    return [...defaultConversationMessageTypes];
  }
  if (value === "all") {
    return [];
  }
  return value.split(",").map((item) => item.trim()).filter(Boolean);
}
