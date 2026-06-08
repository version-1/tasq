import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { modalIDs } from "@/constants";
import { ModalProvider, useModal } from "@/lib/modal";

function ModalStateProbe() {
  const modal = useModal();

  return (
    <div>
      <p>active: {modal.activeModalID ?? "none"}</p>
      <p>add issue open: {String(modal.isModalOpen(modalIDs.addIssue))}</p>
      <button type="button" onClick={() => modal.openModal(modalIDs.addIssue)}>
        open
      </button>
      <button type="button" onClick={modal.closeModal}>
        close
      </button>
    </div>
  );
}

function renderModalStateProbe() {
  return render(
    <ModalProvider>
      <ModalStateProbe />
    </ModalProvider>,
  );
}

describe("ModalProvider", () => {
  it("opens and closes the active modal id", async () => {
    const user = userEvent.setup();
    renderModalStateProbe();

    expect(screen.getByText("active: none")).toBeTruthy();
    expect(screen.getByText("add issue open: false")).toBeTruthy();

    await user.click(screen.getByRole("button", { name: "open" }));

    expect(screen.getByText("active: addIssue")).toBeTruthy();
    expect(screen.getByText("add issue open: true")).toBeTruthy();

    await user.click(screen.getByRole("button", { name: "close" }));

    expect(screen.getByText("active: none")).toBeTruthy();
    expect(screen.getByText("add issue open: false")).toBeTruthy();
  });

  it("requires a provider for useModal", () => {
    const consoleError = vi.spyOn(console, "error").mockImplementation(() => undefined);

    expect(() => render(<ModalStateProbe />)).toThrow("useModal must be used within ModalProvider");

    consoleError.mockRestore();
  });
});
