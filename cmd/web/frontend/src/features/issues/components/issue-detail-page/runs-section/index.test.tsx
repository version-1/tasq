import { render, screen, within } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it } from "vitest";
import { RunsSection } from ".";
import type { OrchestratorIssueRun } from "@/lib/types";

describe("RunsSection", () => {
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
