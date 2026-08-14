import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import "@/lib/i18n";
import { RejectAction } from ".";

describe("RejectAction", () => {
  it("opens the dialog from the primary action", async () => {
    const user = userEvent.setup();
    const onOpenDialog = vi.fn();

    render(
      <RejectAction onOpenDialog={onOpenDialog} onSelectShortcut={async () => undefined} />,
    );

    await user.click(screen.getByRole("button", { name: "Reject" }));
    expect(onOpenDialog).toHaveBeenCalledOnce();
  });

  it("submits the selected shortcut from the menu", async () => {
    const user = userEvent.setup();
    const onSelectShortcut = vi.fn(async () => undefined);

    render(
      <RejectAction onOpenDialog={() => undefined} onSelectShortcut={onSelectShortcut} />,
    );

    await user.click(screen.getByRole("button", { name: "Reject shortcuts" }));
    const menu = screen.getByRole("menu", { name: "Reject shortcuts" });
    expect(within(menu).getByRole("menuitem", { name: "Write a comment…" })).toBeVisible();
    await user.click(within(menu).getByRole("menuitem", { name: "Fix CI & Conflict" }));

    expect(onSelectShortcut).toHaveBeenCalledWith({
      id: "fix-ci-conflict",
      label: "Fix CI & Conflict",
      body: "Fix CI & Conflict",
    });
  });
});
