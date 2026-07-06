import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { TabsProvider } from "@/context/tabs";
import type { Issue, IssueListResponse } from "@/lib/types";
import { Header } from "./index";

const fetchIssuesMock = vi.hoisted(() => vi.fn());

vi.mock("@/lib/api", () => ({
  fetchIssues: fetchIssuesMock,
}));

const issue: Issue = {
  id: 42,
  projectId: 1,
  projectKey: "TASQ",
  title: "Add issue search dropdown",
  description: "",
  status: "ready",
  priority: "normal",
  assignee: "",
  dependency_ids: [],
  createdAt: "2026-06-01T00:00:00.000Z",
  updatedAt: "2026-06-02T00:00:00.000Z",
};

function response(data: Issue[]): IssueListResponse {
  return {
    data,
    meta: {
      limit: 6,
      offset: 0,
      total: data.length,
      nextOffset: null,
    },
  };
}

function renderHeader() {
  return render(
    <MemoryRouter>
      <TabsProvider activeKey="issues">
        <Header
          activePage="issues"
          canDeleteProject={true}
          issueCount={1}
          language="en"
          projectName="Tasq"
          onAddTask={() => undefined}
          onLanguageChange={() => undefined}
        />
      </TabsProvider>
    </MemoryRouter>,
  );
}

describe("Header", () => {
  beforeEach(() => {
    fetchIssuesMock.mockReset();
  });

  it("shows search result links to issue details", async () => {
    const user = userEvent.setup();
    fetchIssuesMock.mockResolvedValueOnce(response([issue]));

    renderHeader();
    await user.type(screen.getByRole("searchbox"), "search");

    const title = await screen.findByText("Add issue search dropdown");
    const link = title.closest("a");
    expect(link).toHaveAttribute("href", "/issues/42");
    expect(fetchIssuesMock).toHaveBeenCalledWith(
      expect.objectContaining({
        limit: 6,
        search: "search",
        sort_by: "updated_at",
        sort_direction: "desc",
      }),
      { silent: true },
    );
  });

  it("shows an empty result message", async () => {
    const user = userEvent.setup();
    fetchIssuesMock.mockResolvedValueOnce(response([]));

    renderHeader();
    await user.type(screen.getByRole("searchbox"), "missing");

    await waitFor(() => {
      expect(screen.getByText("No matching issues")).toBeInTheDocument();
    });
  });

  it("opens project actions and selects project delete", async () => {
    const user = userEvent.setup();
    const onDeleteProject = vi.fn();
    render(
      <MemoryRouter>
        <TabsProvider activeKey="issues">
          <Header
            activePage="issues"
            canDeleteProject={true}
            issueCount={1}
            language="en"
            projectName="Tasq"
            onAddTask={() => undefined}
            onDeleteProject={onDeleteProject}
            onLanguageChange={() => undefined}
          />
        </TabsProvider>
      </MemoryRouter>,
    );

    await user.click(screen.getByRole("button", { name: "More project actions" }));
    await user.click(screen.getByRole("menuitem", { name: "Delete project" }));

    expect(onDeleteProject).toHaveBeenCalledTimes(1);
  });
});
