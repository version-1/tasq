import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { storyIssue } from "@/stories/fixtures";
import "@/lib/i18n";
import { AutoOKControl } from "./index";

describe("AutoOKControl", () => {
  it("sends the issue setting and prevents duplicate changes while saving", async () => {
    const user = userEvent.setup();
    let finish: (() => void) | undefined;
    const onChange = vi.fn(() => new Promise<void>((resolve) => { finish = resolve; }));
    render(<AutoOKControl issue={storyIssue()} onChange={onChange} />);
    const toggle = screen.getByRole("checkbox");
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
    render(<AutoOKControl issue={storyIssue()} onChange={vi.fn().mockRejectedValue(new Error("offline"))} />);
    const toggle = screen.getByRole("checkbox");
    await user.click(screen.getByText("Auto OK"));
    await waitFor(() => expect(toggle).not.toBeDisabled());
    expect(toggle).not.toBeChecked();
    expect(screen.getByRole("alert")).toBeInTheDocument();
  });

  it("shows the persisted count and automatic stop reason", () => {
    render(<AutoOKControl issue={{ ...storyIssue(), autoOkCount: 5, autoOkStopReason: "limit" }} onChange={vi.fn()} />);
    expect(screen.getByText("5/5")).toBeInTheDocument();
    expect(screen.getByText(/5回に達した|after 5 replies/)).toBeInTheDocument();
    expect(screen.getByRole("checkbox")).not.toBeChecked();
  });
});
