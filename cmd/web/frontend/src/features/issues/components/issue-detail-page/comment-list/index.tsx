import { useTranslation } from "react-i18next";
import type { Comment } from "@/lib/types";
import type { ChangeRequestShortcut } from "@/features/issues/change-request-shortcuts";
import { CommentCard } from "../comment-card";
import styles from "./index.module.css";

export function CommentList({
  comments,
  error,
  hasMore,
  isLoading,
  latestActionableBlockerCommentID,
  onLoadMore,
  onContinueWithComment,
}: {
  comments: Comment[];
  error: string;
  hasMore: boolean;
  isLoading: boolean;
  latestActionableBlockerCommentID?: number;
  onLoadMore: () => void;
  onContinueWithComment?: (shortcut?: ChangeRequestShortcut) => void;
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
          <CommentCard
            key={comment.id}
            comment={comment}
            onContinueWithComment={
              comment.id === latestActionableBlockerCommentID
                ? onContinueWithComment
                : undefined
            }
          />
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
