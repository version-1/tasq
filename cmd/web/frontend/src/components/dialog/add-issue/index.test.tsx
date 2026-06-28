import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import type { CreateIssueInput, IssueSummary, Project } from "@/lib/types";
import { AddIssueDialog } from "./index";

const project: Project = {
  createdAt: "2026-06-08T00:00:00Z",
  description: "Product work",
  id: 7,
  key: "product",
  location: "/workspace/product",
  name: "Product Website",
  workflowChecksum: "",
  updatedAt: "2026-06-08T00:00:00Z",
};

const docsProject: Project = {
  createdAt: "2026-06-08T00:00:00Z",
  description: "Docs work",
  id: 8,
  key: "docs",
  location: "/workspace/docs",
  name: "Docs",
  workflowChecksum: "",
  updatedAt: "2026-06-08T00:00:00Z",
};

const projects = [project, docsProject];

const dependencyOptions: IssueSummary[] = [
  {
    assignee: "web",
    createdAt: "2026-06-08T00:00:00Z",
    dependency_ids: [],
    description: "Set up create issue tests",
    id: 2,
    priority: "high",
    projectId: 7,
    projectKey: "product",
    queueStatus: "queued",
    stats: { commentCount: 0 },
    status: "ready",
    title: "Add issue form test coverage",
    updatedAt: "2026-06-08T00:00:00Z",
  },
  {
    assignee: "design",
    createdAt: "2026-06-08T00:00:00Z",
    dependency_ids: [],
    description: "Review modal copy",
    id: 1,
    priority: "normal",
    projectId: 7,
    projectKey: "product",
    queueStatus: "backlog",
    stats: { commentCount: 1 },
    status: "backlog",
    title: "Review add issue dialog copy",
    updatedAt: "2026-06-08T00:00:00Z",
  },
  {
    assignee: "docs",
    createdAt: "2026-06-08T00:00:00Z",
    dependency_ids: [],
    description: "Draft docs",
    id: 3,
    priority: "normal",
    projectId: 8,
    projectKey: "docs",
    queueStatus: "queued",
    stats: { commentCount: 0 },
    status: "ready",
    title: "Draft docs dependency",
    updatedAt: "2026-06-08T00:00:00Z",
  },
  {
    assignee: "docs",
    createdAt: "2026-06-08T00:00:00Z",
    dependency_ids: [],
    description: "Done docs",
    id: 4,
    priority: "low",
    projectId: 8,
    projectKey: "docs",
    queueStatus: "completed",
    stats: { commentCount: 0 },
    status: "done",
    title: "Done docs dependency",
    updatedAt: "2026-06-08T00:00:00Z",
  },
];

function renderAddIssueDialog({
  dependencyOptionsOverride = dependencyOptions,
  error = "",
  initialStatus = "backlog",
  onCancel = vi.fn(),
  onSubmit = vi.fn<(_: CreateIssueInput) => Promise<void>>().mockResolvedValue(undefined),
  projectOverride = project,
}: {
  dependencyOptionsOverride?: IssueSummary[];
  error?: string;
  initialStatus?: CreateIssueInput["status"];
  onCancel?: () => void;
  onSubmit?: (input: CreateIssueInput) => Promise<void>;
  projectOverride?: Project | null;
} = {}) {
  render(
    <AddIssueDialog
      dependencyOptions={dependencyOptionsOverride}
      error={error}
      initialStatus={initialStatus}
      project={projectOverride}
      projects={projects}
      onCancel={onCancel}
      onSubmit={onSubmit}
    />,
  );

  return { onCancel, onSubmit };
}

describe("AddIssueDialog", () => {
  it("submits trimmed values with the selected project and initial status", async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn<(_: CreateIssueInput) => Promise<void>>().mockResolvedValue(undefined);
    renderAddIssueDialog({ initialStatus: "ready", onSubmit });

    await user.type(screen.getByLabelText("Title"), "  Ship modal tests  ");
    await user.type(screen.getByLabelText("Description"), "  Covers the modal flow  ");
    await user.type(screen.getByLabelText("Assignee"), "  Yuki  ");
    await user.selectOptions(screen.getByLabelText("Priority"), "high");
    expect(screen.getByLabelText("Project")).toHaveValue("7");
    await user.click(screen.getByRole("button", { name: "Add" }));

    await waitFor(() => expect(onSubmit).toHaveBeenCalledTimes(1));
    expect(onSubmit).toHaveBeenCalledWith({
      assignee: "Yuki",
      description: "  Covers the modal flow  ",
      priority: "high",
      projectId: 7,
      status: "ready",
      title: "Ship modal tests",
    });
    expect(onSubmit.mock.calls[0][0]).not.toHaveProperty("dependency_ids");
  });

  it("omits blank markdown descriptions", async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn<(_: CreateIssueInput) => Promise<void>>().mockResolvedValue(undefined);
    renderAddIssueDialog({ onSubmit });

    await user.type(screen.getByLabelText("Title"), "Create issue without docs");
    fireEvent.change(screen.getByLabelText("Description"), { target: { value: "  \n  " } });
    await user.click(screen.getByRole("button", { name: "Add" }));

    await waitFor(() => expect(onSubmit).toHaveBeenCalledTimes(1));
    expect(onSubmit.mock.calls[0][0]).not.toHaveProperty("description");
  });

  it("previews markdown descriptions", async () => {
    const user = userEvent.setup();
    renderAddIssueDialog();

    await user.type(screen.getByLabelText("Description"), "## Preview title");
    await user.click(screen.getByRole("tab", { name: "Preview" }));

    expect(screen.getByRole("heading", { name: "Preview title" })).toBeInTheDocument();
  });

  it("submits with the selected project and filters dependency options", async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn<(_: CreateIssueInput) => Promise<void>>().mockResolvedValue(undefined);
    renderAddIssueDialog({ onSubmit });

    await user.selectOptions(screen.getByLabelText("Project"), "8");
    await user.type(screen.getByLabelText("Title"), "Create docs issue");
    await user.click(screen.getByRole("button", { name: "Select dependencies" }));

    expect(screen.queryByLabelText(/#2 Add issue form test coverage/)).not.toBeInTheDocument();
    expect(screen.getByLabelText(/#3 Draft docs dependency/)).toBeInTheDocument();
    expect(screen.queryByLabelText(/#4 Done docs dependency/)).not.toBeInTheDocument();

    await user.click(screen.getByLabelText(/#3 Draft docs dependency/));
    await user.click(screen.getByRole("button", { name: "Add" }));

    await waitFor(() => expect(onSubmit).toHaveBeenCalledTimes(1));
    expect(onSubmit).toHaveBeenCalledWith(
      expect.objectContaining({
        dependency_ids: [3],
        projectId: 8,
        title: "Create docs issue",
      }),
    );
  });

  it("submits selected dependency IDs", async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn<(_: CreateIssueInput) => Promise<void>>().mockResolvedValue(undefined);
    renderAddIssueDialog({ onSubmit });

    await user.type(screen.getByLabelText("Title"), "Create blocked issue");
    await user.click(screen.getByRole("button", { name: "Select dependencies" }));
    await user.click(screen.getByLabelText(/#2 Add issue form test coverage/));
    await user.click(screen.getByLabelText(/#1 Review add issue dialog copy/));
    await user.click(screen.getByRole("button", { name: "Add" }));

    await waitFor(() => expect(onSubmit).toHaveBeenCalledTimes(1));
    expect(onSubmit).toHaveBeenCalledWith(
      expect.objectContaining({
        dependency_ids: [1, 2],
        title: "Create blocked issue",
      }),
    );
  });

  it("shows validation errors without submitting", async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn<(_: CreateIssueInput) => Promise<void>>().mockResolvedValue(undefined);
    renderAddIssueDialog({ onSubmit });

    await user.click(screen.getByRole("button", { name: "Add" }));

    expect(screen.getByText("Enter a title")).toBeTruthy();
    expect(onSubmit).not.toHaveBeenCalled();
  });

  it("shows API errors passed from the layout", () => {
    renderAddIssueDialog({ error: "failed to create issue" });

    expect(screen.getByText("failed to create issue")).toBeTruthy();
  });

  it("calls onCancel from the close button", async () => {
    const user = userEvent.setup();
    const onCancel = vi.fn();
    renderAddIssueDialog({ onCancel });

    await user.click(screen.getByRole("button", { name: "Close" }));

    expect(onCancel).toHaveBeenCalledTimes(1);
  });
});
