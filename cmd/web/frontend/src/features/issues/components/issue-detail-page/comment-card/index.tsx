import { useId, useState } from "react";
import { useTranslation } from "react-i18next";
import { ContextMenu, ContextMenuItem } from "@/components/ui/context-menu";
import { IconProxy } from "@/components/ui/icon-proxy";
import { Markdown } from "@/components/ui/markdown";
import { CommentTypeBadge } from "@/features/issues/components/comment-type-badge";
import type { Comment } from "@/lib/types";
import { formatDateTime } from "../format";
import styles from "./index.module.css";

export function CommentCard({
  comment,
  onContinueWithComment,
}: {
  comment: Comment;
  onContinueWithComment?: () => void;
}) {
  const { t } = useTranslation();
  const menuID = useId();
  const [isMenuOpen, setIsMenuOpen] = useState(false);
  const menuLabel = t("issues.detailPage.commentActions", { author: comment.author });

  return (
    <article className={styles.comment}>
      <div className={styles.commentHeader}>
        <div className={styles.commentIdentity}>
          <strong>{comment.author}</strong>
          <CommentTypeBadge type={comment.type} />
        </div>
        <div className={styles.headerActions}>
          <time dateTime={comment.createdAt}>{formatDateTime(comment.createdAt)}</time>
          {onContinueWithComment ? (
            <ContextMenu
              id={menuID}
              isOpen={isMenuOpen}
              label={menuLabel}
              onOpenChange={setIsMenuOpen}
              trigger={(triggerProps) => (
                <button
                  {...triggerProps}
                  aria-label={menuLabel}
                  className={styles.menuButton}
                  title={menuLabel}
                  type="button"
                >
                  <IconProxy name="ellipsis" size={16} />
                </button>
              )}
            >
              <ContextMenuItem
                icon={<IconProxy name="arrow-right" />}
                onSelect={() => {
                  setIsMenuOpen(false);
                  onContinueWithComment();
                }}
              >
                {t("issues.continueWithComment.action")}
              </ContextMenuItem>
            </ContextMenu>
          ) : (
            <button
              aria-label={menuLabel}
              className={styles.menuButton}
              title={menuLabel}
              type="button"
            >
              <IconProxy name="ellipsis" size={16} />
            </button>
          )}
        </div>
      </div>
      <Markdown
        className={styles.markdown}
        content={comment.body}
        emptyText={t("issues.detailPage.emptyComment")}
      />
    </article>
  );
}
