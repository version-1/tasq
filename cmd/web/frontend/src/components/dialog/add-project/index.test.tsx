import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";
import { AddProjectDialog } from "./index";

const projectAddCommand = "tq project add --key tasq .";

function renderAddProjectDialog({ onCancel = vi.fn() }: { onCancel?: () => void } = {}) {
  render(<AddProjectDialog onCancel={onCancel} />);

  return { onCancel };
}

describe("AddProjectDialog", () => {
  afterEach(() => {
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
  });

  it("shows a copyable tq project add command", () => {
    renderAddProjectDialog();

    expect(screen.getByRole("heading", { name: "Add project" })).toBeInTheDocument();
    expect(screen.getByText(projectAddCommand)).toBeInTheDocument();
  });

  it("copies the command to the clipboard", async () => {
    const user = userEvent.setup();
    const writeText = vi.fn().mockResolvedValue(undefined);
    vi.stubGlobal("navigator", { ...navigator, clipboard: { writeText } });
    renderAddProjectDialog();

    await user.click(screen.getByRole("button", { name: "Copy command" }));

    await waitFor(() => {
      expect(writeText).toHaveBeenCalledWith(projectAddCommand);
    });
    expect(screen.getByRole("button", { name: "Copied" })).toBeInTheDocument();
  });

  it("calls onCancel from the close button", async () => {
    const user = userEvent.setup();
    const onCancel = vi.fn();
    renderAddProjectDialog({ onCancel });

    await user.click(screen.getAllByRole("button", { name: "Close" })[0]);

    expect(onCancel).toHaveBeenCalledTimes(1);
  });
});
