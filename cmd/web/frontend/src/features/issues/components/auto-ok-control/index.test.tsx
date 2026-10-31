import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { storyIssue } from "@/stories/fixtures";
import "@/lib/i18n";
import { AutoOKControl } from "./index";

function renderControl(props: Parameters<typeof AutoOKControl>[0]) {
  return render(<div role="menu" aria-label="Issue actions"><AutoOKControl {...props} /></div>);
}

describe("AutoOKControl", () => {
  it("sends the issue setting and prevents duplicate changes while saving", async () => {
    const user = userEvent.setup();
    let finish: (() => void) | undefined;
    const onChange = vi.fn(() => new Promise<void>((resolve) => { finish = resolve; }));
    renderControl({ issue: storyIssue(), onChange });
    const toggle = screen.getByRole("menuitemcheckbox");
    await user.click(screen.getByText("Auto OK"));
    expect(onChange).toHaveBeenCalledWith(storyIssue().id, true);
    expect(toggle).toBeDisabled();
    await user.click(toggle);
    expect(onChange).toHaveBeenCalledTimes(1);
    finish?.();
    await waitFor(() => expect(toggle).not.toBeDisabled());
  });

  it("keeps the saved setting after an API failure", async () => {
    const user = userEvent.setup();
    renderControl({ issue: storyIssue(), onChange: vi.fn().mockRejectedValue(new Error("offline")) });
    const toggle = screen.getByRole("menuitemcheckbox");
    await user.click(screen.getByText("Auto OK"));
    await waitFor(() => expect(toggle).not.toBeDisabled());
    expect(toggle).toHaveAttribute("aria-checked", "false");
    expect(screen.getByRole("alert")).toBeInTheDocument();
  });

  it("shows the persisted count and automatic stop reason", () => {
    renderControl({ issue: { ...storyIssue(), autoOkCount: 5, autoOkStopReason: "limit" }, onChange: vi.fn() });
    expect(screen.getByText("5/5")).toBeInTheDocument();
    expect(screen.getByText(/5回に達した|after 5 replies/)).toBeInTheDocument();
    expect(screen.getByRole("menuitemcheckbox")).toHaveAttribute("aria-checked", "false");
  });

  it("exposes the saved checked state and supports keyboard toggling", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn().mockResolvedValue(undefined);
    renderControl({ issue: { ...storyIssue(), autoOk: true, autoOkCount: 2 }, onChange });
    const item = screen.getByRole("menuitemcheckbox");
    expect(item).toHaveAttribute("aria-checked", "true");
    expect(item).toHaveTextContent("2/5");
    expect(item.querySelector('[data-checked="true"]')).toBeInTheDocument();
    expect(item.querySelector("input")).not.toBeInTheDocument();
    item.focus();
    await user.keyboard(" ");
    expect(onChange).toHaveBeenCalledWith(storyIssue().id, false);
  });

  it("disables editing for read-only cards", () => {
    renderControl({ issue: storyIssue(), disabled: true, onChange: vi.fn() });
    expect(screen.getByRole("menuitemcheckbox")).toBeDisabled();
  });
});
