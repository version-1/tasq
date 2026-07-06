import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import type { Project } from "@/lib/types";
import { DeleteProjectDialog } from "./index";

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

function renderDeleteProjectDialog({
  error = "",
  isDeleting = false,
  onCancel = vi.fn(),
  onConfirm = vi.fn<() => Promise<void>>().mockResolvedValue(undefined),
}: {
  error?: string;
  isDeleting?: boolean;
  onCancel?: () => void;
  onConfirm?: () => Promise<void>;
} = {}) {
  render(
    <DeleteProjectDialog
      error={error}
      isDeleting={isDeleting}
      project={project}
      onCancel={onCancel}
      onConfirm={onConfirm}
    />,
  );

  return { onCancel, onConfirm };
}

describe("DeleteProjectDialog", () => {
  it("keeps delete disabled until the project key matches", async () => {
    const user = userEvent.setup();
    const { onConfirm } = renderDeleteProjectDialog();
    const deleteButton = screen.getByRole("button", { name: "Delete project" });

    expect(deleteButton).toBeDisabled();
    await user.type(screen.getByLabelText("Type product to confirm"), "prod");
    expect(deleteButton).toBeDisabled();
    await user.click(deleteButton);
    expect(onConfirm).not.toHaveBeenCalled();
  });

  it("submits when the project key matches", async () => {
    const user = userEvent.setup();
    const onConfirm = vi.fn<() => Promise<void>>().mockResolvedValue(undefined);
    renderDeleteProjectDialog({ onConfirm });

    await user.type(screen.getByLabelText("Type product to confirm"), "product");
    await user.click(screen.getByRole("button", { name: "Delete project" }));

    await waitFor(() => expect(onConfirm).toHaveBeenCalledTimes(1));
  });

  it("shows a running run refusal reason", () => {
    renderDeleteProjectDialog({
      error: "This project still has running runs. Stop them before deleting the project.",
    });

    expect(screen.getByText(/still has running runs/i)).toBeInTheDocument();
  });
});
