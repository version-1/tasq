import { beforeEach, describe, expect, it, vi } from "vitest";
import { fetchComments } from "@/lib/api";
import type { CommentListResponse } from "@/lib/types";
import { fetchLatestBlockerComment } from "./latest-blocker-comment";

vi.mock("@/lib/api", () => ({ fetchComments: vi.fn() }));

describe("fetchLatestBlockerComment", () => {
  beforeEach(() => vi.clearAllMocks());

  it("returns the blocker with the greatest ID across all pages", async () => {
    vi.mocked(fetchComments)
      .mockResolvedValueOnce({
        data: [comment(4, "blocker", "Older blocker"), comment(8, "general", "Note")],
        meta: { cursor: 0, limit: 100, direction: "forward", nextCursor: 8 },
      } satisfies CommentListResponse)
      .mockResolvedValueOnce({
        data: [comment(12, "blocker", "Latest blocker")],
        meta: { cursor: 8, limit: 100, direction: "forward", nextCursor: null },
      } satisfies CommentListResponse);

    await expect(fetchLatestBlockerComment(42)).resolves.toMatchObject({
      id: 12,
      body: "Latest blocker",
    });
    expect(fetchComments).toHaveBeenNthCalledWith(2, 42, 8, 100, { silent: true });
  });
});

function comment(id: number, type: "blocker" | "general", body: string) {
  return {
    id,
    issueId: 42,
    author: "runner",
    type,
    body,
    createdAt: "2026-08-14T00:00:00.000Z",
  };
}
