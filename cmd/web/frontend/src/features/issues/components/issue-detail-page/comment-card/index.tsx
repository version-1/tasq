import { useTranslation } from "react-i18next";
import { Markdown } from "@/components/ui/markdown";
import { CommentTypeBadge } from "@/features/issues/components/comment-type-badge";
import type { Comment } from "@/lib/types";
import { formatDateTime } from "../format";
import styles from "./index.module.css";

export function CommentCard({ comment }: { comment: Comment }) {
  const { t } = useTranslation();

  return (
    <article className={styles.comment}>
      <div className={styles.commentHeader}>
        <div>
          <strong>{comment.author}</strong>
          <CommentTypeBadge type={comment.type} />
        </div>
        <time dateTime={comment.createdAt}>{formatDateTime(comment.createdAt)}</time>
      </div>
      <Markdown
        className={styles.markdown}
        content={comment.body}
        emptyText={t("issues.detailPage.emptyComment")}
      />
    </article>
  );
}
