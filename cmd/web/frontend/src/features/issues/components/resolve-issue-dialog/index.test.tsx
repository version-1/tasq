import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createChangeRequest, fetchComments } from "@/lib/api";
import type { CommentListResponse } from "@/lib/types";
import "@/lib/i18n";
import { ResolveIssueDialog } from "./index";

vi.mock("@/lib/api", () => ({
  createChangeRequest: vi.fn(),
  fetchComments: vi.fn(),
}));

describe("ResolveIssueDialog", () => {
  beforeEach(() => vi.clearAllMocks());

  it("shows the latest blocker and immediately submits a shortcut", async () => {
    const user = userEvent.setup();
    const onMoveIssueReady = vi.fn().mockResolvedValue(undefined);
    const onSuccess = vi.fn();
    vi.mocked(fetchComments).mockResolvedValue({
      data: [{
        id: 7,
        issueId: 42,
        author: "runner",
        type: "blocker",
        body: "Approval reason was missing.",
        createdAt: "2026-08-14T00:00:00.000Z",
      }],
      meta: { cursor: 0, limit: 100, direction: "forward", nextCursor: null },
    } satisfies CommentListResponse);
    vi.mocked(createChangeRequest).mockResolvedValue({} as never);

    render(
      <ResolveIssueDialog
        issueID={42}
        issueTitle="Approval request blocked"
        onCancel={vi.fn()}
        onMoveIssueReady={onMoveIssueReady}
        onSuccess={onSuccess}
      />,
    );

    expect(await screen.findByText("Approval reason was missing.")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Retry" }));

    await waitFor(() => {
      expect(createChangeRequest).toHaveBeenCalledWith(42, {
        author: "reviewer",
        body: "Retry",
      }, { silent: true });
    });
    expect(onMoveIssueReady).toHaveBeenCalledOnce();
    expect(onSuccess).toHaveBeenCalledOnce();
  });

  it("keeps submissions disabled and allows retrying when blocker loading fails", async () => {
    const user = userEvent.setup();
    vi.mocked(fetchComments)
      .mockRejectedValueOnce(new Error("comments unavailable"))
      .mockResolvedValueOnce({
        data: [{
          id: 7,
          issueId: 42,
          author: "runner",
          type: "blocker",
          body: "Recovered blocker.",
          createdAt: "2026-08-14T00:00:00.000Z",
        }],
        meta: { cursor: 0, limit: 100, direction: "forward", nextCursor: null },
      } satisfies CommentListResponse);

    render(
      <ResolveIssueDialog
        issueID={42}
        issueTitle="Approval request blocked"
        onCancel={vi.fn()}
        onMoveIssueReady={vi.fn()}
        onSuccess={vi.fn()}
      />,
    );

    expect(await screen.findByRole("alert")).toHaveTextContent("comments unavailable");
    expect(screen.getByRole("button", { name: "Retry" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Continue with comment" })).toBeDisabled();

    await user.click(screen.getByRole("button", { name: "Reload" }));
    expect(await screen.findByText("Recovered blocker.")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Retry" })).toBeEnabled();
  });

  it("does not create a second request when status recovery follows a shortcut", async () => {
    const user = userEvent.setup();
    const onMoveIssueReady = vi.fn()
      .mockRejectedValueOnce(new Error("status unavailable"))
      .mockResolvedValueOnce(undefined);
    vi.mocked(fetchComments).mockResolvedValue({
      data: [{
        id: 7,
        issueId: 42,
        author: "runner",
        type: "blocker",
        body: "Retry deployment.",
        createdAt: "2026-08-14T00:00:00.000Z",
      }],
      meta: { cursor: 0, limit: 100, direction: "forward", nextCursor: null },
    } satisfies CommentListResponse);
    vi.mocked(createChangeRequest).mockResolvedValue({} as never);

    render(
      <ResolveIssueDialog
        issueID={42}
        issueTitle="Deployment blocked"
        onCancel={vi.fn()}
        onMoveIssueReady={onMoveIssueReady}
        onSuccess={vi.fn()}
      />,
    );

    await screen.findByText("Retry deployment.");
    await user.click(screen.getByRole("button", { name: "Retry" }));
    expect(await screen.findByText("status unavailable")).toBeInTheDocument();
    expect(screen.getByRole("textbox", { name: "Change request" })).toHaveValue("Retry");
    expect(screen.getByRole("textbox", { name: "Change request" })).toHaveAttribute("readonly");

    await user.click(screen.getByRole("button", { name: "Continue with comment" }));
    expect(createChangeRequest).toHaveBeenCalledOnce();
    expect(onMoveIssueReady).toHaveBeenCalledTimes(2);
  });
});
