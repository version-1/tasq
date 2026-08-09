import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import "@/lib/i18n";
import type { Comment } from "@/lib/types";
import { CommentCard } from "./index";

const comment: Comment = {
  id: 1,
  issueId: 24,
  author: "codex",
  type: "progress",
  body: "Implemented the context menu icons.",
  createdAt: "2026-07-31T08:29:15.000Z",
};

describe("CommentCard", () => {
  it("renders an icon for the continue action", async () => {
    const user = userEvent.setup();
    const onContinueWithComment = vi.fn();

    render(
      <CommentCard comment={comment} onContinueWithComment={onContinueWithComment} />,
    );

    await user.click(screen.getByRole("button", { name: "Comment actions for codex" }));
    const menu = screen.getByRole("menu", { name: "Comment actions for codex" });
    const action = within(menu).getByRole("menuitem", { name: "Continue with comment" });

    expect(action.querySelector(".lucide-arrow-right")).toBeInTheDocument();
    await user.click(action);
    expect(onContinueWithComment).toHaveBeenCalledOnce();
    expect(screen.queryByRole("menu")).not.toBeInTheDocument();
  });
});
