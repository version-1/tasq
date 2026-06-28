import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { MarkdownEditor } from "./index";

const labels = {
  cancel: "Cancel",
  edit: "Edit description",
  empty: "No description",
  raw: "Raw",
  preview: "Preview",
  save: "Save",
  saving: "Saving...",
  textarea: "Description",
};

describe("MarkdownEditor", () => {
  it("renders markdown in read mode and opens the raw editor", async () => {
    const user = userEvent.setup();
    render(<MarkdownEditor labels={labels} title="Description" value="## Current body" />);

    expect(screen.getByRole("heading", { name: "Current body" })).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Edit description" }));

    expect(screen.getByRole("textbox", { name: "Description" })).toHaveValue("## Current body");
    expect(screen.getByRole("tab", { name: "Raw" })).toHaveAttribute("aria-selected", "true");
  });

  it("previews the draft and returns to read mode after saving", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn<(_: string) => Promise<void>>().mockResolvedValue(undefined);
    render(
      <MarkdownEditor
        labels={labels}
        title="Description"
        value="Initial body"
        onSave={onSave}
      />,
    );

    await user.click(screen.getByRole("button", { name: "Edit description" }));
    await user.clear(screen.getByRole("textbox", { name: "Description" }));
    await user.type(screen.getByRole("textbox", { name: "Description" }), "Updated **body**");
    await user.click(screen.getByRole("tab", { name: "Preview" }));

    expect(screen.getByText("body")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Save" }));

    await waitFor(() => expect(onSave).toHaveBeenCalledWith("Updated **body**"));
    expect(screen.queryByRole("textbox", { name: "Description" })).not.toBeInTheDocument();
  });

  it("keeps the draft visible when saving fails", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn<(_: string) => Promise<void>>().mockRejectedValue(new Error("Save failed"));
    render(
      <MarkdownEditor
        labels={labels}
        title="Description"
        value="Initial body"
        onSave={onSave}
      />,
    );

    await user.click(screen.getByRole("button", { name: "Edit description" }));
    await user.clear(screen.getByRole("textbox", { name: "Description" }));
    await user.type(screen.getByRole("textbox", { name: "Description" }), "Unsaved draft");
    await user.click(screen.getByRole("button", { name: "Save" }));

    expect(await screen.findByText("Save failed")).toBeInTheDocument();
    expect(screen.getByRole("textbox", { name: "Description" })).toHaveValue("Unsaved draft");
  });
});
