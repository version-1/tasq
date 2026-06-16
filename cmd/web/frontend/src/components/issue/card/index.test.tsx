import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it, vi } from "vitest";
import type { IssueSummary } from "@/lib/types";
import "@/lib/i18n";
import { IssueCard } from "./index";

const issue: IssueSummary = {
  id: 24,
  projectId: 1,
  projectKey: "tasq",
  title: "Wire issue board to generated client",
  description: "Issue body should not render in the card.",
  status: "ready",
  priority: "high",
  assignee: "web",
  createdAt: "2026-06-01T00:00:00.000Z",
  updatedAt: "2026-06-01T00:00:00.000Z",
};

function renderCard(props: Partial<Parameters<typeof IssueCard>[0]> = {}) {
  const onStatusChange = vi.fn(async () => undefined);
  render(
    <MemoryRouter>
      <IssueCard
        issue={issue}
        onStatusChange={onStatusChange}
        {...props}
      />
    </MemoryRouter>,
  );
  return { onStatusChange };
}

describe("IssueCard", () => {
  it("renders summary fields without the issue body", () => {
    renderCard({ commentCount: 3, runCount: 2 });

    expect(screen.getByRole("link", { name: "#24 Wire issue board to generated client" })).toBeInTheDocument();
    expect(screen.queryByText("Issue body should not render in the card.")).not.toBeInTheDocument();
    expect(screen.getByText("tasq")).toBeInTheDocument();
    expect(screen.getByText("high")).toBeInTheDocument();
    expect(screen.getByLabelText("3 comments")).toHaveTextContent("3");
    expect(screen.getByLabelText("2 runs")).toHaveTextContent("2");
  });

  it("renders zero metrics when counts are not provided", () => {
    renderCard();

    expect(screen.getByLabelText("0 comments")).toHaveTextContent("0");
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

    expect(items).toEqual(["Ready (current)", "Backlog", "Cancelled", "Done", "Duplicate"]);
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

    expect(items).toEqual(["Blocked (current)", "Cancelled", "Done", "Duplicate"]);

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

  it("hides quick status actions for readonly issues", () => {
    renderCard({
      issue: {
        ...issue,
        status: "backlog",
      },
      readonly: true,
    });

    expect(screen.queryByRole("button", { name: "Ready" })).not.toBeInTheDocument();
  });

  it("closes the action menu when clicking outside the card", async () => {
    const user = userEvent.setup();
    render(
      <MemoryRouter>
        <IssueCard
          issue={issue}
          onStatusChange={async () => undefined}
        />
        <button type="button">Outside target</button>
      </MemoryRouter>,
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
});
