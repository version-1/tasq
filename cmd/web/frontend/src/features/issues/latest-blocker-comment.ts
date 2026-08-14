import { fetchComments } from "@/lib/api";
import type { Comment } from "@/lib/types";

const commentPageSize = 100;

export async function fetchLatestBlockerComment(issueID: number): Promise<Comment | null> {
  let cursor: number | undefined;
  let latest: Comment | null = null;

  do {
    const page = await fetchComments(issueID, cursor, commentPageSize, { silent: true });
    for (const comment of page.data) {
      if (comment.type === "blocker" && (latest === null || comment.id > latest.id)) {
        latest = comment;
      }
    }
    cursor = page.meta.nextCursor ?? undefined;
  } while (cursor !== undefined);

  return latest;
}
