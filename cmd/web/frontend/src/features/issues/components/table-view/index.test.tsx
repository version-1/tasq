import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { Issue, IssueListResponse, Project } from "@/lib/types";
import { IssuesTableView } from "./index";

const fetchIssuesMock = vi.hoisted(() => vi.fn());

vi.mock("@/lib/api", () => ({
  fetchIssues: fetchIssuesMock,
}));

const baseIssue: Issue = {
  id: 42,
  projectId: 1,
  projectKey: "TASQ",
  title: "Add issue table view",
  description: "",
  status: "ready",
  priority: "high",
  assignee: "frontend",
  dependency_ids: [],
  createdAt: "2026-06-01T00:00:00.000Z",
  updatedAt: "2026-06-02T00:00:00.000Z",
};

const projectOptions: Project[] = [
  {
    id: 1,
    key: "TASQ",
    name: "Tasq",
    description: "",
    location: ".",
    workflowChecksum: "",
    createdAt: "2026-06-01T00:00:00.000Z",
    updatedAt: "2026-06-01T00:00:00.000Z",
  },
  {
    id: 2,
    key: "WEB",
    name: "Web",
    description: "",
    location: ".",
    workflowChecksum: "",
    createdAt: "2026-06-01T00:00:00.000Z",
    updatedAt: "2026-06-01T00:00:00.000Z",
  },
];

function response(data: Issue[], total = data.length, nextOffset: number | null = null): IssueListResponse {
  return {
    data,
    meta: {
      limit: 50,
      offset: 0,
      total,
      nextOffset,
    },
  };
}

function renderTable({ projectID = 1 }: { projectID?: number | null } = {}) {
  return render(
    <MemoryRouter>
      <IssuesTableView
        projectID={projectID}
        projectOptions={projectOptions}
        refreshIntervalMs={60_000}
      />
    </MemoryRouter>,
  );
}

describe("IssuesTableView", () => {
  beforeEach(() => {
    fetchIssuesMock.mockReset();
  });

  it("renders issue rows from the API response", async () => {
    fetchIssuesMock.mockResolvedValueOnce(response([baseIssue]));

    renderTable();

    expect(await screen.findByRole("link", { name: "Add issue table view" })).toHaveAttribute(
      "href",
      "/issues/42",
    );
    const row = screen.getByRole("row", { name: /#42 Add issue table view/ });
    expect(within(row).getByText("#42")).toBeInTheDocument();
    expect(within(row).getByText("frontend")).toBeInTheDocument();
    expect(fetchIssuesMock).toHaveBeenCalledWith(
      expect.objectContaining({
        limit: 50,
        offset: 0,
        project_id: 1,
        sort_by: "updated_at",
        sort_direction: "desc",
        states: "backlog,ready,in_progress",
      }),
      { silent: true },
    );
  });

  it("reloads with applied status filters", async () => {
    const user = userEvent.setup();
    fetchIssuesMock
      .mockResolvedValueOnce(response([baseIssue]))
      .mockResolvedValueOnce(response([]));

    renderTable();
    await screen.findByRole("link", { name: "Add issue table view" });

    await user.click(screen.getByRole("button", { name: /Status:/ }));
    await user.click(screen.getByLabelText("backlog"));
    expect(fetchIssuesMock).toHaveBeenCalledTimes(1);
    await user.click(screen.getByRole("button", { name: "Apply" }));
    await waitFor(() => {
      expect(fetchIssuesMock).toHaveBeenLastCalledWith(
        expect.objectContaining({ states: "ready,in_progress", offset: 0 }),
        { silent: true },
      );
    });
  });

  it("reloads with applied project and priority filters", async () => {
    const user = userEvent.setup();
    fetchIssuesMock
      .mockResolvedValueOnce(response([baseIssue]))
      .mockResolvedValueOnce(response([]))
      .mockResolvedValueOnce(response([]));

    renderTable({ projectID: null });
    await screen.findByRole("link", { name: "Add issue table view" });

    await user.click(screen.getByRole("button", { name: /Project:/ }));
    await user.click(screen.getByLabelText("Web"));
    await user.click(screen.getByRole("button", { name: "Apply" }));
    await waitFor(() => {
      expect(fetchIssuesMock).toHaveBeenLastCalledWith(
        expect.objectContaining({ project_ids: "2", project_id: undefined, offset: 0 }),
        { silent: true },
      );
    });

    await user.click(screen.getByRole("button", { name: /Priority:/ }));
    await user.click(screen.getByLabelText("high"));
    await user.click(screen.getByRole("button", { name: "Apply" }));
    await waitFor(() => {
      expect(fetchIssuesMock).toHaveBeenLastCalledWith(
        expect.objectContaining({ priorities: "high", project_ids: "2", offset: 0 }),
        { silent: true },
      );
    });
  });

  it("reloads with server-side sort parameters", async () => {
    const user = userEvent.setup();
    fetchIssuesMock
      .mockResolvedValueOnce(response([baseIssue]))
      .mockResolvedValueOnce(response([]))
      .mockResolvedValueOnce(response([]));

    renderTable();
    await screen.findByRole("link", { name: "Add issue table view" });

    await user.click(screen.getByRole("button", { name: "Sort by ID" }));
    await waitFor(() => {
      expect(fetchIssuesMock).toHaveBeenLastCalledWith(
        expect.objectContaining({ sort_by: "id", sort_direction: "desc", offset: 0 }),
        { silent: true },
      );
    });

    await user.click(screen.getByRole("button", { name: "Sort by ID" }));
    await waitFor(() => {
      expect(fetchIssuesMock).toHaveBeenLastCalledWith(
        expect.objectContaining({ sort_by: "id", sort_direction: "asc", offset: 0 }),
        { silent: true },
      );
    });
  });

  it("reloads with priority sort parameters", async () => {
    const user = userEvent.setup();
    fetchIssuesMock
      .mockResolvedValueOnce(response([baseIssue]))
      .mockResolvedValueOnce(response([]));

    renderTable();
    await screen.findByRole("link", { name: "Add issue table view" });

    await user.click(screen.getByRole("button", { name: "Sort by Priority" }));
    await waitFor(() => {
      expect(fetchIssuesMock).toHaveBeenLastCalledWith(
        expect.objectContaining({ sort_by: "priority", sort_direction: "desc", offset: 0 }),
        { silent: true },
      );
    });
  });
});
