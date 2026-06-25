import { fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";
import { RunsSection } from ".";
import type { OrchestratorIssueRun } from "@/lib/types";
import { toastStore } from "@/lib/toast";

describe("RunsSection", () => {
  afterEach(() => {
    toastStore.clear();
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
  });

  it("shows thread IDs when present and a placeholder when absent", () => {
    render(
      <MemoryRouter>
        <RunsSection issueID={12} error="" isLoading={false} runs={runs} />
      </MemoryRouter>,
    );

    const links = screen.getAllByRole("link");
    expect(within(links[0]).getByText("thread-latest")).toBeInTheDocument();
    expect(within(links[1]).getByText("not set")).toBeInTheDocument();
  });

  it("copies a present thread ID to the clipboard", async () => {
    const writeText = vi.fn().mockResolvedValue(undefined);
    vi.stubGlobal("navigator", { ...navigator, clipboard: { writeText } });

    render(
      <MemoryRouter>
        <RunsSection issueID={12} error="" isLoading={false} runs={runs} />
      </MemoryRouter>,
    );

    const copyButtons = screen.getAllByRole("button", { name: "Copy thread ID" });
    expect(copyButtons[0]).toBeEnabled();
    fireEvent.click(copyButtons[0]);

    expect(writeText).toHaveBeenCalledWith("thread-latest");
    await waitFor(() => {
      expect(toastStore.getSnapshot()).toMatchObject([
        { type: "success", message: "Thread ID copied" },
      ]);
    });
  });

  it("does not allow copying when the thread ID is absent", () => {
    render(
      <MemoryRouter>
        <RunsSection issueID={12} error="" isLoading={false} runs={runs} />
      </MemoryRouter>,
    );

    expect(screen.getAllByRole("button", { name: "Copy thread ID" })[1]).toBeDisabled();
  });
});

const runs: OrchestratorIssueRun[] = [
  {
    run_id: "run-latest",
    thread_id: "thread-latest",
    status: "succeeded",
    attempt: 2,
    created_at: "2026-06-08T01:00:00.000Z",
    updated_at: "2026-06-08T02:00:00.000Z",
  },
  {
    run_id: "run-previous",
    status: "failed",
    attempt: 1,
    created_at: "2026-06-08T00:00:00.000Z",
    updated_at: "2026-06-08T00:30:00.000Z",
  },
];
