import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import type { CreateIssueInput, Project } from "@/lib/types";
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

function renderAddIssueDialog({
  error = "",
  initialStatus = "backlog",
  onCancel = vi.fn(),
  onSubmit = vi.fn<(_: CreateIssueInput) => Promise<void>>().mockResolvedValue(undefined),
  projectOverride = project,
}: {
  error?: string;
  initialStatus?: CreateIssueInput["status"];
  onCancel?: () => void;
  onSubmit?: (input: CreateIssueInput) => Promise<void>;
  projectOverride?: Project | null;
} = {}) {
  render(
    <AddIssueDialog
      error={error}
      initialStatus={initialStatus}
      project={projectOverride}
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
    await user.click(screen.getByRole("button", { name: "Add" }));

    await waitFor(() => expect(onSubmit).toHaveBeenCalledTimes(1));
    expect(onSubmit).toHaveBeenCalledWith({
      assignee: "Yuki",
      description: "Covers the modal flow",
      priority: "high",
      projectId: 7,
      status: "ready",
      title: "Ship modal tests",
    });
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
