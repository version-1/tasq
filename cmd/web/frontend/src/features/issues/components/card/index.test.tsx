import { render, screen, waitFor, within } from "@testing-library/react";
import { QueryClientProvider, type QueryClient } from "@tanstack/react-query";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import * as api from "@/lib/api";
import { createAppQueryClient } from "@/lib/query-client";
import { toastStore } from "@/lib/toast";
import type { IssueSummary, OrchestratorIssueRuntime } from "@/lib/types";
import "@/lib/i18n";
import { IssueCard } from "./index";

vi.mock("@/lib/api", () => ({
  fetchOrchestratorIssueRuntime: vi.fn(),
}));

const issue: IssueSummary = {
  id: 24,
  projectId: 1,
  projectKey: "tasq",
  title: "Wire issue board to generated client",
  description: "Issue body should not render in the card.",
  status: "ready",
  priority: "high",
  assignee: "web",
  dependency_ids: [],
  artifacts: [],
  queueStatus: "queued",
  createdAt: "2026-06-01T00:00:00.000Z",
  updatedAt: "2026-06-01T00:00:00.000Z",
  stats: {
    commentCount: 0,
  },
};

let queryClient: QueryClient;

function issueWithCommentCount(commentCount: number): IssueSummary {
  return {
    ...issue,
    stats: {
      ...issue.stats,
      commentCount,
    },
  };
}

function renderCard(props: Partial<Parameters<typeof IssueCard>[0]> = {}) {
  const onRejectIssue = vi.fn();
  const onStatusChange = vi.fn(async () => undefined);
  const rendered = render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter>
        <IssueCard
          issue={issue}
          onRejectIssue={onRejectIssue}
          onStatusChange={onStatusChange}
          {...props}
        />
      </MemoryRouter>
    </QueryClientProvider>,
  );
  return { onRejectIssue, onStatusChange, unmount: rendered.unmount };
}

describe("IssueCard", () => {
  beforeEach(() => {
    queryClient = createAppQueryClient();
  });

  afterEach(() => {
    queryClient.clear();
    toastStore.clear();
    vi.clearAllMocks();
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
  });

  it("renders summary fields without the issue body", () => {
    renderCard({ issue: issueWithCommentCount(3), runCount: 2 });

    expect(screen.getByRole("link", { name: "#24 Wire issue board to generated client" })).toBeInTheDocument();
    expect(screen.queryByText("Issue body should not render in the card.")).not.toBeInTheDocument();
    expect(screen.getByText("tasq")).toBeInTheDocument();
    expect(screen.getByText("ready")).toBeInTheDocument();
    expect(screen.getByText("high")).toBeInTheDocument();
    expect(screen.getByLabelText("3 comments")).toHaveTextContent("3");
    expect(screen.getByLabelText("2 runs")).toHaveTextContent("2");
  });

  it("renders a right-aligned pending badge only for pending queue status", () => {
    renderCard({
      issue: {
        ...issue,
        queueStatus: "pending",
      },
    });

    expect(screen.getByText("pending")).toBeInTheDocument();
  });

  it("does not render a pending badge for queued issues", () => {
    renderCard({
      issue: {
        ...issue,
        queueStatus: "queued",
      },
    });

    expect(screen.queryByText("pending")).not.toBeInTheDocument();
  });

  it("renders zero comment count from issue stats", () => {
    renderCard();

    expect(screen.getByLabelText("0 comments")).toHaveTextContent("0");
  });

  it("renders zero run count when it is not provided", () => {
    renderCard();

    expect(screen.getByLabelText("0 runs")).toHaveTextContent("0");
  });

  it("limits ready status transitions in the action menu", async () => {
    const user = userEvent.setup();
    renderCard();

    await user.click(screen.getByRole("button", {
      name: "Issue actions for Wire issue board to generated client",
    }));
    const menu = screen.getByRole("menu", {
      name: "Issue actions for Wire issue board to generated client",
    });
    const items = within(menu).getAllByRole("menuitem").map((item) => item.textContent);

    expect(items).toEqual(["Copy thread ID", "Ready (current)", "Backlog", "Cancelled", "Done", "Duplicate"]);
    expect(within(menu).getByRole("menuitem", { name: "Copy thread ID" }).querySelector(".lucide-copy"))
      .toBeInTheDocument();
    expect(within(menu).getByRole("menuitem", { name: "Ready (current)" }).querySelector(".lucide-circle-play"))
      .toBeInTheDocument();
    expect(within(menu).getByRole("menuitem", { name: "Ready (current)" }).querySelector(".lucide-check"))
      .toBeInTheDocument();
    expect(within(menu).getByRole("menuitem", { name: "Backlog" }).querySelector(".lucide-inbox"))
      .toBeInTheDocument();
    expect(within(menu).getByRole("menuitem", { name: "Cancelled" }).querySelector(".lucide-ban"))
      .toBeInTheDocument();
    expect(within(menu).getByRole("menuitem", { name: "Done" }).querySelector(".lucide-circle-check"))
      .toBeInTheDocument();
    expect(within(menu).getByRole("menuitem", { name: "Duplicate" }).querySelector(".lucide-copy"))
      .toBeInTheDocument();
    expect(within(menu).getAllByRole("separator")).toHaveLength(2);
  });

  it("runs allowed status changes from the action menu", async () => {
    const user = userEvent.setup();
    const { onStatusChange } = renderCard();

    await user.click(screen.getByRole("button", {
      name: "Issue actions for Wire issue board to generated client",
    }));
    await user.click(screen.getByRole("menuitem", { name: "Backlog" }));

    expect(onStatusChange).toHaveBeenCalledWith(24, "backlog");
    expect(screen.queryByRole("menu")).not.toBeInTheDocument();
  });

  it("opens the pull request artifact in a safe new tab and closes the action menu", async () => {
    const user = userEvent.setup();
    const open = vi.spyOn(window, "open").mockImplementation(() => null);
    renderCard({
      issue: {
        ...issue,
        artifacts: [
          {
            type: "pull_request",
            data_type: "url",
            data_value: "https://github.com/version-1/tasq/pull/14",
          },
        ],
      },
    });

    await user.click(screen.getByRole("button", {
      name: "Issue actions for Wire issue board to generated client",
    }));
    await user.click(screen.getByRole("menuitem", { name: "Open pull request" }));

    expect(open).toHaveBeenCalledWith(
      "https://github.com/version-1/tasq/pull/14",
      "_blank",
      "noopener,noreferrer",
    );
    expect(screen.queryByRole("menu")).not.toBeInTheDocument();
  });

  it("does not show a pull request action when no artifact is registered", async () => {
    const user = userEvent.setup();
    renderCard();

    await user.click(screen.getByRole("button", {
      name: "Issue actions for Wire issue board to generated client",
    }));

    expect(screen.queryByRole("menuitem", { name: "Open pull request" })).not.toBeInTheDocument();
  });

  it("renders a ready quick action for backlog issues", async () => {
    const user = userEvent.setup();
    const { onStatusChange } = renderCard({
      issue: {
        ...issue,
        status: "backlog",
      },
    });

    const quickAction = screen.getByRole("button", { name: "Ready" });

    expect(quickAction).toHaveTextContent("Ready");
    expect(quickAction.querySelector("svg")).toBeInTheDocument();
    await user.click(quickAction);

    expect(onStatusChange).toHaveBeenCalledWith(24, "ready");
  });

  it("renders draft actions for blocked issues", async () => {
    const user = userEvent.setup();
    const { onStatusChange } = renderCard({
      issue: {
        ...issue,
        status: "blocked",
      },
    });

    const quickAction = screen.getByRole("button", { name: "Ready" });

    await user.click(screen.getByRole("button", {
      name: "Issue actions for Wire issue board to generated client",
    }));
    const menu = screen.getByRole("menu", {
      name: "Issue actions for Wire issue board to generated client",
    });
    const items = within(menu).getAllByRole("menuitem").map((item) => item.textContent);

    expect(items).toEqual(["Copy thread ID", "Blocked (current)", "Cancelled", "Done", "Duplicate"]);

    await user.click(quickAction);

    expect(onStatusChange).toHaveBeenCalledWith(24, "ready");
  });

  it("renders a done quick action for review issues", async () => {
    const user = userEvent.setup();
    const { onStatusChange } = renderCard({
      issue: {
        ...issue,
        status: "review",
      },
    });

    const quickAction = screen.getByRole("button", { name: "Done" });

    expect(quickAction).toHaveTextContent("Done");
    expect(quickAction.querySelector("svg")).not.toBeInTheDocument();
    await user.click(quickAction);

    expect(onStatusChange).toHaveBeenCalledWith(24, "done");
  });

  it("renders a reject action for review issues", async () => {
    const user = userEvent.setup();
    const { onRejectIssue } = renderCard({
      issue: {
        ...issue,
        status: "review",
      },
    });

    await user.click(screen.getByRole("button", { name: "Reject" }));

    expect(onRejectIssue).toHaveBeenCalledWith(24);
  });

  it("hides the reject action for non-review issues", () => {
    renderCard();

    expect(screen.queryByRole("button", { name: "Reject" })).not.toBeInTheDocument();
  });

  it("hides quick status actions for readonly issues", () => {
    renderCard({
      issue: {
        ...issue,
        status: "backlog",
      },
      readonly: true,
    });

    expect(screen.queryByRole("button", { name: "Ready" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Reject" })).not.toBeInTheDocument();
  });

  it("closes the action menu when clicking outside the card", async () => {
    const user = userEvent.setup();
    render(
      <QueryClientProvider client={queryClient}>
        <MemoryRouter>
          <IssueCard
            issue={issue}
            onStatusChange={async () => undefined}
          />
          <button type="button">Outside target</button>
        </MemoryRouter>
      </QueryClientProvider>,
    );

    await user.click(screen.getByRole("button", {
      name: "Issue actions for Wire issue board to generated client",
    }));
    expect(screen.getByRole("menu")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Outside target" }));

    expect(screen.queryByRole("menu")).not.toBeInTheDocument();
  });

  it("keeps the action menu open when clicking inside the card", async () => {
    const user = userEvent.setup();
    renderCard();

    await user.click(screen.getByRole("button", {
      name: "Issue actions for Wire issue board to generated client",
    }));
    await user.click(screen.getByRole("link", {
      name: "#24 Wire issue board to generated client",
    }));

    expect(screen.getByRole("menu")).toBeInTheDocument();
  });

  it("locks in-progress issues from web UI status changes", async () => {
    const user = userEvent.setup();
    renderCard({
      issue: {
        ...issue,
        status: "in_progress",
      },
    });

    await user.click(screen.getByRole("button", {
      name: "Issue actions for Wire issue board to generated client",
    }));
    const menu = screen.getByRole("menu", {
      name: "Issue actions for Wire issue board to generated client",
    });

    expect(within(menu).getByRole("menuitem", { name: "In Progress (current)" })).toBeDisabled();
    expect(within(menu).getByText("In Progress cannot be changed from the Web UI")).toBeInTheDocument();
  });

  it("loads and caches a thread ID when the action menu first opens", async () => {
    const user = userEvent.setup();
    vi.mocked(api.fetchOrchestratorIssueRuntime).mockResolvedValue(runtimeWithRuns([
      { thread_id: "  thread-latest  " },
      { thread_id: "thread-previous" },
    ]));
    renderCard();

    const menuButton = screen.getByRole("button", {
      name: "Issue actions for Wire issue board to generated client",
    });
    await user.click(menuButton);

    await waitFor(() => {
      expect(screen.getByRole("menuitem", { name: "Copy thread ID" })).toBeEnabled();
    });
    expect(api.fetchOrchestratorIssueRuntime).toHaveBeenCalledWith(24, { silent: true });

    await user.click(menuButton);
    await user.click(menuButton);

    expect(api.fetchOrchestratorIssueRuntime).toHaveBeenCalledTimes(1);
  });

  it("reuses the issue thread ID cache after the card remounts", async () => {
    const user = userEvent.setup();
    vi.mocked(api.fetchOrchestratorIssueRuntime).mockResolvedValue(runtimeWithRuns([
      { thread_id: "thread-latest" },
    ]));
    const firstRender = renderCard();

    await user.click(screen.getByRole("button", {
      name: "Issue actions for Wire issue board to generated client",
    }));
    await waitFor(() => expect(screen.getByRole("menuitem", { name: "Copy thread ID" })).toBeEnabled());
    firstRender.unmount();

    renderCard();
    await user.click(screen.getByRole("button", {
      name: "Issue actions for Wire issue board to generated client",
    }));

    await waitFor(() => expect(screen.getByRole("menuitem", { name: "Copy thread ID" })).toBeEnabled());
    expect(api.fetchOrchestratorIssueRuntime).toHaveBeenCalledTimes(1);
  });

  it("disables the copy action while the runtime request is pending", async () => {
    const user = userEvent.setup();
    let resolveRuntime: (runtime: OrchestratorIssueRuntime) => void;
    const runtime = new Promise<OrchestratorIssueRuntime>((resolve) => {
      resolveRuntime = resolve;
    });
    vi.mocked(api.fetchOrchestratorIssueRuntime).mockReturnValue(runtime);
    renderCard();

    await user.click(screen.getByRole("button", {
      name: "Issue actions for Wire issue board to generated client",
    }));

    expect(screen.getByRole("menuitem", { name: "Copy thread ID" })).toBeDisabled();
    resolveRuntime!(runtimeWithRuns([{ thread_id: "thread-latest" }]));
    await waitFor(() => expect(screen.getByRole("menuitem", { name: "Copy thread ID" })).toBeEnabled());
  });

  it("falls back to an older run with a non-empty trimmed thread ID", async () => {
    const user = userEvent.setup();
    const writeText = vi.fn().mockResolvedValue(undefined);
    vi.stubGlobal("navigator", { ...navigator, clipboard: { writeText } });
    vi.mocked(api.fetchOrchestratorIssueRuntime).mockResolvedValue(runtimeWithRuns([
      { thread_id: "  " },
      { thread_id: "  thread-previous  " },
    ]));
    renderCard({ readonly: true });

    await user.click(screen.getByRole("button", {
      name: "Issue actions for Wire issue board to generated client",
    }));
    const copyItem = screen.getByRole("menuitem", { name: "Copy thread ID" });
    await waitFor(() => expect(copyItem).toBeEnabled());
    await user.click(copyItem);

    await waitFor(() => expect(writeText).toHaveBeenCalledWith("thread-previous"));
    expect(screen.queryByRole("menu")).not.toBeInTheDocument();
    await waitFor(() => {
      expect(toastStore.getSnapshot()).toMatchObject([
        { type: "success", message: "Thread ID copied" },
      ]);
    });
  });

  it("keeps the copy action disabled when loading fails without showing a toast", async () => {
    const user = userEvent.setup();
    vi.mocked(api.fetchOrchestratorIssueRuntime).mockRejectedValue(new Error("unavailable"));
    renderCard();

    await user.click(screen.getByRole("button", {
      name: "Issue actions for Wire issue board to generated client",
    }));

    const copyItem = screen.getByRole("menuitem", { name: "Copy thread ID" });
    await waitFor(() => expect(api.fetchOrchestratorIssueRuntime).toHaveBeenCalledTimes(1));
    await user.click(screen.getByRole("button", {
      name: "Issue actions for Wire issue board to generated client",
    }));
    await user.click(screen.getByRole("button", {
      name: "Issue actions for Wire issue board to generated client",
    }));

    expect(copyItem).toBeDisabled();
    expect(api.fetchOrchestratorIssueRuntime).toHaveBeenCalledTimes(1);
    expect(toastStore.getSnapshot()).toEqual([]);
  });

  it("keeps the copy action disabled when no run has a thread ID", async () => {
    const user = userEvent.setup();
    vi.mocked(api.fetchOrchestratorIssueRuntime).mockResolvedValue(runtimeWithRuns([
      { thread_id: "  " },
      {},
    ]));
    renderCard();

    await user.click(screen.getByRole("button", {
      name: "Issue actions for Wire issue board to generated client",
    }));

    await waitFor(() => expect(api.fetchOrchestratorIssueRuntime).toHaveBeenCalledTimes(1));
    expect(screen.getByRole("menuitem", { name: "Copy thread ID" })).toBeDisabled();
  });

  it("shows the existing clipboard error toast when copying fails", async () => {
    const user = userEvent.setup();
    const writeText = vi.fn().mockRejectedValue(new Error("clipboard unavailable"));
    vi.stubGlobal("navigator", { ...navigator, clipboard: { writeText } });
    vi.mocked(api.fetchOrchestratorIssueRuntime).mockResolvedValue(runtimeWithRuns([
      { thread_id: "thread-latest" },
    ]));
    renderCard();

    await user.click(screen.getByRole("button", {
      name: "Issue actions for Wire issue board to generated client",
    }));
    const copyItem = screen.getByRole("menuitem", { name: "Copy thread ID" });
    await waitFor(() => expect(copyItem).toBeEnabled());
    await user.click(copyItem);

    await waitFor(() => {
      expect(toastStore.getSnapshot()).toMatchObject([
        { type: "error", message: "Could not copy to the clipboard" },
      ]);
    });
    expect(screen.queryByRole("menu")).not.toBeInTheDocument();
  });
});

function runtimeWithRuns(runs: Array<{ thread_id?: string }>): OrchestratorIssueRuntime {
  return { runs } as OrchestratorIssueRuntime;
}
