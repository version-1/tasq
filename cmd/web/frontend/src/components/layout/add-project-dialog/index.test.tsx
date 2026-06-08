import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import type { CreateProjectInput } from "@/lib/types";
import { AddProjectDialog } from "./index";

function renderAddProjectDialog({
  error = "",
  onCancel = vi.fn(),
  onSubmit = vi.fn<(_: CreateProjectInput) => Promise<void>>().mockResolvedValue(undefined),
}: {
  error?: string;
  onCancel?: () => void;
  onSubmit?: (input: CreateProjectInput) => Promise<void>;
} = {}) {
  render(<AddProjectDialog error={error} onCancel={onCancel} onSubmit={onSubmit} />);

  return { onCancel, onSubmit };
}

describe("AddProjectDialog", () => {
  it("submits trimmed project values", async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn<(_: CreateProjectInput) => Promise<void>>().mockResolvedValue(undefined);
    renderAddProjectDialog({ onSubmit });
    const file = new File([""], "README.md", { type: "text/markdown" });
    Object.defineProperty(file, "webkitRelativePath", {
      value: "product-website/README.md",
    });

    await user.upload(screen.getByLabelText("Choose directory"), file);
    await user.clear(screen.getByLabelText("Location"));
    await user.type(screen.getByLabelText("Location"), "/workspace/product-website");
    await user.clear(screen.getByLabelText("Name"));
    await user.type(screen.getByLabelText("Name"), "  Product Website  ");
    await user.type(screen.getByLabelText("Description"), "  Customer-facing site  ");
    await user.click(screen.getByRole("button", { name: "Add" }));

    await waitFor(() => expect(onSubmit).toHaveBeenCalledTimes(1));
    expect(onSubmit).toHaveBeenCalledWith({
      description: "Customer-facing site",
      key: "product-website",
      location: "/workspace/product-website",
      name: "Product Website",
    });
  });

  it("fills key and name from the selected directory", async () => {
    const user = userEvent.setup();
    renderAddProjectDialog();
    const file = new File([""], "index.ts", { type: "text/typescript" });
    Object.defineProperty(file, "webkitRelativePath", {
      value: "agent_docs/index.ts",
    });

    await user.upload(screen.getByLabelText("Choose directory"), file);

    expect(screen.getByLabelText("Key")).toHaveValue("agent-docs");
    expect(screen.getByLabelText("Location")).toHaveValue("agent_docs");
    expect(screen.getByLabelText("Name")).toHaveValue("Agent Docs");
    expect(screen.getByText("agent_docs")).toBeTruthy();
  });

  it("shows validation errors without submitting", async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn<(_: CreateProjectInput) => Promise<void>>().mockResolvedValue(undefined);
    renderAddProjectDialog({ onSubmit });

    await user.click(screen.getByRole("button", { name: "Add" }));

    expect(screen.getByText("Enter a project location")).toBeTruthy();
    expect(onSubmit).not.toHaveBeenCalled();
  });

  it("requires an absolute project location", async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn<(_: CreateProjectInput) => Promise<void>>().mockResolvedValue(undefined);
    renderAddProjectDialog({ onSubmit });
    const file = new File([""], "index.ts", { type: "text/typescript" });
    Object.defineProperty(file, "webkitRelativePath", {
      value: "agent-docs/index.ts",
    });

    await user.upload(screen.getByLabelText("Choose directory"), file);
    await user.click(screen.getByRole("button", { name: "Add" }));

    expect(screen.getByText("Enter an absolute project location")).toBeTruthy();
    expect(onSubmit).not.toHaveBeenCalled();
  });

  it("calls onCancel from the close button", async () => {
    const user = userEvent.setup();
    const onCancel = vi.fn();
    renderAddProjectDialog({ onCancel });

    await user.click(screen.getByRole("button", { name: "Close" }));

    expect(onCancel).toHaveBeenCalledTimes(1);
  });
});
