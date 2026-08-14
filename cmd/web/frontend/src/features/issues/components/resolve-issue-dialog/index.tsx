import { useCallback, useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { Markdown } from "@/components/ui/markdown";
import type { Comment } from "@/lib/types";
import { builtInChangeRequestShortcuts } from "@/features/issues/change-request-shortcuts";
import { fetchLatestBlockerComment } from "@/features/issues/latest-blocker-comment";
import { ChangeRequestDialog } from "@/features/issues/components/change-request-dialog";
import styles from "./index.module.css";

type BlockerState =
  | { kind: "loading" }
  | { kind: "ready"; comment: Comment }
  | { kind: "error"; message: string };

export function ResolveIssueDialog({
  error,
  isMovingIssue,
  issueID,
  issueTitle,
  loadLatestBlocker = fetchLatestBlockerComment,
  onCancel,
  onMoveIssueReady,
  onSuccess,
}: {
  error?: string;
  isMovingIssue?: boolean;
  issueID: number;
  issueTitle: string;
  loadLatestBlocker?: (issueID: number) => Promise<Comment | null>;
  onCancel: () => void;
  onMoveIssueReady: () => Promise<void>;
  onSuccess: () => void;
}) {
  const { t } = useTranslation();
  const [blockerState, setBlockerState] = useState<BlockerState>({ kind: "loading" });

  const loadBlocker = useCallback(async () => {
    setBlockerState({ kind: "loading" });
    try {
      const comment = await loadLatestBlocker(issueID);
      setBlockerState(
        comment
          ? { kind: "ready", comment }
          : { kind: "error", message: t("issues.resolve.blockerNotFound") },
      );
    } catch (caught) {
      setBlockerState({
        kind: "error",
        message:
          caught instanceof Error ? caught.message : t("issues.resolve.blockerLoadFailed"),
      });
    }
  }, [issueID, loadLatestBlocker, t]);

  useEffect(() => {
    void loadBlocker();
  }, [loadBlocker]);

  const context = (
    <section className={styles.blockerContext} aria-label={t("issues.resolve.latestBlocker")}>
      <h3>{t("issues.resolve.latestBlocker")}</h3>
      {blockerState.kind === "loading" ? <p>{t("issues.resolve.loadingBlocker")}</p> : null}
      {blockerState.kind === "error" ? (
        <div className={styles.blockerError}>
          <p role="alert">{blockerState.message}</p>
          <button type="button" onClick={() => void loadBlocker()}>
            {t("issues.resolve.retryLoad")}
          </button>
        </div>
      ) : null}
      {blockerState.kind === "ready" ? (
        <Markdown
          className={styles.blockerBody}
          content={blockerState.comment.body}
          emptyText={t("issues.resolve.emptyBlocker")}
        />
      ) : null}
    </section>
  );

  return (
    <ChangeRequestDialog
      context={context}
      error={error}
      isMovingIssue={isMovingIssue}
      isSubmissionDisabled={blockerState.kind !== "ready"}
      issueID={issueID}
      issueTitle={issueTitle}
      onCancel={onCancel}
      onMoveIssueReady={onMoveIssueReady}
      onSuccess={onSuccess}
      shortcuts={builtInChangeRequestShortcuts.continue}
      variant="continue"
    />
  );
}
